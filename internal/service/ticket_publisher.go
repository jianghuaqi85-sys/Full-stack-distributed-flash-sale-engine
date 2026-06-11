package service

import (
	"context"

	"order-system/internal/mq"
)

// TicketPublisher 票务消息发布接口
// 支持 Kafka 和 Redis Stream 两种实现
type TicketPublisher interface {
	// PublishTicket 发布票务消息（同步，确保消息不丢失）
	PublishTicket(ctx context.Context, msg *mq.TicketMessage) error
}
