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

		// Phase 1: Optimized Indices for RBAC & Auditing
		db.Exec("CREATE INDEX IF NOT EXISTS idx_user_roles_composite ON user_roles(user_id, role_id)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_role_permissions_composite ON role_permissions(role_id, permission_id)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_lookup ON audit_logs(entity, entity_id, created_at DESC)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_user_activity ON audit_logs(user_id, created_at DESC)")

		// Phase 2: Ensure Cascading Deletes for Foreign Keys
		// This fixes the issue where deleting a permission/role/user/menu fails due to foreign key constraints.
		
		// Role Permissions
		db.Exec(`ALTER TABLE "role_permissions" DROP CONSTRAINT IF EXISTS "role_permissions_role_id_fkey"`)
		db.Exec(`ALTER TABLE "role_permissions" ADD CONSTRAINT "role_permissions_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON DELETE CASCADE`)
		db.Exec(`ALTER TABLE "role_permissions" DROP CONSTRAINT IF EXISTS "role_permissions_permission_id_fkey"`)
		db.Exec(`ALTER TABLE "role_permissions" ADD CONSTRAINT "role_permissions_permission_id_fkey" FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id") ON DELETE CASCADE`)

		// Menu Permissions
		db.Exec(`ALTER TABLE "menu_permissions" DROP CONSTRAINT IF EXISTS "menu_permissions_menu_id_fkey"`)
		db.Exec(`ALTER TABLE "menu_permissions" ADD CONSTRAINT "menu_permissions_menu_id_fkey" FOREIGN KEY ("menu_id") REFERENCES "menus" ("id") ON DELETE CASCADE`)
		db.Exec(`ALTER TABLE "menu_permissions" DROP CONSTRAINT IF EXISTS "menu_permissions_permission_id_fkey"`)
		db.Exec(`ALTER TABLE "menu_permissions" ADD CONSTRAINT "menu_permissions_permission_id_fkey" FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id") ON DELETE CASCADE`)

		// User Roles
		db.Exec(`ALTER TABLE "user_roles" DROP CONSTRAINT IF EXISTS "user_roles_user_id_fkey"`)
		db.Exec(`ALTER TABLE "user_roles" ADD CONSTRAINT "user_roles_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE`)
		db.Exec(`ALTER TABLE "user_roles" DROP CONSTRAINT IF EXISTS "user_roles_role_id_fkey"`)
		db.Exec(`ALTER TABLE "user_roles" ADD CONSTRAINT "user_roles_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON DELETE CASCADE`)

		// User Sessions
		db.Exec(`ALTER TABLE "user_sessions" DROP CONSTRAINT IF EXISTS "user_sessions_user_id_fkey"`)
		db.Exec(`ALTER TABLE "user_sessions" ADD CONSTRAINT "user_sessions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE`)
	}
}
