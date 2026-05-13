package router

import (
	"context"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"order-system/internal/handler"
	"order-system/internal/middleware"
	"order-system/internal/mq"
	"order-system/internal/pkg/constants"
	"order-system/internal/pkg/redis/pkgredis"
	"order-system/internal/pkg/ws"
	"order-system/internal/queue"
	"order-system/internal/repository"
	"order-system/internal/service"
)

func NewRouter(ctx context.Context, db *gorm.DB, jwtSecret string, redisClient *redis.Client) *gin.Engine {
	r := gin.New()

	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.OpenTelemetry())
	r.Use(middleware.ErrorHandler())
	r.Use(gin.Logger())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	ticketTypeRepo := repository.NewTicketTypeRepository(db)
	ticketRepo := repository.NewTicketRepository(db)

	// 初始化 Redis 客户端
	redisWrapper := pkgredis.NewClientFromRaw(redisClient)

	authService := service.NewAuthService(userRepo, jwtSecret, constants.DefaultJWTExpire)
	eventService := service.NewEventService(eventRepo, ticketTypeRepo, redisWrapper)

	authHandler := handler.NewAuthHandler(authService)
	healthHandler := handler.NewHealthHandler(db, redisClient)
	eventHandler := handler.NewEventHandler(eventService)

	hub := ws.NewHub(jwtSecret, redisClient)

	// 初始化票务消息队列
	ticketProducer := mq.NewTicketProducer(redisWrapper)

	// 初始化票务服务
	ticketService := service.NewTicketService(redisWrapper, ticketProducer, ticketRepo, ticketTypeRepo, eventRepo)
	ticketHandler := handler.NewTicketHandler(ticketService)

	// 启动票务消费者
	ticketConsumer := mq.NewTicketConsumer(redisWrapper, db, hub)
	go ticketConsumer.Start(ctx)

	// 启动票务过期检查（每分钟检查一次）
	go service.StartTicketExpireChecker(ctx, ticketRepo, ticketTypeRepo, redisWrapper, 1*time.Minute)

	// 初始化排队管理器
	queueManager := queue.NewQueueManager(redisClient)
	queueHandler := handler.NewQueueHandler(queueManager, hub)

	// 初始化等候名单管理器
	waitlistManager := queue.NewWaitlistManager(redisClient)
	waitlistHandler := handler.NewWaitlistHandler(waitlistManager)

	// 初始化促销码服务
	promoCodeRepo := repository.NewPromoCodeRepository(db)
	promoCodeService := service.NewPromoCodeService(promoCodeRepo)
	promoCodeHandler := handler.NewPromoCodeHandler(promoCodeService)

	// 初始化统计服务
	statsService := service.NewStatsService(db)
	statsHandler := handler.NewStatsHandler(statsService)

	// 初始化票务转让服务
	transferRepo := repository.NewTicketTransferRepository(db)
	transferService := service.NewTicketTransferService(ticketRepo, transferRepo, userRepo)
	transferHandler := handler.NewTicketTransferHandler(transferService)

	// 初始化场次服务
	showRepo := repository.NewShowRepository(db)
	showService := service.NewShowService(showRepo, eventRepo, ticketTypeRepo)
	showHandler := handler.NewShowHandler(showService)

	// 初始化转让市场服务
	marketplaceRepo := repository.NewMarketplaceRepository(db)
	marketplaceService := service.NewMarketplaceService(marketplaceRepo, ticketRepo, ticketTypeRepo, userRepo)
	marketplaceHandler := handler.NewMarketplaceHandler(marketplaceService)

	r.GET("/health", healthHandler.HealthCheck)

	public := r.Group("")
	public.Use(middleware.RateLimitMiddleware(redisClient, constants.PublicRateLimit, constants.RateLimitWindow))
	{
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
	}

	// API 认证路由（前端通过 /api 前缀访问）
	apiAuth := r.Group("/api")
	{
		apiAuth.POST("/auth/register", authHandler.Register)
		apiAuth.POST("/auth/login", authHandler.Login)
	}

	api := r.Group("/api")
	api.Use(middleware.JWTAuthWithBlacklist(db, jwtSecret, redisClient))
	{
		api.GET("/profile", authHandler.GetProfile)

		// 活动浏览路由
		events := api.Group("/events")
		{
			events.GET("", eventHandler.ListEvents)
			events.GET("/:id", eventHandler.GetEvent)
			events.GET("/:id/stock", eventHandler.GetEventStock)
			events.GET("/:id/shows", showHandler.ListShows)
		}

		// 场次路由
		shows := api.Group("/shows")
		{
			shows.GET("/:id", showHandler.GetShow)
		}

		// 排队路由
		queueRoutes := api.Group("/queue")
		{
			queueRoutes.POST("/:event_id/join", queueHandler.JoinQueue)
			queueRoutes.GET("/:event_id/position", queueHandler.GetPosition)
			queueRoutes.POST("/:event_id/leave", queueHandler.LeaveQueue)
		}

		// 等候名单路由
		waitlistRoutes := api.Group("/waitlist")
		{
			waitlistRoutes.POST("/:event_id/join", waitlistHandler.JoinWaitlist)
			waitlistRoutes.GET("/:event_id/position", waitlistHandler.GetWaitlistPosition)
			waitlistRoutes.POST("/:event_id/leave", waitlistHandler.LeaveWaitlist)
		}

		// 促销码路由
		promoRoutes := api.Group("/promo")
		{
			promoRoutes.POST("/validate", promoCodeHandler.ValidatePromoCode)
			promoRoutes.GET("/:event_id", promoCodeHandler.GetPromoCodes)
		}

		// 票务路由
		tickets := api.Group("/tickets")
		{
			tickets.Use(middleware.SeckillRateLimitMiddleware(redisWrapper, constants.SeckillRateLimit, constants.SeckillWindow))
			tickets.POST("/purchase", ticketHandler.PurchaseTicket)
		}
		ticketsNoLimit := api.Group("/tickets")
		{
			ticketsNoLimit.GET("", ticketHandler.GetMyTickets)
			ticketsNoLimit.GET("/:id", ticketHandler.GetTicketDetail)
			ticketsNoLimit.POST("/:id/pay", ticketHandler.PayTicket)
			ticketsNoLimit.POST("/:id/cancel", ticketHandler.CancelTicket)
			ticketsNoLimit.POST("/:id/use", ticketHandler.UseTicket)
		}

		// 票务转让路由
		transferRoutes := api.Group("/transfer")
		{
			transferRoutes.POST("", transferHandler.RequestTransfer)
			transferRoutes.POST("/gift", transferHandler.DirectGift)
			transferRoutes.GET("/history", transferHandler.GetTransferHistory)
		}

		// 转让市场路由
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
	}

	// 管理员路由
	admin := r.Group("/admin")
	admin.Use(middleware.JWTAuthWithBlacklist(db, jwtSecret, redisClient))
	admin.Use(middleware.RoleAuth("admin"))
	{
		// 活动管理路由
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

			// 场次管理路由
			adminEvent.POST("/:id/shows", showHandler.CreateShow)
			adminEvent.PUT("/shows/:id", showHandler.UpdateShow)
			adminEvent.DELETE("/shows/:id", showHandler.DeleteShow)
			adminEvent.POST("/shows/:id/publish", showHandler.PublishShow)
			adminEvent.POST("/shows/:id/unpublish", showHandler.UnpublishShow)
		}

		// 促销码管理路由
		adminPromo := admin.Group("/promo")
		{
			adminPromo.POST("", promoCodeHandler.CreatePromoCode)
			adminPromo.DELETE("/:id", promoCodeHandler.DeletePromoCode)
		}

		// 统计数据路由
		adminStats := admin.Group("/stats")
		{
			adminStats.GET("/dashboard", statsHandler.GetDashboardStats)
			adminStats.GET("/sales-trend", statsHandler.GetSalesTrend)
			adminStats.GET("/ticket-types", statsHandler.GetTicketTypeStats)
			adminStats.GET("/funnel/:event_id", statsHandler.GetConversionFunnel)
		}

		// 票务转让审核路由
		adminTransfer := admin.Group("/transfer")
		{
			adminTransfer.GET("/pending", transferHandler.GetPendingTransfers)
			adminTransfer.POST("/:id/approve", transferHandler.ApproveTransfer)
			adminTransfer.POST("/:id/reject", transferHandler.RejectTransfer)
		}
	}

	r.GET("/ws", hub.ServeWS)

	return r
}
