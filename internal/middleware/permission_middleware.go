package middleware

import (
	"auth-service/internal/module/user"
	"auth-service/internal/utils"
	"auth-service/internal/utils/cache"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PermissionMiddleware(userRepo user.Repository, csh cache.Cache, requiredPermission string) gin.HandlerFunc {
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

		ctx := c.Request.Context()
		cacheKey := fmt.Sprintf("user_perms:%d", uid)
		
		// Attempt O(1) check using Set cache
		hasPermission, err := csh.SIsMember(ctx, cacheKey, requiredPermission)
		if err == nil {
			if !hasPermission {
				utils.AbortWithError(c, http.StatusForbidden, "You do not have permission to access this resource", nil)
				return
			}
			c.Next()
			return
		}

		// Cache miss: Fetch from source of truth (Recursive CTE)
		permissions, err := userRepo.GetUserPermissions(ctx, uid)
		if err != nil {
			utils.AbortWithError(c, http.StatusInternalServerError, "Failed to retrieve user permissions", nil)
			return
		}
		
		// Re-populate Set cache with all permissions
		if len(permissions) > 0 {
			_ = csh.SAdd(ctx, cacheKey, permissions...)
		}
		
		// Final validation for current request
		authorized := false
		for _, p := range permissions {
			if p == requiredPermission {
				authorized = true
				break
			}
		}

		if !authorized {
			utils.AbortWithError(c, http.StatusForbidden, "You do not have permission to access this resource", nil)
			return
		}

		c.Next()
	}
}
