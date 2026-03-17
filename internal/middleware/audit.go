package middleware

import (
	"auth-service/internal/module/audit"

	"github.com/gin-gonic/gin"
)

func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use RequestID if available from previous headers or generate one
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = c.GetString("RequestID")
		}

		val, _ := c.Get("user_id")
		userID, _ := val.(uint)

		auditCtx := &audit.AuditContext{
			UserID:    userID,
			RequestID: requestID,
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}

		// Inject into Go context
		ctx := audit.NewContext(c.Request.Context(), auditCtx)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
