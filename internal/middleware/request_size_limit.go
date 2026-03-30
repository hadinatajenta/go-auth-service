package middleware

import (
	"auth-service/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestSizeLimitMiddleware prevents DoS attacks by limiting request body size
// CRITICAL SECURITY: Prevents memory exhaustion from large payloads
func RequestSizeLimitMiddleware() gin.HandlerFunc {
	const maxBodySize = 10 * 1024 * 1024 // 10 MB max

	return func(c *gin.Context) {
		// Set max body size for request
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodySize)

		// If body is too large, return 413 Payload Too Large
		if err := c.Request.ParseForm(); err != nil {
			if err.Error() == "http: request body too large" {
				utils.AbortWithError(c, http.StatusRequestEntityTooLarge, "Request body too large (max 10 MB)", nil)
				return
			}
		}

		c.Next()
	}
}
