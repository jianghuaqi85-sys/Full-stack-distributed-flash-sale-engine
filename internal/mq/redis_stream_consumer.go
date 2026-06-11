package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"order-system/internal/pkg/db"
	"order-system/internal/pkg/redis/pkgredis"
	"order-system/internal/pkg/ws"
	"order-system/internal/repository"
)

// RedisStreamConsumer Redis Stream 消费者（降级方案）
type RedisStreamConsumer struct {
	redis          *pkgredis.RedisClientWrapper
	db             *gorm.DB
	ticketRepo     repository.TicketRepository
	ticketTypeRepo repository.TicketTypeRepository
	wsHub          *ws.Hub
	name           string
}

// NewRedisStreamConsumer 创建 Redis Stream 消费者
func NewRedisStreamConsumer(redis *pkgredis.RedisClientWrapper, database *gorm.DB, wsHub *ws.Hub) *RedisStreamConsumer {
	return &RedisStreamConsumer{
		redis:          redis,
		db:             database,
		ticketRepo:     repository.NewTicketRepository(database),
		ticketTypeRepo: repository.NewTicketTypeRepository(database),
		wsHub:          wsHub,
		name:           fmt.Sprintf("ticket-worker-%d", time.Now().UnixNano()),
	}
}

// Start 启动消费循环
func (c *RedisStreamConsumer) Start(ctx context.Context) {
	c.redis.XGroupCreate(ctx, TicketStreamKey, TicketConsumerGroup)

	log.Printf("[RedisMQ] Ticket consumer %s started", c.name)

	consecutiveErrors := 0
	for {
		select {
		case <-ctx.Done():
			log.Printf("[RedisMQ] Ticket consumer %s stopped", c.name)
			return
		default:
			c.consume(ctx, &consecutiveErrors)
		}
	}
}

func (c *RedisStreamConsumer) consume(ctx context.Context, consecutiveErrors *int) {
	streams, err := c.redis.XReadGroup(ctx, TicketConsumerGroup, c.name,
		[]string{TicketStreamKey, ">"}, 10, 2*time.Second)

	if err != nil {
		*consecutiveErrors++
		log.Printf("[RedisMQ] XReadGroup error (consecutive=%d): %v", *consecutiveErrors, err)
		backoff := time.Duration(*consecutiveErrors) * time.Second
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}
		time.Sleep(backoff)
		return
	}

	*consecutiveErrors = 0

	for _, stream := range streams {
		for _, msg := range stream.Messages {
			c.processTicket(ctx, msg)
			c.redis.XAck(ctx, TicketStreamKey, TicketConsumerGroup, msg.ID)
		}
	}
}

func (c *RedisStreamConsumer) processTicket(ctx context.Context, msg goredis.XMessage) {
	data, ok := msg.Values["data"].(string)
	if !ok {
		log.Printf("[RedisMQ] Invalid ticket message format: %v", msg.Values)
		return
	}

	var ticketMsg TicketMessage
	if err := json.Unmarshal([]byte(data), &ticketMsg); err != nil {
		log.Printf("[RedisMQ] Failed to unmarshal ticket message: %v", err)
		return
	}

	log.Printf("[RedisMQ] Processing ticket for user %d, event %d", ticketMsg.UserID, ticketMsg.EventID)

	// 幂等检查（与 Kafka 消费者保持一致）
	idempotentKey := fmt.Sprintf("mq:processed:%d:%d:%d:%d",
		ticketMsg.UserID, ticketMsg.EventID, ticketMsg.TicketTypeID, ticketMsg.Timestamp)
	set, err := c.redis.SetNX(ctx, idempotentKey, "1", 24*time.Hour)
	if err != nil {
		log.Printf("[RedisMQ] Idempotent check failed: %v", err)
		return // Redis 故障，不确认消息，等待重试
	}
	if !set {
		log.Printf("[RedisMQ] Duplicate message detected, skipping: user=%d event=%d", ticketMsg.UserID, ticketMsg.EventID)
		return // 重复消息，跳过
	}

	orderNo := generateOrderNo()

	ticketType, err := c.ticketTypeRepo.FindByID(ticketMsg.TicketTypeID)
	if err != nil || ticketType == nil {
		log.Printf("[RedisMQ] Ticket type %d not found", ticketMsg.TicketTypeID)
		c.sendResult(ticketMsg.UserID, 0, "failed", "票种不存在", "", "")
		return
	}

	if err := c.ticketTypeRepo.AtomicDeductStock(ticketMsg.TicketTypeID, ticketMsg.Quantity); err != nil {
		c.redis.SeckillRollback(ctx, fmt.Sprint(ticketMsg.EventID),
			fmt.Sprint(ticketMsg.TicketTypeID), fmt.Sprint(ticketMsg.UserID))
		// 删除幂等 key，允许重试
		c.redis.Client().Del(ctx, idempotentKey)
		c.sendResult(ticketMsg.UserID, 0, "failed", "库存不足", "", "")
		return
	}

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
		log.Printf("[RedisMQ] Failed to create ticket: %v", err)
		c.ticketTypeRepo.UpdateStock(ticketMsg.TicketTypeID, ticketMsg.Quantity)
		c.redis.SeckillRollback(ctx, fmt.Sprint(ticketMsg.EventID),
			fmt.Sprint(ticketMsg.TicketTypeID), fmt.Sprint(ticketMsg.UserID))
		// 删除幂等 key，允许重试
		c.redis.Client().Del(ctx, idempotentKey)
		c.sendResult(ticketMsg.UserID, 0, "failed", "创建票务失败", "", "")
		return
	}

	c.sendResult(ticketMsg.UserID, ticket.ID, "success", "购票成功！", ticket.OrderNo, ticketType.Name)
	log.Printf("[RedisMQ] Ticket %d created successfully for user %d, order_no=%s", ticket.ID, ticketMsg.UserID, ticket.OrderNo)
}

func (c *RedisStreamConsumer) sendResult(userID uint, ticketID uint, status, message, orderNo, ticketType string) {
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
