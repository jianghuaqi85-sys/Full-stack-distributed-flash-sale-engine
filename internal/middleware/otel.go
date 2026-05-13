package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func OpenTelemetry() gin.HandlerFunc {
	return func(c *gin.Context) {
		tracer := otel.Tracer("gin-tracer")
		ctx, span := tracer.Start(c.Request.Context(),
			fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path))
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.Request.URL.Path),
			attribute.String("http.host", c.Request.Host),
			attribute.String("http.remote_addr", c.ClientIP()),
		)

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))
	}
}

func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tracer := otel.Tracer("gin-middleware")
		ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+" "+c.Request.URL.Path)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				tracer := otel.Tracer("gin-recovery")
				_, span := tracer.Start(c.Request.Context(), "panic-recovery")
				span.RecordError(fmt.Errorf("%v", r))
				span.End()

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}

