package mq

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"order-system/internal/pkg/db"
	"order-system/internal/pkg/redis/pkgredis"
	"order-system/internal/pkg/ws"
	"order-system/internal/repository"
)

type TicketConsumer struct {
	redis          *pkgredis.RedisClientWrapper
	db             *gorm.DB
	ticketRepo     repository.TicketRepository
	ticketTypeRepo repository.TicketTypeRepository
	wsHub          *ws.Hub
	name           string
}

func NewTicketConsumer(redis *pkgredis.RedisClientWrapper, database *gorm.DB, wsHub *ws.Hub) *TicketConsumer {
	return &TicketConsumer{
		redis:          redis,
		db:             database,
		ticketRepo:     repository.NewTicketRepository(database),
		ticketTypeRepo: repository.NewTicketTypeRepository(database),
		wsHub:          wsHub,
		name:           fmt.Sprintf("ticket-worker-%d", time.Now().UnixNano()),
	}
}

func (c *TicketConsumer) Start(ctx context.Context) {
	c.redis.XGroupCreate(ctx, TicketStreamKey, TicketConsumerGroup)

	log.Printf("[MQ] Ticket consumer %s started", c.name)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[MQ] Ticket consumer %s stopped", c.name)
			return
		default:
			c.consume(ctx)
		}
	}
}

func (c *TicketConsumer) consume(ctx context.Context) {
	streams, err := c.redis.XReadGroup(ctx, TicketConsumerGroup, c.name,
		[]string{TicketStreamKey, ">"}, 10, 2*time.Second)

	if err != nil {
		return
	}

	for _, stream := range streams {
		for _, msg := range stream.Messages {
			c.processTicket(ctx, msg)
			c.redis.XAck(ctx, TicketStreamKey, TicketConsumerGroup, msg.ID)
		}
	}
}

func (c *TicketConsumer) processTicket(ctx context.Context, msg redis.XMessage) {
	data, ok := msg.Values["data"].(string)
	if !ok {
		log.Printf("[MQ] Invalid ticket message format: %v", msg.Values)
		return
	}

	var ticketMsg TicketMessage
	if err := json.Unmarshal([]byte(data), &ticketMsg); err != nil {
		log.Printf("[MQ] Failed to unmarshal ticket message: %v", err)
		return
	}

	log.Printf("[MQ] Processing ticket for user %d, event %d", ticketMsg.UserID, ticketMsg.EventID)

	ticketType, err := c.ticketTypeRepo.FindByID(ticketMsg.TicketTypeID)
	if err != nil || ticketType == nil {
		log.Printf("[MQ] Ticket type %d not found", ticketMsg.TicketTypeID)
		c.sendResult(ticketMsg.UserID, 0, "failed", "票种不存在", "", "")
		return
	}

	// 原子扣减数据库库存
	if err := c.ticketTypeRepo.AtomicDeductStock(ticketMsg.TicketTypeID, ticketMsg.Quantity); err != nil {
		c.redis.SeckillRollback(ctx, fmt.Sprint(ticketMsg.EventID),
			fmt.Sprint(ticketMsg.TicketTypeID), fmt.Sprint(ticketMsg.UserID))
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
		OrderNo:      generateOrderNo(),
		QRCode:       uuid.New().String(),
	}

	if err := c.ticketRepo.Create(ticket); err != nil {
		log.Printf("[MQ] Failed to create ticket: %v", err)
		// 回滚已扣减的数据库库存
		c.ticketTypeRepo.UpdateStock(ticketMsg.TicketTypeID, ticketMsg.Quantity)
		c.redis.SeckillRollback(ctx, fmt.Sprint(ticketMsg.EventID),
			fmt.Sprint(ticketMsg.TicketTypeID), fmt.Sprint(ticketMsg.UserID))
		c.sendResult(ticketMsg.UserID, 0, "failed", "创建票务失败", "", "")
		return
	}

	c.sendResult(ticketMsg.UserID, ticket.ID, "success", "购票成功！", ticket.OrderNo, ticketType.Name)
	log.Printf("[MQ] Ticket %d created successfully for user %d", ticket.ID, ticketMsg.UserID)
}

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

func generateOrderNo() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("TK%s%s", time.Now().Format("20060102150405"), hex.EncodeToString(b))
}
