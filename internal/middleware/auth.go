package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	pkgdb "order-system/internal/pkg/db"
)

func JWTAuth(db *gorm.DB, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in token"})
			return
		}

		var user pkgdb.User
		if err := db.First(&user, uint(userID)).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func JWTAuthWithBlacklist(db *gorm.DB, secret string, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			return
		}

		ctx := c.Request.Context()

		// 检查 token 黑名单
		if exists, err := redisClient.Exists(ctx, "blacklist:"+tokenString).Result(); err == nil && exists > 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			return
		}

		// 解析 JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in token"})
			return
		}

		// 从 Redis 缓存读取用户信息，避免每次查库
		var user pkgdb.User
		cacheKey := fmt.Sprintf("auth:user:%d", uint(userID))
		cached, err := redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			// 缓存命中
			json.Unmarshal([]byte(cached), &user)
		} else {
			// 缓存未命中，查库
			if err := db.First(&user, uint(userID)).Error; err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
				return
			}
			// 写入缓存，TTL 5 分钟
			if data, err := json.Marshal(user); err == nil {
				redisClient.Set(ctx, cacheKey, string(data), 5*time.Minute)
			}
		}

		c.Set("user", user)
		c.Set("token", tokenString)
		c.Next()
	}
}

func RoleAuth(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		userModel, ok := user.(pkgdb.User)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user type"})
			return
		}

		if userModel.Role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}

func RevokeToken(redisClient *redis.Client, tokenString string, expireSeconds int64) error {
	return redisClient.Set(context.Background(), "blacklist:"+tokenString, "true", time.Duration(expireSeconds)*time.Second).Err()
}
