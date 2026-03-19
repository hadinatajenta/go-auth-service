package middleware

import (
	"auth-service/internal/utils/metrics"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsMiddleware records per-request latency and counts.
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// Only track API routes, not /metrics itself
		if c.Request.URL.Path != "/metrics" {
			metrics.AuthRequestTotal.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
				status,
			).Inc()

			metrics.AuthRequestDuration.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(duration)
		}

		// Track 401/403 as explicit auth failures
		if c.Writer.Status() == http.StatusUnauthorized || c.Writer.Status() == http.StatusForbidden {
			metrics.LoginFailureTotal.Inc()
		}
	}
}
