package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBPort              string
	DBSSLMode           string
	JWTSecret           string
	AppPort             string
	CORSAllowedOrigins  string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		slog.Info("No .env file found, using system environment variables")
	}

	// CRITICAL: JWT_SECRET must be set and strong
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("FATAL: JWT_SECRET environment variable is required and not set")
		os.Exit(1)
	}
	
	if len(jwtSecret) < 32 {
		slog.Error("FATAL: JWT_SECRET must be at least 32 characters long", "length", len(jwtSecret))
		os.Exit(1)
	}

	return &Config{
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBUser:              getEnv("DB_USER", "postgres"),
		DBPassword:          getEnv("DB_PASSWORD", "postgres"),
		DBName:              getEnv("DB_NAME", "auth_db"),
		DBPort:              getEnv("DB_PORT", "5432"),
		DBSSLMode:           getEnv("DB_SSLMODE", "disable"),
		JWTSecret:           jwtSecret,
		AppPort:             getEnv("APP_PORT", "8080"),
		CORSAllowedOrigins:  getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
