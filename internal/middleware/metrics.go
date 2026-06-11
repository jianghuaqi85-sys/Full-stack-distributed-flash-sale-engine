package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsData holds metrics information for a request
type MetricsData struct {
	Method     string
	Path       string
	StatusCode int
	Duration   time.Duration
}

// MetricsCollector is the interface for collecting metrics
type MetricsCollector interface {
	RecordHTTPRequest(method, path string, statusCode int, duration time.Duration)
}

// MetricsMiddleware creates a middleware that records HTTP request metrics
func MetricsMiddleware(collector MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.FullPath() // Use route pattern, not actual path

		if path == "" {
			path = "unknown"
		}

		collector.RecordHTTPRequest(method, path, statusCode, duration)
	}
}

// PrometheusMetricsCollector implements MetricsCollector using Prometheus-style counters
// In production, use prometheus/client_golang. This is a lightweight in-memory collector.
type PrometheusMetricsCollector struct {
	requestCount    map[string]int64
	requestDuration map[string]time.Duration
}

// NewPrometheusMetricsCollector creates a new in-memory metrics collector
func NewPrometheusMetricsCollector() *PrometheusMetricsCollector {
	return &PrometheusMetricsCollector{
		requestCount:    make(map[string]int64),
		requestDuration: make(map[string]time.Duration),
	}
}

func (c *PrometheusMetricsCollector) RecordHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	key := method + ":" + path + ":" + strconv.Itoa(statusCode)
	c.requestCount[key]++
	c.requestDuration[key] += duration
}

// GetMetrics returns current metrics snapshot
func (c *PrometheusMetricsCollector) GetMetrics() (map[string]int64, map[string]time.Duration) {
	return c.requestCount, c.requestDuration
}
