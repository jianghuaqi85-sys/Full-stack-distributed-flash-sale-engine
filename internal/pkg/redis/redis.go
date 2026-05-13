package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewClient(addr, password string, db int) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolSize:     100,
		MinIdleConns: 10,
	})
	return &RedisClient{client: client}
}

func (r *RedisClient) RateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	now := time.Now().Unix()
	pipeline := r.client.Pipeline()

	pipeline.Incr(ctx, key)
	pipeline.ExpireAt(ctx, key, time.Unix(now+int64(window.Seconds()), 0))

	results, err := pipeline.Exec(ctx)
	if err != nil {
		return false, err
	}

	count, _ := results[0].(*redis.IntCmd).Result()
	return count <= limit, nil
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// XAdd 添加消息到 Stream
func (r *RedisClient) XAdd(ctx context.Context, stream string, values map[string]interface{}) error {
	return r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Err()
}

// XReadGroup 从 Stream 读取消息
func (r *RedisClient) XReadGroup(ctx context.Context, group, consumer string, streams []string, count int64, block time.Duration) ([]redis.XStream, error) {
	return r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  streams,
		Count:    count,
		Block:    block,
	}).Result()
}

// XAck 确认消息
func (r *RedisClient) XAck(ctx context.Context, stream, group string, IDs ...string) error {
	return r.client.XAck(ctx, stream, group, IDs...).Err()
}

// XGroupCreate 创建消费者组
func (r *RedisClient) XGroupCreate(ctx context.Context, stream, group string) error {
	return r.client.XGroupCreateMkStream(ctx, stream, group, "0").Err()
}

// HIncrBy Hash 字段自增
func (r *RedisClient) HIncrBy(ctx context.Context, key, field string, incr int64) error {
	return r.client.HIncrBy(ctx, key, field, incr).Err()
}

// SAdd 添加元素到 Set
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SRem 移除 Set 元素
func (r *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// HGet 获取 Hash 字段
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HSet 设置 Hash 字段
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// Pipeline 获取 pipeline
func (r *RedisClient) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// Incr 自增
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Expire 设置过期时间
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

