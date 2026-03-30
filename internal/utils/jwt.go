package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID uint, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Minute * 15).Unix(), // Access token short-lived
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID uint, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // Refresh token 7 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ValidateToken validates JWT with strict algorithm enforcement
// CRITICAL SECURITY: Only accepts HS256 algorithm
func ValidateToken(tokenString, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// ✅ ENFORCE: Only allow HMAC algorithms
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// ✅ ENFORCE: Only allow HS256 specifically
		if token.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("invalid algorithm: expected HS256, got %s", token.Method.Alg())
		}

		return []byte(secret), nil
	})
}
