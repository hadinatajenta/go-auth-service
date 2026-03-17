package middleware

import (
	"auth-service/internal/module/user"
	"auth-service/internal/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PermissionMiddleware(userRepo user.Repository, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.AbortWithError(c, http.StatusUnauthorized, utils.MsgTokenRequired, nil)
			return
		}

		uid, ok := userID.(uint)
		if !ok {
			utils.AbortWithError(c, http.StatusUnauthorized, utils.MsgTokenInvalid, nil)
			return
		}

		permissions, err := userRepo.GetUserPermissions(c.Request.Context(), uid)
		fmt.Println("permissions : ", permissions)
		if err != nil {
			utils.AbortWithError(c, http.StatusInternalServerError, "Failed to retrieve user permissions", nil)
			return
		}

		hasPermission := false
		for _, p := range permissions {
			if p == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			utils.AbortWithError(c, http.StatusForbidden, "You do not have permission to access this resource", nil)
			return
		}

		c.Next()
	}
}
