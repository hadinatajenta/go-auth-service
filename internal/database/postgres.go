package database

import (
	"auth-service/internal/config"
	"auth-service/internal/module/auth"
	"auth-service/internal/module/audit"
	"auth-service/internal/module/menu"
	"auth-service/internal/module/permission"
	"auth-service/internal/module/role"
	"auth-service/internal/module/user"
	"fmt"
	"log/slog"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB(cfg *config.Config) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	DB = db
	slog.Info("Database connection established")

	// Automated Migrations
	err = db.AutoMigrate(
		&user.User{},
		&auth.UserSession{},
		&role.Role{},
		&role.UserRole{},
		&permission.Permission{},
		&permission.RolePermission{},
		&menu.Menu{},
		&menu.MenuPermission{},
		&audit.AuditLog{},
	)
	if err != nil {
		slog.Error("Database migration failed", "error", err)
	} else {
		slog.Info("Database migration completed successfully")
		
		// Apply Partial Unique Indexes for Soft Delete support
		// This ensures unique constraints only apply to non-deleted records
		db.Exec("DROP INDEX IF EXISTS idx_users_username_active")
		db.Exec("CREATE UNIQUE INDEX idx_users_username_active ON users(username) WHERE deleted_at IS NULL")
		
		db.Exec("DROP INDEX IF EXISTS idx_users_email_active")
		db.Exec("CREATE UNIQUE INDEX idx_users_email_active ON users(email) WHERE deleted_at IS NULL")
		
		db.Exec("DROP INDEX IF EXISTS idx_roles_name_active")
		db.Exec("CREATE UNIQUE INDEX idx_roles_name_active ON roles(name) WHERE deleted_at IS NULL")
	}
}
