package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"order-system/internal/pkg/db"
)

// getUser 从 gin.Context 提取并验证用户信息
func getUser(c *gin.Context) (db.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return db.User{}, false
	}

	userModel, ok := user.(db.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户类型错误"})
		return db.User{}, false
	}

	return userModel, true
}

// parsePageLimit 解析分页参数，提供默认值和边界检查
func parsePageLimit(c *gin.Context, defaultPage, defaultLimit, maxLimit int) (int, int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultPage)))
	if err != nil || page < 1 {
		page = defaultPage
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return page, limit
}
