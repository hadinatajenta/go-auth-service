package middleware

import (
	"github.com/gin-gonic/gin"
)

func PermissionMiddleware(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Placeholder for permission check logic
		// In a real app, you would get user roles and permissions from DB/Context
		c.Next()
	}
}
