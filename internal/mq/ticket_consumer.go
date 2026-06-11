package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"order-system/internal/pkg/cache"
	"order-system/internal/pkg/db"
	appkafka "order-system/internal/pkg/kafka"
	"order-system/internal/pkg/logger"
	"order-system/internal/pkg/redis/pkgredis"
	"order-system/internal/pkg/ws"
	"order-system/internal/repository"
)

// TicketConsumer 票务消息消费者（Kafka）
type TicketConsumer struct {
	consumer       *appkafka.Consumer
	ticketRepo     repository.TicketRepository
	ticketTypeRepo repository.TicketTypeRepository
	localCache     *cache.LocalCache
	wsHub          *ws.Hub
	redis          *pkgredis.RedisClientWrapper
}

// NewTicketConsumer 创建 Kafka 票务消费者
func NewTicketConsumer(
	brokers []string,
	database *gorm.DB,
	wsHub *ws.Hub,
	redis *pkgredis.RedisClientWrapper,
	localCache *cache.LocalCache,
) (*TicketConsumer, error) {
	c := &TicketConsumer{
		ticketRepo:     repository.NewTicketRepository(database),
		ticketTypeRepo: repository.NewTicketTypeRepository(database),
		localCache:     localCache,
		wsHub:          wsHub,
		redis:          redis,
	}

	consumer, err := appkafka.NewConsumer(appkafka.ConsumerConfig{
		Brokers:    brokers,
		Topic:      TicketOrderTopic,
		Group:      TicketOrderConsumerGroup,
		Handler:    c.handleMessage,
		MaxRetries: 3,
		DLQTopic:   TicketOrderDLQTopic,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket consumer: %w", err)
	}

	c.consumer = consumer
	return c, nil
}

// Start 启动消费者
func (c *TicketConsumer) Start(ctx context.Context) {
	c.consumer.Start(ctx)
}

// handleMessage 处理单条票务消息
func (c *TicketConsumer) handleMessage(ctx context.Context, key, value []byte) error {
	var ticketMsg TicketMessage
	if err := json.Unmarshal(value, &ticketMsg); err != nil {
		logger.Error("[Kafka] Failed to unmarshal ticket message", "error", err)
		return nil // 解析失败不重试，直接丢弃
	}

	logger.Info("[Kafka] Processing ticket", "user_id", ticketMsg.UserID, "event_id", ticketMsg.EventID, "ticket_type_id", ticketMsg.TicketTypeID)

	// 生成订单号（在幂等检查之前，确保同一消息总是生成相同的订单号）
	orderNo := generateOrderNo()

	// 幂等检查：使用订单号作为唯一键（基于消息内容的确定性生成）
	// 订单号 = TK + Sonyflake ID，Sonyflake ID 基于时间戳+机器ID+序列号，同一消息重复消费时
	// 由于消息内容相同，generateOrderNo() 会生成不同的订单号（因为时间不同），
	// 所以我们使用消息内容哈希作为幂等 key，确保同一消息的幂等性
	idempotentKey := fmt.Sprintf("mq:processed:%d:%d:%d:%d",
		ticketMsg.UserID, ticketMsg.EventID, ticketMsg.TicketTypeID, ticketMsg.Timestamp)
	set, err := c.redis.SetNX(ctx, idempotentKey, orderNo, 24*time.Hour)
	if err != nil {
		logger.Error("[Kafka] Idempotent check failed", "error", err)
		return err // Redis 故障，重试
	}
	if !set {
		// 检查是否是 DLQ 重试场景：如果 idempotent key 存在但值是旧的订单号格式
		// 说明是之前失败的消息重试，需要检查之前的订单是否已创建
		existingOrderNo, _ := c.redis.Client().Get(ctx, idempotentKey).Result()
		logger.Info("[Kafka] Duplicate message detected", "user_id", ticketMsg.UserID, "event_id", ticketMsg.EventID, "existing_order_no", existingOrderNo)
		return nil // 重复消息，跳过
	}

	// 查询票种信息
	ticketType, err := c.ticketTypeRepo.FindByID(ticketMsg.TicketTypeID)
	if err != nil || ticketType == nil {
		logger.Error("[Kafka] Ticket type not found", "ticket_type_id", ticketMsg.TicketTypeID)
		c.sendResult(ticketMsg.UserID, 0, "failed", "票种不存在", "", "")
		return nil // 数据问题不重试
	}

	// 原子扣减数据库库存
	if err := c.ticketTypeRepo.AtomicDeductStock(ticketMsg.TicketTypeID, ticketMsg.Quantity); err != nil {
		// 回滚 Redis 库存
		activityID := fmt.Sprintf("ticket:%d", ticketMsg.EventID)
		c.redis.SeckillRollback(ctx, activityID,
			fmt.Sprint(ticketMsg.TicketTypeID), fmt.Sprint(ticketMsg.UserID))
		// 删除幂等 key，允许重试
		c.redis.Client().Del(ctx, idempotentKey)
		c.sendResult(ticketMsg.UserID, 0, "failed", "库存不足", "", "")
		return nil // 库存不足不重试
	}

	// 创建票务记录（使用预先生成的订单号）
	ticket := &db.Ticket{
		UserID:       ticketMsg.UserID,
		EventID:      ticketMsg.EventID,
		TicketTypeID: ticketMsg.TicketTypeID,
		Quantity:     ticketMsg.Quantity,
		TotalPrice:   float64(ticketMsg.Quantity) * ticketType.Price,
		Status:       "reserved",
		OrderNo:      orderNo,
	}

	if err := c.ticketRepo.Create(ticket); err != nil {
		// 回滚数据库库存
		c.ticketTypeRepo.UpdateStock(ticketMsg.TicketTypeID, ticketMsg.Quantity)
		// 回滚 Redis 库存
		activityID := fmt.Sprintf("ticket:%d", ticketMsg.EventID)
		c.redis.SeckillRollback(ctx, activityID,
			fmt.Sprint(ticketMsg.TicketTypeID), fmt.Sprint(ticketMsg.UserID))
		// 删除幂等 key，允许重试
		c.redis.Client().Del(ctx, idempotentKey)
		c.sendResult(ticketMsg.UserID, 0, "failed", "创建票务失败", "", "")
		return err // DB 故障，重试
	}

	// 更新本地缓存中的库存快照
	if c.localCache != nil {
		cacheKey := cache.StockCacheKey(ticketMsg.EventID, ticketMsg.TicketTypeID)
		c.localCache.Delete(cacheKey)
	}

	c.sendResult(ticketMsg.UserID, ticket.ID, "success", "购票成功！", ticket.OrderNo, ticketType.Name)
	logger.Info("[Kafka] Ticket created successfully", "ticket_id", ticket.ID, "user_id", ticketMsg.UserID, "order_no", ticket.OrderNo)
	return nil
}

// sendResult 通过 WebSocket 推送结果给用户
func (c *TicketConsumer) sendResult(userID uint, ticketID uint, status, message, orderNo, ticketType string) {
	result := ws.WSMessage{
		Type: "ticket_result",
		Payload: ws.TicketResultPayload{
			TicketID:   ticketID,
			UserID:     userID,
			Status:     status,
			Message:    message,
			OrderNo:    orderNo,
			TicketType: ticketType,
			Timestamp:  time.Now().Unix(),
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		return
	}

	c.wsHub.SendToUser(fmt.Sprint(userID), data)
}

// generateOrderNo 生成订单号（使用 Sonyflake，降级为 rand）
func generateOrderNo() string {
	// 优先使用 Sonyflake
	if idgen := getIDGen(); idgen != nil {
		return idgen.OrderNo()
	}
	// 降级方案
	return generateFallbackOrderNo()
}
