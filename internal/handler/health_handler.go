package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func NewHealthHandler(db *gorm.DB, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redisClient: redisClient}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	dbStatus := h.checkDatabase()
	redisStatus := h.checkRedis()

	status := "healthy"
	if dbStatus != "up" || redisStatus != "up" {
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    status,
		"timestamp": time.Now().Unix(),
		"services": gin.H{
			"database": dbStatus,
			"redis":    redisStatus,
		},
	})
}

func (h *HealthHandler) checkDatabase() string {
	sqlDB, err := h.db.DB()
	if err != nil {
		return "down"
	}

	if err := sqlDB.Ping(); err != nil {
		return "down"
	}

	return "up"
}

func (h *HealthHandler) checkRedis() string {
	if h.redisClient == nil {
		return "down"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		return "down"
	}

	return "up"
}
