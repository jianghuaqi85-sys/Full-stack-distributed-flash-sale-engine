package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"order-system/internal/pkg/redis/pkgredis"
)

type TicketProducer struct {
	redis *pkgredis.RedisClientWrapper
}

func NewTicketProducer(redis *pkgredis.RedisClientWrapper) *TicketProducer {
	return &TicketProducer{redis: redis}
}

func (p *TicketProducer) PublishTicket(ctx context.Context, msg *TicketMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal ticket message: %w", err)
	}

	return p.redis.XAdd(ctx, TicketStreamKey, map[string]interface{}{
		"data": string(data),
	})
}
