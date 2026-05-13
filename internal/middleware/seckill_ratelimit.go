package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"order-system/internal/pkg/db"
	"order-system/internal/pkg/redis/pkgredis"
)

// SeckillRateLimitMiddleware 用户级秒杀限流
func SeckillRateLimitMiddleware(redisClient *pkgredis.RedisClientWrapper, maxPerUser int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			return
		}

		userModel, ok := user.(db.User)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "用户类型错误"})
			return
		}

		key := fmt.Sprintf("seckill:ratelimit:user:%d", userModel.ID)

		allowed, err := redisClient.RateLimit(c.Request.Context(), key, maxPerUser, window)
		if err != nil {
			// Redis 出错时放行
			c.Next()
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "操作过于频繁，请稍后再试",
				"retry_after": window.Seconds(),
			})
			return
		}

		c.Next()
	}
}

// StockAwareLimitMiddleware 库存感知限流
func StockAwareLimitMiddleware(redisClient *pkgredis.RedisClientWrapper) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityID := c.Param("activity_id")
		productID := c.Query("product_id")

		if activityID == "" || productID == "" {
			c.Next()
			return
		}

		stock, err := redisClient.GetSeckillStock(c.Request.Context(), activityID, productID)
		if err != nil {
			// Redis 出错时放行
			c.Next()
			return
		}

		if stock <= 0 {
			c.AbortWithStatusJSON(http.StatusGone, gin.H{"error": "已售罄"})
			return
		}

		c.Next()
	}
}
