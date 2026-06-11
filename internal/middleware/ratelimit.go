package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 限流 Lua 脚本（原子操作，避免 Pipeline 非原子问题）
var rateLimitLuaScript = redis.NewScript(`
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

if current <= limit then
    return 1
else
    return 0
end
`)

func RateLimitMiddleware(redisClient *redis.Client, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s", c.ClientIP())

		ctx := c.Request.Context()

		result, err := rateLimitLuaScript.Run(ctx, redisClient,
			[]string{key}, limit, int64(window.Seconds())).Int()
		if err != nil {
			// Redis 故障时放行（fail-open）
			c.Next()
			return
		}

		if result == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Next()
	}
}
