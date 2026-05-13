package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "order-system/internal/pkg/errors"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.Code, gin.H{
				"error": appErr.Message,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}
