package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client       *redis.Client
	clusterClient *redis.ClusterClient
	isCluster     bool
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

// NewClusterClient 创建 Redis Cluster 客户端
func NewClusterClient(addrs []string, password string) *RedisClient {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addrs,
		Password:     password,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolSize:     100,
		MinIdleConns: 10,
	})
	return &RedisClient{clusterClient: client, isCluster: true}
}

// IsCluster 是否为 Cluster 模式
func (r *RedisClient) IsCluster() bool {
	return r.isCluster
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
	if r.isCluster {
		return r.clusterClient.Get(ctx, key).Result()
	}
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if r.isCluster {
		return r.clusterClient.Set(ctx, key, value, expiration).Err()
	}
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	if r.isCluster {
		return r.clusterClient.Del(ctx, keys...).Err()
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisClient) Close() error {
	if r.isCluster {
		return r.clusterClient.Close()
	}
	return r.client.Close()
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}

// ClusterClient 获取 Cluster 客户端（仅 Cluster 模式可用）
func (r *RedisClient) ClusterClient() *redis.ClusterClient {
	return r.clusterClient
}

func (r *RedisClient) Ping(ctx context.Context) error {
	if r.isCluster {
		return r.clusterClient.Ping(ctx).Err()
	}
	return r.client.Ping(ctx).Err()
}

// XAdd 添加消息到 Stream
func (r *RedisClient) XAdd(ctx context.Context, stream string, values map[string]interface{}) error {
	if r.isCluster {
		return r.clusterClient.XAdd(ctx, &redis.XAddArgs{
			Stream: stream,
			Values: values,
		}).Err()
	}
	return r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Err()
}

// XReadGroup 从 Stream 读取消息
func (r *RedisClient) XReadGroup(ctx context.Context, group, consumer string, streams []string, count int64, block time.Duration) ([]redis.XStream, error) {
	if r.isCluster {
		return r.clusterClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  streams,
			Count:    count,
			Block:    block,
		}).Result()
	}
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
	if r.isCluster {
		return r.clusterClient.XAck(ctx, stream, group, IDs...).Err()
	}
	return r.client.XAck(ctx, stream, group, IDs...).Err()
}

// XGroupCreate 创建消费者组
func (r *RedisClient) XGroupCreate(ctx context.Context, stream, group string) error {
	if r.isCluster {
		return r.clusterClient.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	}
	return r.client.XGroupCreateMkStream(ctx, stream, group, "0").Err()
}

// HIncrBy Hash 字段自增
func (r *RedisClient) HIncrBy(ctx context.Context, key, field string, incr int64) error {
	if r.isCluster {
		return r.clusterClient.HIncrBy(ctx, key, field, incr).Err()
	}
	return r.client.HIncrBy(ctx, key, field, incr).Err()
}

// SAdd 添加元素到 Set
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if r.isCluster {
		return r.clusterClient.SAdd(ctx, key, members...).Err()
	}
	return r.client.SAdd(ctx, key, members...).Err()
}

// SRem 移除 Set 元素
func (r *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	if r.isCluster {
		return r.clusterClient.SRem(ctx, key, members...).Err()
	}
	return r.client.SRem(ctx, key, members...).Err()
}

// HGet 获取 Hash 字段
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	if r.isCluster {
		return r.clusterClient.HGet(ctx, key, field).Result()
	}
	return r.client.HGet(ctx, key, field).Result()
}

// HSet 设置 Hash 字段
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	if r.isCluster {
		return r.clusterClient.HSet(ctx, key, values...).Err()
	}
	return r.client.HSet(ctx, key, values...).Err()
}

// Pipeline 获取 pipeline
func (r *RedisClient) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// Incr 自增
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	if r.isCluster {
		return r.clusterClient.Incr(ctx, key).Result()
	}
	return r.client.Incr(ctx, key).Result()
}

// Expire 设置过期时间
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if r.isCluster {
		return r.clusterClient.Expire(ctx, key, expiration).Err()
	}
	return r.client.Expire(ctx, key, expiration).Err()
}

