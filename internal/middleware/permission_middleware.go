package middleware

import (
	"auth-service/internal/module/user"
	"auth-service/internal/utils"
	"auth-service/internal/utils/cache"
	"fmt"
	"net/http"
	"strings"
	"time"

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
		
		var permissions []string
		cachedData, err := csh.Get(ctx, cacheKey)
		
		if err == nil {
			permissions = strings.Split(cachedData, ",")
		} else {
			permissions, err = userRepo.GetUserPermissions(ctx, uid)
			if err != nil {
				utils.AbortWithError(c, http.StatusInternalServerError, "Failed to retrieve user permissions", nil)
				return
			}
			// Save to cache for 1 hour (event-based invalidation will handle updates)
			_ = csh.Set(ctx, cacheKey, strings.Join(permissions, ","), 1*time.Hour)
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
