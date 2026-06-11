package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"order-system/internal/handler"
	"order-system/internal/middleware"
	"order-system/internal/mq"
	"order-system/internal/pkg/cache"
	"order-system/internal/pkg/constants"
	"order-system/internal/pkg/redis/pkgredis"
	"order-system/internal/pkg/ws"
	"order-system/internal/queue"
	"order-system/internal/repository"
	"order-system/internal/service"
)

func NewSeckillRouter(ctx context.Context, db *gorm.DB, jwtSecret string, redisClient *redis.Client, allowOrigins []string, kafkaBrokers []string, kafkaEnabled bool) *gin.Engine {
	r := gin.New()

	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.OpenTelemetry())
	r.Use(middleware.ErrorHandler())
	r.Use(gin.Logger())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// === 基础设施 ===
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	ticketTypeRepo := repository.NewTicketTypeRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	redisWrapper := pkgredis.NewClientFromRaw(redisClient)

	localCache, err := cache.NewLocalCache(100_000, 64<<20)
	if err != nil {
		log.Printf("[Seckill] Failed to create local cache: %v", err)
	}

	// === WebSocket Hub ===
	hub := ws.NewHub(jwtSecret, redisClient)
	if len(allowOrigins) > 0 {
		hub.SetAllowedOrigins(allowOrigins)
	}

	// === 消息队列 ===
	var ticketProducer service.TicketPublisher

	if kafkaEnabled && len(kafkaBrokers) > 0 {
		log.Printf("[Seckill] Using Kafka: brokers=%v", kafkaBrokers)
		kafkaProducer, err := mq.NewTicketProducer(kafkaBrokers)
		if err != nil {
			log.Printf("[Seckill] Failed to create Kafka producer: %v, falling back to Redis Stream", err)
			kafkaEnabled = false
		} else {
			ticketProducer = kafkaProducer
			kafkaConsumer, err := mq.NewTicketConsumer(kafkaBrokers, db, hub, redisWrapper, localCache)
			if err != nil {
				log.Printf("[Seckill] Failed to create Kafka consumer: %v", err)
			} else {
				go kafkaConsumer.Start(ctx)
			}
		}
	}

	if !kafkaEnabled {
		log.Printf("[Seckill] Using Redis Stream")
		redisProducer := mq.NewRedisStreamProducer(redisWrapper)
		ticketProducer = redisProducer
		redisConsumer := mq.NewRedisStreamConsumer(redisWrapper, db, hub)
		go redisConsumer.Start(ctx)
	}

	// === 服务层 ===
	authService := service.NewAuthService(userRepo, jwtSecret, constants.DefaultJWTExpire, redisClient)
	ticketService := service.NewTicketService(db, redisWrapper, ticketProducer, ticketRepo, ticketTypeRepo, eventRepo)
	eventService := service.NewEventService(eventRepo, ticketTypeRepo, redisWrapper)

	// === Handler 层 ===
	authHandler := handler.NewAuthHandler(authService)
	healthHandler := handler.NewHealthHandler(db, redisClient)
	ticketHandler := handler.NewTicketHandler(ticketService)
	eventHandler := handler.NewEventHandler(eventService)

	// 排队管理器
	queueManager := queue.NewQueueManager(redisClient)
	queueHandler := handler.NewQueueHandler(queueManager, hub)

	// 启动票务过期检查
	go service.StartTicketExpireChecker(ctx, ticketRepo, ticketTypeRepo, redisWrapper, 1*time.Minute)

	// ========== 路由注册 ==========

	r.GET("/health", healthHandler.HealthCheck)

	// 公开接口
	public := r.Group("")
	public.Use(middleware.RateLimitMiddleware(redisClient, constants.PublicRateLimit, constants.RateLimitWindow))
	{
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
	}

	// API 认证路由（前端兼容）
	apiAuth := r.Group("/api")
	{
		apiAuth.POST("/auth/register", authHandler.Register)
		apiAuth.POST("/auth/login", authHandler.Login)
	}

	// 秒杀核心路由
	api := r.Group("/api")
	api.Use(middleware.JWTAuthWithBlacklist(db, jwtSecret, redisClient))
	{
		// 活动浏览
		api.GET("/events", eventHandler.ListEvents)
		api.GET("/events/:id", eventHandler.GetEvent)
		api.GET("/events/:id/stock", eventHandler.GetEventStock)

		// 排队
		queueRoutes := api.Group("/queue")
		{
			queueRoutes.POST("/:event_id/join", queueHandler.JoinQueue)
			queueRoutes.GET("/:event_id/position", queueHandler.GetPosition)
			queueRoutes.POST("/:event_id/leave", queueHandler.LeaveQueue)
		}

		// 秒杀购票（限流 10/s/user）
		tickets := api.Group("/tickets")
		tickets.Use(middleware.SeckillRateLimitMiddleware(redisWrapper, constants.SeckillRateLimit, constants.SeckillWindow))
		{
			tickets.POST("/purchase", ticketHandler.PurchaseTicket)
		}

		// 票务状态查询
		api.GET("/tickets", ticketHandler.GetMyTickets)
		api.GET("/tickets/:id", ticketHandler.GetTicketDetail)
	}

	r.GET("/ws", func(c *gin.Context) {
		hub.ServeWS(c)
	})

	return r
}
