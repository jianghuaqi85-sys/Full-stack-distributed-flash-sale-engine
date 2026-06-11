package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"order-system/internal/pkg/kafka"
)

// TicketProducer 票务消息生产者（Kafka）
type TicketProducer struct {
	producer *kafka.Producer
}

// NewTicketProducer 创建 Kafka 票务生产者
func NewTicketProducer(brokers []string) (*TicketProducer, error) {
	producer, err := kafka.NewProducer(brokers, TicketOrderTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket producer: %w", err)
	}
	return &TicketProducer{producer: producer}, nil
}

// PublishTicket 发布票务消息到 Kafka（同步，确保消息不丢失）
func (p *TicketProducer) PublishTicket(ctx context.Context, msg *TicketMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal ticket message: %w", err)
	}

	// 使用 UserID 作为 key，确保同一用户的消息路由到同一 partition
	key := []byte(fmt.Sprintf("%d", msg.UserID))
	return p.producer.ProduceSync(ctx, key, data)
}

// PublishTicketAsync 异步发布票务消息（高性能路径，适用于秒杀高峰）
func (p *TicketProducer) PublishTicketAsync(ctx context.Context, msg *TicketMessage, cb func(error)) {
	data, err := json.Marshal(msg)
	if err != nil {
		if cb != nil {
			cb(fmt.Errorf("failed to marshal ticket message: %w", err))
		}
		return
	}

	key := []byte(fmt.Sprintf("%d", msg.UserID))
	p.producer.ProduceAsync(ctx, key, data, cb)
}

// Close 关闭生产者
func (p *TicketProducer) Close() {
	p.producer.Close()
}
