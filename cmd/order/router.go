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
	"order-system/internal/pkg/constants"
	"order-system/internal/pkg/redis/pkgredis"
	"order-system/internal/pkg/ws"
	"order-system/internal/repository"
	"order-system/internal/service"
)

func NewOrderRouter(ctx context.Context, db *gorm.DB, jwtSecret string, redisClient *redis.Client, allowOrigins []string) *gin.Engine {
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

	// === WebSocket Hub ===
	hub := ws.NewHub(jwtSecret, redisClient)
	if len(allowOrigins) > 0 {
		hub.SetAllowedOrigins(allowOrigins)
	}

	// === 服务层 ===
	authService := service.NewAuthService(userRepo, jwtSecret, constants.DefaultJWTExpire, redisClient)
	ticketService := service.NewTicketService(db, redisWrapper, nil, ticketRepo, ticketTypeRepo, eventRepo) // 无 producer，订单服务不直接购票

	// === Handler 层 ===
	authHandler := handler.NewAuthHandler(authService)
	healthHandler := handler.NewHealthHandler(db, redisClient)
	ticketHandler := handler.NewTicketHandler(ticketService)

	// 票务转让服务
	transferRepo := repository.NewTicketTransferRepository(db)
	transferService := service.NewTicketTransferService(ticketRepo, transferRepo, userRepo)
	transferHandler := handler.NewTicketTransferHandler(transferService)

	// 二手市场服务
	marketplaceRepo := repository.NewMarketplaceRepository(db)
	marketplaceService := service.NewMarketplaceService(db, marketplaceRepo, ticketRepo, ticketTypeRepo, userRepo, eventRepo)
	marketplaceHandler := handler.NewMarketplaceHandler(marketplaceService)

	// 促销码服务
	promoCodeRepo := repository.NewPromoCodeRepository(db)
	promoCodeService := service.NewPromoCodeService(promoCodeRepo)
	promoCodeHandler := handler.NewPromoCodeHandler(promoCodeService)

	// 启动票务过期检查
	go service.StartTicketExpireChecker(ctx, ticketRepo, ticketTypeRepo, redisWrapper, 1*time.Minute)

	log.Printf("[Order] Order service initialized")

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

	// 订单核心路由
	api := r.Group("/api")
	api.Use(middleware.JWTAuthWithBlacklist(db, jwtSecret, redisClient))
	{
		api.GET("/profile", authHandler.GetProfile)

		// 票务管理（不含秒杀）
		tickets := api.Group("/tickets")
		{
			tickets.GET("", ticketHandler.GetMyTickets)
			tickets.GET("/:id", ticketHandler.GetTicketDetail)
			tickets.POST("/:id/pay", ticketHandler.PayTicket)
			tickets.POST("/:id/cancel", ticketHandler.CancelTicket)
			tickets.POST("/:id/use", ticketHandler.UseTicket)
		}

		// 票务转让
		transferRoutes := api.Group("/transfer")
		{
			transferRoutes.POST("", transferHandler.RequestTransfer)
			transferRoutes.POST("/gift", transferHandler.DirectGift)
			transferRoutes.GET("/history", transferHandler.GetTransferHistory)
		}

		// 二手市场
		marketplaceRoutes := api.Group("/marketplace")
		{
			marketplaceRoutes.GET("", marketplaceHandler.ListActive)
			marketplaceRoutes.GET("/my", marketplaceHandler.ListMyListings)
			marketplaceRoutes.GET("/purchases", marketplaceHandler.ListMyPurchases)
			marketplaceRoutes.GET("/event/:id", marketplaceHandler.ListByEvent)
			marketplaceRoutes.GET("/:id", marketplaceHandler.GetListing)
			marketplaceRoutes.POST("", marketplaceHandler.CreateListing)
			marketplaceRoutes.POST("/:id/buy", marketplaceHandler.BuyListing)
			marketplaceRoutes.POST("/:id/cancel", marketplaceHandler.CancelListing)
		}

		// 促销码
		promoRoutes := api.Group("/promo")
		{
			promoRoutes.POST("/validate", promoCodeHandler.ValidatePromoCode)
			promoRoutes.GET("/:event_id", promoCodeHandler.GetPromoCodes)
		}
	}

	r.GET("/ws", func(c *gin.Context) {
		hub.ServeWS(c)
	})

	return r
}
