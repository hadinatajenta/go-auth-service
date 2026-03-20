package main

import (
	"auth-service/internal/config"
	"auth-service/internal/database"
	"auth-service/internal/module/menu"
	"auth-service/internal/module/permission"
	"auth-service/internal/module/role"
	"auth-service/internal/module/user"
	"auth-service/internal/utils"
	"log"
	"log/slog"

	"gorm.io/gorm"
)

func main() {
	// 1. Load config
	cfg := config.LoadConfig()

	// 2. Connect to database
	database.ConnectDB(cfg)
	db := database.DB

	// 3. Start transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		// Check if DB is empty logically by checking the users table
		var userCount int64
		if err := tx.Model(&user.User{}).Count(&userCount).Error; err != nil {
			return err
		}

		if userCount > 0 {
			slog.Info("Database already initialized — skipping bootstrap seed")
			return nil
		}

		slog.Info("Starting Bootstrap Seeding Process...")

		// PERMISSIONS
		slog.Info("Seeding System Admin permissions...")
		adminPerms := []string{
			"users.manage",
			"roles.manage",
			"permissions.manage",
			"menus.manage",
			"service_accounts.manage",
			"audit.view",
			"rbac.debug",
		}

		var createdPerms []permission.Permission
		for _, name := range adminPerms {
			p := permission.Permission{Name: name, Description: "Admin permission for " + name}
			if err := tx.FirstOrCreate(&p, permission.Permission{Name: name}).Error; err != nil {
				return err
			}
			createdPerms = append(createdPerms, p)
		}

		// ROLE
		slog.Info("Seeding System Admin role...")
		sysAdminRole := role.Role{Name: "System Admin", Description: "Superuser role with full access"}
		if err := tx.FirstOrCreate(&sysAdminRole, role.Role{Name: "System Admin"}).Error; err != nil {
			return err
		}

		// ROLE PERMISSIONS
		slog.Info("Assigning permissions to System Admin role...")
		for _, p := range createdPerms {
			rp := permission.RolePermission{
				RoleID:       sysAdminRole.ID,
				PermissionID: p.ID,
			}
			if err := tx.FirstOrCreate(&rp, permission.RolePermission{RoleID: sysAdminRole.ID, PermissionID: p.ID}).Error; err != nil {
				return err
			}
		}

		// USER
		slog.Info("Seeding minimal bootstrap identity...")
		password, err := utils.HashPassword("password123")
		if err != nil {
			return err
		}

		adminUser := user.User{
			Email:     "admin@system.local",
			Username:  "admin",
			Password:  password,
			FirstName: "System",
			LastName:  "Administrator",
		}
		
		if err := tx.FirstOrCreate(&adminUser, user.User{Email: "admin@system.local"}).Error; err != nil {
			return err
		}
		
		// Wait, FirstOrCreate doesn't update if found, but since userCount == 0, it will create.
		// However, to be extra safe let's ensure the user is inserted correctly.
		// `FirstOrCreate` will use the `Email` to find, and if not found, it creates with all fields in `adminUser`.

		// USER ROLE
		slog.Info("Assigning System Admin role to admin user...")
		ur := role.UserRole{
			UserID: adminUser.ID,
			RoleID: sysAdminRole.ID,
		}
		if err := tx.FirstOrCreate(&ur, role.UserRole{UserID: adminUser.ID, RoleID: sysAdminRole.ID}).Error; err != nil {
			return err
		}

		// MENUS
		slog.Info("Seeding System Admin menus...")

		// Dashboard Menu
		dashMenu := menu.Menu{Name: "Dashboard", Path: "/dashboard", SortOrder: 1}
		if err := tx.FirstOrCreate(&dashMenu, menu.Menu{Name: "Dashboard"}).Error; err != nil {
			return err
		}

		// Administration Menu
		adminMenu := menu.Menu{Name: "Administration", Path: "#", SortOrder: 2}
		if err := tx.FirstOrCreate(&adminMenu, menu.Menu{Name: "Administration"}).Error; err != nil {
			return err
		}

		adminChildren := []struct {
			Name       string
			Path       string
			Permission string
			SortOrder  int
		}{
			{"Users", "/users", "users.manage", 1},
			{"Roles", "/roles", "roles.manage", 2},
			{"Permissions", "/permissions", "permissions.manage", 3},
			{"Menus", "/menus", "menus.manage", 4},
			{"Service Accounts", "/service-accounts", "service_accounts.manage", 5},
			{"Audit Logs", "/audit-logs", "audit.view", 6},
		}

		for _, child := range adminChildren {
			m := menu.Menu{
				Name:      child.Name,
				Path:      child.Path,
				ParentID:  adminMenu.ID,
				SortOrder: child.SortOrder,
			}
			if err := tx.FirstOrCreate(&m, menu.Menu{Name: child.Name, ParentID: adminMenu.ID}).Error; err != nil {
				return err
			}

			// Map to permission
			var p permission.Permission
			if err := tx.Where("name = ?", child.Permission).First(&p).Error; err != nil {
				return err
			}

			mp := menu.MenuPermission{
				MenuID:       m.ID,
				PermissionID: p.ID,
			}
			if err := tx.FirstOrCreate(&mp, menu.MenuPermission{MenuID: m.ID, PermissionID: p.ID}).Error; err != nil {
				return err
			}
		}

		slog.Info("Bootstrap System Admin created")
		slog.Info("email: admin@system.local")
		slog.Info("password: password123")

		return nil
	})

	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}
}
