package router

import (
	"auth-service/internal/config"
	"auth-service/internal/middleware"
	"auth-service/internal/module/audit"
	"auth-service/internal/module/auth"
	"auth-service/internal/module/menu"
	"auth-service/internal/module/permission"
	"auth-service/internal/module/role"
	sa "auth-service/internal/module/service_account"
	"auth-service/internal/module/user"
	"auth-service/internal/utils"
	"auth-service/internal/utils/cache"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"

	_ "auth-service/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Initialize utilities
	memoryCache := cache.NewMemoryCache()

	// Initialize repositories
	userRepo := user.NewRepository(db)
	authRepo := auth.NewRepository(db)
	roleRepo := role.NewRepository(db)
	permRepo := permission.NewRepository(db)
	menuRepo := menu.NewRepository(db)
	auditRepo := audit.NewRepository(db)
	saRepo := sa.NewRepository(db)

	// Initialize services
	auditService := audit.NewService(auditRepo)
	authService := auth.NewService(userRepo, authRepo, cfg)
	userService := user.NewService(userRepo, auditService, memoryCache)
	roleService := role.NewService(roleRepo, auditService, memoryCache)
	permService := permission.NewService(permRepo, memoryCache)
	menuService := menu.NewService(menuRepo)
	saService := sa.NewService(saRepo)

	// Initialize handlers
	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)
	roleHandler := role.NewHandler(roleService)
	permHandler := permission.NewHandler(permService)
	menuHandler := menu.NewHandler(menuService)
	auditHandler := audit.NewHandler(auditService)
	saHandler := sa.NewHandler(saService)

	// Global middleware (in order)
	r.Use(gin.Recovery())                          // Panic recovery
	r.Use(gin.Logger())                            // Request logging
	r.Use(middleware.CORSMiddleware(cfg))          // CORS with whitelist
	r.Use(middleware.AuditMiddleware())            // Audit logging
	r.Use(middleware.MetricsMiddleware())          // Prometheus metrics
	r.Use(middleware.RequestSizeLimitMiddleware()) // DoS protection: limit request body size

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Database connection error", err.Error())
			return
		}

		if err := sqlDB.Ping(); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Database ping failed", err.Error())
			return
		}

		utils.SuccessResponse(c, "Server is healthy", gin.H{
			"status":   "up",
			"database": "connected",
		})
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API V1 Group
	v1 := r.Group("/api/v1")
	{
		authorizedAuth := v1.Group("/auth")
		authorizedAuth.Use(middleware.RateLimitMiddleware(5.0/60.0, 10))
		{
			authorizedAuth.POST("/login", authHandler.Login)
			authorizedAuth.POST("/register", authHandler.Register)
			authorizedAuth.POST("/refresh", authHandler.Refresh)
			authorizedAuth.POST("/logout", authHandler.Logout)
			authorizedAuth.POST("/forgot-password", userHandler.ForgotPassword)
			authorizedAuth.POST("/reset-password", userHandler.ResetPassword)
			authorizedAuth.POST("/introspect", authHandler.Introspect)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg, saService))
		{
			protected.GET("/me", userHandler.GetProfile)
			protected.POST("/change-password", userHandler.ChangePassword)
			protected.POST("/auth/logout-all", authHandler.LogoutAll)
			protected.GET("/me/permissions", userHandler.GetMyPermissions)
			protected.GET("/me/roles", userHandler.GetMyRoles)

			// Menu endpoints accessible to ALL authenticated users (user-facing navigation)
			protected.GET("/menus/allowed", menuHandler.GetAllowed)
			protected.GET("/menus/tree", menuHandler.GetTree)

			// User Routes
			userGroup := protected.Group("/users")
			userGroup.Use(middleware.PermissionMiddleware(userRepo, memoryCache, "users.manage"))
			{
				userGroup.POST("", userHandler.Create)
				userGroup.GET("", userHandler.List)
				userGroup.GET("/:id", userHandler.GetProfile)
				userGroup.PUT("/:id", userHandler.Update)
				userGroup.DELETE("/:id", userHandler.Delete)
				userGroup.POST("/:id/roles", userHandler.AddRole)
				userGroup.DELETE("/:id/roles/:roleId", userHandler.RemoveRole)
				userGroup.GET("/:id/roles", userHandler.ListRoles)
				userGroup.GET("/:id/permissions", userHandler.GetPermissions)
			}

			// Role Routes
			roleGroup := protected.Group("/roles")
			roleGroup.Use(middleware.PermissionMiddleware(userRepo, memoryCache, "roles.manage"))
			{
				roleGroup.POST("", roleHandler.Create)
				roleGroup.GET("", roleHandler.List)
				roleGroup.GET("/:id", roleHandler.GetByID)
				roleGroup.PUT("/:id", roleHandler.Update)
				roleGroup.DELETE("/:id", roleHandler.Delete)
				roleGroup.POST("/:id/permissions", roleHandler.AddPermission)
				roleGroup.DELETE("/:id/permissions/:permissionId", roleHandler.RemovePermission)
				roleGroup.GET("/:id/permissions", roleHandler.ListPermissions)
				roleGroup.GET("/:id/users", roleHandler.ListUsers)
			}

			// Permission Routes
			permGroup := protected.Group("/permissions")
			{
				// GET /grouped is accessible to anyone who can manage permissions OR roles
				permGroup.GET("/grouped", middleware.PermissionAnyMiddleware(userRepo, memoryCache, "permissions.manage", "roles.manage"), permHandler.GetGrouped)

				// Other endpoints strictly require permissions.manage
				permGroup.POST("", middleware.PermissionMiddleware(userRepo, memoryCache, "permissions.manage"), permHandler.Create)
				permGroup.GET("", middleware.PermissionMiddleware(userRepo, memoryCache, "permissions.manage"), permHandler.List)
				permGroup.GET("/:id", middleware.PermissionMiddleware(userRepo, memoryCache, "permissions.manage"), permHandler.GetByID)
				permGroup.PUT("/:id", middleware.PermissionMiddleware(userRepo, memoryCache, "permissions.manage"), permHandler.Update)
				permGroup.DELETE("/:id", middleware.PermissionMiddleware(userRepo, memoryCache, "permissions.manage"), permHandler.Delete)
			}

			// Menu Admin Routes (menus.manage permission required)
			menuGroup := protected.Group("/menus")
			menuGroup.Use(middleware.PermissionMiddleware(userRepo, memoryCache, "menus.manage"))
			{
				menuGroup.POST("", menuHandler.Create)
				menuGroup.GET("/:id", menuHandler.GetByID)
				menuGroup.PUT("/:id", menuHandler.Update)
				menuGroup.DELETE("/:id", menuHandler.Delete)
			}
			// Audit Logs
			auditGroup := protected.Group("/audit-logs")
			auditGroup.Use(middleware.PermissionMiddleware(userRepo, memoryCache, "audit.view"))
			{
				auditGroup.GET("", auditHandler.List)
			}

			// RBAC Debug
			rbacGroup := protected.Group("/rbac")
			rbacGroup.Use(middleware.PermissionMiddleware(userRepo, memoryCache, "rbac.debug"))
			{
				rbacGroup.GET("/debug/user/:id", roleHandler.DebugUser)
			}

			// Service Accounts
			saGroup := protected.Group("/service-accounts")
			saGroup.Use(middleware.PermissionMiddleware(userRepo, memoryCache, "service_accounts.manage"))
			{
				saGroup.POST("", saHandler.Create)
				saGroup.GET("", saHandler.List)
				saGroup.POST("/:id/revoke", saHandler.Revoke)
			}
		}
	}

	return r
}
