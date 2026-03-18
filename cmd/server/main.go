package main

import (
	"auth-service/internal/config"
	"auth-service/internal/database"
	"auth-service/internal/router"
	"auth-service/pkg/logger"
	"log/slog"
)

// @title Auth Service API
// @version 1.0
// @description Advanced RBAC and Authentication Service with Audit Logs.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /api/v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
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
