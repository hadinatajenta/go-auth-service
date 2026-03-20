package main

import (
	"auth-service/internal/config"
	"auth-service/internal/database"
	"auth-service/internal/module/menu"
	"auth-service/internal/module/permission"
	"auth-service/internal/module/role"
	"auth-service/internal/module/user"
	"fmt"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	database.ConnectDB(cfg)
	db := database.DB

	var u user.User
	if err := db.First(&u, 1).Error; err != nil {
		log.Fatalf("No user 1: %v", err)
	}
	fmt.Printf("User 1: %s\n", u.Email)

	var roles []role.Role
	db.Raw(`SELECT r.* FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id = 1`).Scan(&roles)
	for _, r := range roles {
		fmt.Printf("Role: %s\n", r.Name)
	}

	var perms []permission.Permission
	db.Raw(`SELECT p.* FROM permissions p JOIN role_permissions rp ON p.id = rp.permission_id JOIN user_roles ur ON rp.role_id = ur.role_id WHERE ur.user_id = 1`).Scan(&perms)
	fmt.Printf("Permissions count: %d\n", len(perms))
	for _, p := range perms {
		fmt.Printf("  - %s\n", p.Name)
	}

	var menus []menu.Menu
	db.Raw(`
		SELECT DISTINCT m.*
		FROM menus m
		LEFT JOIN menu_permissions mp ON m.id = mp.menu_id
		LEFT JOIN role_permissions rp ON mp.permission_id = rp.permission_id
		LEFT JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = 1 OR mp.permission_id IS NULL
	`).Scan(&menus)
	
	fmt.Printf("Menus found using SQL query:\n")
	for _, m := range menus {
		fmt.Printf("  - %s (path: %s)\n", m.Name, m.Path)
	}

	fmt.Printf("All MENUS:\n")
	var allMenus []menu.Menu
	db.Find(&allMenus)
	for _, m := range allMenus {
		var mp menu.MenuPermission
		db.Where("menu_id = ?", m.ID).First(&mp)
		fmt.Printf("  - %s (path: %s, perm_id: %v)\n", m.Name, m.Path, mp.PermissionID)
	}
}
