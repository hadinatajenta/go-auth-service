package main

import (
	"auth-service/internal/config"
	"auth-service/internal/database"
	"auth-service/internal/router"
	"auth-service/pkg/logger"
	"log/slog"
)

func main() {
	// Initialize logger
	logger.Init()

	// Load config
	cfg := config.LoadConfig()

	// Connect to database
	database.ConnectDB(cfg)

	// Setup router
	r := router.SetupRouter(database.DB, cfg)

	slog.Info("Server starting", "port", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		slog.Error("Failed to run server", "error", err)
	}
}
