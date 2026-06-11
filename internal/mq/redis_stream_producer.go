package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"order-system/internal/pkg/redis/pkgredis"
)

// RedisStreamProducer Redis Stream 消息生产者（降级方案）
type RedisStreamProducer struct {
	redis *pkgredis.RedisClientWrapper
}

// NewRedisStreamProducer 创建 Redis Stream 生产者
func NewRedisStreamProducer(redis *pkgredis.RedisClientWrapper) *RedisStreamProducer {
	return &RedisStreamProducer{redis: redis}
}

// PublishTicket 发布票务消息到 Redis Stream
func (p *RedisStreamProducer) PublishTicket(ctx context.Context, msg *TicketMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal ticket message: %w", err)
	}

	return p.redis.XAdd(ctx, TicketStreamKey, map[string]interface{}{
		"data": string(data),
	})
}
