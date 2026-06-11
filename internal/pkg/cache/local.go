package cache

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

// LocalCache 基于 ristretto 的本地缓存封装
type LocalCache struct {
	cache *ristretto.Cache[string, []byte]
}

// NewLocalCache 创建本地缓存
// numCounters: 用于跟踪频率的计数器数量（建议为 maxItems 的 10 倍）
// maxCost: 缓存最大成本（字节数）
func NewLocalCache(numCounters int64, maxCost int64) (*LocalCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
		NumCounters: numCounters,
		MaxCost:     maxCost,
		BufferItems: 64, // 异步写缓冲区大小
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
	}

	return &LocalCache{cache: cache}, nil
}

// Get 获取缓存值
func (c *LocalCache) Get(key string) ([]byte, bool) {
	return c.cache.Get(key)
}

// GetString 获取缓存值（字符串）
func (c *LocalCache) GetString(key string) (string, bool) {
	val, found := c.cache.Get(key)
	if !found {
		return "", false
	}
	return string(val), true
}

// Set 设置缓存值（带 TTL）
func (c *LocalCache) Set(key string, value []byte, ttl time.Duration) bool {
	return c.cache.SetWithTTL(key, value, 1, ttl)
}

// SetString 设置缓存值（字符串，带 TTL）
func (c *LocalCache) SetString(key, value string, ttl time.Duration) bool {
	return c.cache.SetWithTTL(key, []byte(value), 1, ttl)
}

// SetWithCost 设置缓存值（带自定义成本和 TTL）
func (c *LocalCache) SetWithCost(key string, value []byte, cost int64, ttl time.Duration) bool {
	return c.cache.SetWithTTL(key, value, cost, ttl)
}

// Delete 删除缓存值
func (c *LocalCache) Delete(key string) {
	c.cache.Del(key)
}

// WaitForWrite 等待所有异步写入完成
func (c *LocalCache) WaitForWrite() {
	c.cache.Wait()
}

// Close 关闭缓存
func (c *LocalCache) Close() {
	c.cache.Close()
}

// === 便捷方法：库存快照缓存 ===

const (
	stockCachePrefix = "stock:"
	stockCacheTTL    = 1 * time.Second // 库存快照 TTL 极短，避免脏读
)

// StockCacheKey 生成库存缓存 key
func StockCacheKey(eventID, ticketTypeID uint) string {
	return fmt.Sprintf("%s%d:%d", stockCachePrefix, eventID, ticketTypeID)
}
