package pkgredis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClientWrapper 包装原生 redis.Client
type RedisClientWrapper struct {
	client *redis.Client
}

func NewClientFromRaw(client *redis.Client) *RedisClientWrapper {
	return &RedisClientWrapper{client: client}
}

func (r *RedisClientWrapper) Client() *redis.Client {
	return r.client
}

// 秒杀库存扣减 Lua 脚本
var seckillDeductScript = redis.NewScript(`
	local stock_key = KEYS[1]
	local bought_key = KEYS[2]
	local product_id = ARGV[1]
	local user_id = ARGV[2]

	-- 检查库存
	local stock = tonumber(redis.call('HGET', stock_key, product_id))
	if stock == nil or stock <= 0 then
		return -1
	end

	-- 检查是否已购买
	if redis.call('SISMEMBER', bought_key, user_id) == 1 then
		return -2
	end

	-- 扣减库存
	redis.call('HINCRBY', stock_key, product_id, -1)
	-- 记录已购买
	redis.call('SADD', bought_key, user_id)

	return 1
`)

// 秒杀库存回滚 Lua 脚本
var seckillRollbackScript = redis.NewScript(`
	local stock_key = KEYS[1]
	local bought_key = KEYS[2]
	local product_id = ARGV[1]
	local user_id = ARGV[2]

	-- 回滚库存
	redis.call('HINCRBY', stock_key, product_id, 1)
	-- 移除购买记录
	redis.call('SREM', bought_key, user_id)

	return 1
`)

// SeckillDeduct 秒杀库存预扣（原子操作）
// 使用 hash tag {activityID} 确保 stock 和 bought key 路由到同一 slot（Redis Cluster 兼容）
func (r *RedisClientWrapper) SeckillDeduct(ctx context.Context, activityID, productID, userID string) (int64, error) {
	result, err := seckillDeductScript.Run(ctx, r.client,
		[]string{
			"seckill:{" + activityID + "}:stock",
			"seckill:{" + activityID + "}:bought",
		},
		productID, userID,
	).Int64()
	return result, err
}

// SeckillRollback 秒杀库存回滚
func (r *RedisClientWrapper) SeckillRollback(ctx context.Context, activityID, productID, userID string) error {
	_, err := seckillRollbackScript.Run(ctx, r.client,
		[]string{
			"seckill:{" + activityID + "}:stock",
			"seckill:{" + activityID + "}:bought",
		},
		productID, userID,
	).Result()
	return err
}

// InitSeckillStock 初始化秒杀库存
func (r *RedisClientWrapper) InitSeckillStock(ctx context.Context, activityID, productID string, stock int) error {
	key := "seckill:{" + activityID + "}:stock"
	return r.client.HSet(ctx, key, productID, stock).Err()
}

// GetSeckillStock 获取秒杀库存
func (r *RedisClientWrapper) GetSeckillStock(ctx context.Context, activityID, productID string) (int, error) {
	key := "seckill:{" + activityID + "}:stock"
	stock, err := r.client.HGet(ctx, key, productID).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return stock, err
}

// SetSeckillActivity 设置秒杀活动信息
func (r *RedisClientWrapper) SetSeckillActivity(ctx context.Context, activityID string, info map[string]interface{}) error {
	key := "seckill:{" + activityID + "}:info"
	return r.client.HSet(ctx, key, info).Err()
}

// GetSeckillActivity 获取秒杀活动信息
func (r *RedisClientWrapper) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

func (r *RedisClientWrapper) GetSeckillActivity(ctx context.Context, activityID string) (map[string]string, error) {
	key := "seckill:{" + activityID + "}:info"
	return r.client.HGetAll(ctx, key).Result()
}

// XAdd 添加消息到 Stream
func (r *RedisClientWrapper) XAdd(ctx context.Context, stream string, values map[string]interface{}) error {
	return r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Err()
}

// XReadGroup 从 Stream 读取消息
func (r *RedisClientWrapper) XReadGroup(ctx context.Context, group, consumer string, streams []string, count int64, block time.Duration) ([]redis.XStream, error) {
	return r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  streams,
		Count:    count,
		Block:    block,
	}).Result()
}

// XAck 确认消息
func (r *RedisClientWrapper) XAck(ctx context.Context, stream, group string, IDs ...string) error {
	return r.client.XAck(ctx, stream, group, IDs...).Err()
}

// XGroupCreate 创建消费者组
func (r *RedisClientWrapper) XGroupCreate(ctx context.Context, stream, group string) error {
	return r.client.XGroupCreateMkStream(ctx, stream, group, "0").Err()
}

// RateLimit 限流检查（Lua 脚本原子操作）
var rateLimitScript = redis.NewScript(`
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local current = tonumber(redis.call('GET', key) or '0')
if current >= limit then
    return 0
end
current = redis.call('INCR', key)
if current == 1 then
    redis.call('EXPIRE', key, window)
end
return current <= limit and 1 or 0
`)

func (r *RedisClientWrapper) RateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	result, err := rateLimitScript.Run(ctx, r.client, []string{key}, limit, int(window.Seconds())).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}
