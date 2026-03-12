package middleware

import (
	"auth-service/internal/config"
	"auth-service/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			if userID, exists := claims["user_id"]; exists {
				c.Set("user_id", uint(userID.(float64)))
			}
		}

		c.Next()
	}
}
