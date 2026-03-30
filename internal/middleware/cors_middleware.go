package middleware

import (
	"auth-service/internal/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware enforces CORS with whitelist
// Origins must be explicitly configured via CORS_ALLOWED_ORIGINS env var
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	allowedOrigins := parseAllowedOrigins(cfg.CORSAllowedOrigins)

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Only accept whitelisted origins
		if origin != "" && isOriginAllowed(origin, allowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		} else if origin != "" {
			// Reject non-whitelisted origin
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}

func parseAllowedOrigins(originsStr string) []string {
	if originsStr == "" {
		return []string{}
	}
	origins := strings.Split(originsStr, ",")
	for i, o := range origins {
		origins[i] = strings.TrimSpace(o)
	}
	return origins
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}
