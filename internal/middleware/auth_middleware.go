package middleware

import (
	"auth-service/internal/config"
	sa "auth-service/internal/module/service_account"
	"auth-service/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware supports two authentication schemes:
//   - Bearer <jwt>   → sets actor_type="user",    actor_id=user_id
//   - ApiKey <key>   → sets actor_type="service",  actor_id=service_account_id
func AuthMiddleware(cfg *config.Config, saService sa.Service) gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.AbortWithError(c, http.StatusUnauthorized, utils.MsgTokenRequired, nil)
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 {
			utils.AbortWithError(c, http.StatusUnauthorized, utils.MsgInvalidAuthFormat, nil)
			return
		}

		scheme := strings.ToLower(parts[0])
		credential := parts[1]

		switch scheme {
		case "bearer":
			handleJWT(c, cfg, credential)
		case "apikey":
			handleAPIKey(c, saService, credential)
		default:
			utils.AbortWithError(c, http.StatusUnauthorized, utils.MsgInvalidAuthFormat, nil)
		}
	}
}

func handleJWT(c *gin.Context, cfg *config.Config, tokenString string) {
	token, err := utils.ValidateToken(tokenString, cfg.JWTSecret)
	if err != nil || !token.Valid {
		utils.AbortWithError(c, http.StatusUnauthorized, utils.MsgTokenInvalid, nil)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.AbortWithError(c, http.StatusUnauthorized, utils.MsgTokenInvalid, nil)
		return
	}

	userIDRaw, exists := claims["user_id"]
	if !exists {
		utils.AbortWithError(c, http.StatusUnauthorized, utils.MsgTokenInvalid, nil)
		return
	}

	userIDFloat, ok := userIDRaw.(float64)
	if !ok {
		utils.AbortWithError(c, http.StatusUnauthorized, utils.MsgTokenInvalid, nil)
		return
	}

	userID := uint(userIDFloat)
	c.Set("user_id", userID) // kept for backward compatibility
	c.Set("actor_type", "user")
	c.Set("actor_id", userID)
	c.Next()
}

func handleAPIKey(c *gin.Context, saService sa.Service, rawKey string) {
	account, err := saService.AuthenticateByKey(c.Request.Context(), rawKey)
	if err != nil {
		utils.AbortWithError(c, http.StatusUnauthorized, "Invalid or revoked API key", nil)
		return
	}

	c.Set("actor_type", "service")
	c.Set("actor_id", account.ID)
	c.Set("service_account_id", account.ID)
	c.Next()
}
