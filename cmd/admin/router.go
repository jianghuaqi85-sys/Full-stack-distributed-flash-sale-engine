package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"order-system/internal/handler"
	"order-system/internal/middleware"
	"order-system/internal/pkg/constants"
	"order-system/internal/pkg/redis/pkgredis"
	"order-system/internal/repository"
	"order-system/internal/service"
)

func NewAdminRouter(db *gorm.DB, jwtSecret string, redisClient *redis.Client, allowOrigins []string) *gin.Engine {
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

	// === 服务层 ===
	authService := service.NewAuthService(userRepo, jwtSecret, constants.DefaultJWTExpire, redisClient)
	eventService := service.NewEventService(eventRepo, ticketTypeRepo, redisWrapper)

	// === Handler 层 ===
	authHandler := handler.NewAuthHandler(authService)
	healthHandler := handler.NewHealthHandler(db, redisClient)
	eventHandler := handler.NewEventHandler(eventService)

	// 促销码服务
	promoCodeRepo := repository.NewPromoCodeRepository(db)
	promoCodeService := service.NewPromoCodeService(promoCodeRepo)
	promoCodeHandler := handler.NewPromoCodeHandler(promoCodeService)

	// 统计服务（带 Redis 缓存）
	statsService := service.NewStatsServiceWithRedis(db, redisClient)
	statsHandler := handler.NewStatsHandler(statsService)

	// 票务转让服务
	transferRepo := repository.NewTicketTransferRepository(db)
	transferService := service.NewTicketTransferService(ticketRepo, transferRepo, userRepo)
	transferHandler := handler.NewTicketTransferHandler(transferService)

	// 场次服务
	showRepo := repository.NewShowRepository(db)
	showService := service.NewShowService(showRepo, eventRepo, ticketTypeRepo)
	showHandler := handler.NewShowHandler(showService)

	log.Printf("[Admin] Admin service initialized")

	// ========== 路由注册 ==========

	r.GET("/health", healthHandler.HealthCheck)

	// 管理员路由（全部需要认证 + 管理员角色）
	admin := r.Group("/admin")
	admin.Use(middleware.JWTAuthWithBlacklist(db, jwtSecret, redisClient))
	admin.Use(middleware.RoleAuth("admin"))
	{
		// 活动管理
		adminEvent := admin.Group("/events")
		{
			adminEvent.POST("", eventHandler.CreateEvent)
			adminEvent.PUT("/:id", eventHandler.UpdateEvent)
			adminEvent.POST("/:id/publish", eventHandler.PublishEvent)
			adminEvent.POST("/:id/unpublish", eventHandler.UnpublishEvent)
			adminEvent.POST("/:id/end", eventHandler.EndEvent)
			adminEvent.POST("/:id/ticket-types", eventHandler.CreateTicketType)
			adminEvent.PUT("/ticket-types/:id", eventHandler.UpdateTicketType)
			adminEvent.DELETE("/ticket-types/:id", eventHandler.DeleteTicketType)

			// 场次管理
			adminEvent.POST("/:id/shows", showHandler.CreateShow)
			adminEvent.PUT("/shows/:id", showHandler.UpdateShow)
			adminEvent.DELETE("/shows/:id", showHandler.DeleteShow)
			adminEvent.POST("/shows/:id/publish", showHandler.PublishShow)
			adminEvent.POST("/shows/:id/unpublish", showHandler.UnpublishShow)
		}

		// 促销码管理
		adminPromo := admin.Group("/promo")
		{
			adminPromo.POST("", promoCodeHandler.CreatePromoCode)
			adminPromo.DELETE("/:id", promoCodeHandler.DeletePromoCode)
		}

		// 统计数据
		adminStats := admin.Group("/stats")
		{
			adminStats.GET("/dashboard", statsHandler.GetDashboardStats)
			adminStats.GET("/sales-trend", statsHandler.GetSalesTrend)
			adminStats.GET("/ticket-types", statsHandler.GetTicketTypeStats)
			adminStats.GET("/funnel/:event_id", statsHandler.GetConversionFunnel)
		}

		// 票务转让审核
		adminTransfer := admin.Group("/transfer")
		{
			adminTransfer.GET("/pending", transferHandler.GetPendingTransfers)
			adminTransfer.POST("/:id/approve", transferHandler.ApproveTransfer)
			adminTransfer.POST("/:id/reject", transferHandler.RejectTransfer)
		}
	}

	// 管理员也需要登录接口
	public := r.Group("")
	public.Use(middleware.RateLimitMiddleware(redisClient, constants.PublicRateLimit, constants.RateLimitWindow))
	{
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
	}

	return r
}
