package middleware

import (
	"auth-service/internal/config"
	"auth-service/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.AbortWithError(c, http.StatusUnauthorized, utils.MsgTokenRequired, nil)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.AbortWithError(c, http.StatusUnauthorized, utils.MsgInvalidAuthFormat, nil)
			return
		}

		tokenString := parts[1]
		token, err := utils.ValidateToken(tokenString, cfg.JWTSecret)
		if err != nil || !token.Valid {
			utils.AbortWithError(c, http.StatusUnauthorized, utils.MsgTokenInvalid, nil)
			return
		}

		c.Next()
	}
}
