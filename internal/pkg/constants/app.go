package constants

import "time"

const (
	// 票务过期时间
	TicketExpireDuration = 30 * time.Minute

	// 限流配置
	PublicRateLimit    = 100 // 公共接口每分钟请求数
	SeckillRateLimit  = 10  // 秒杀每秒每用户请求数
	RateLimitWindow   = time.Minute
	SeckillWindow     = time.Second

	// JWT 配置
	DefaultJWTExpire = 86400 // 默认 JWT 过期时间（秒）

	// 分页配置
	DefaultPageLimit = 20
	MaxPageLimit     = 100
)
