package router

import (
	"auth-service/internal/config"
	"auth-service/internal/middleware"
	"auth-service/internal/module/auth"
	"auth-service/internal/module/menu"
	"auth-service/internal/module/permission"
	"auth-service/internal/module/role"
	"auth-service/internal/module/user"
	"auth-service/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Initialize repositories
	userRepo := user.NewRepository(db)
	authRepo := auth.NewRepository(db)
	roleRepo := role.NewRepository(db)
	permRepo := permission.NewRepository(db)
	menuRepo := menu.NewRepository(db)

	// Initialize services
	authService := auth.NewService(userRepo, authRepo, cfg)
	userService := user.NewService(userRepo)
	roleService := role.NewService(roleRepo)
	permService := permission.NewService(permRepo)
	menuService := menu.NewService(menuRepo)

	// Initialize handlers
	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)
	roleHandler := role.NewHandler(roleService)
	permHandler := permission.NewHandler(permService)
	menuHandler := menu.NewHandler(menuService)

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORSMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		utils.SuccessResponse(c, "Server is healthy", gin.H{"status": "up"})
	})

	// API V1 Group
	v1 := r.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/register", authHandler.Register)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			protected.GET("/me", userHandler.GetProfile)

			// User Routes
			userGroup := protected.Group("/users")
			{
				userGroup.GET("", userHandler.List)
				userGroup.GET("/:id", userHandler.GetProfile)
				userGroup.PUT("/:id", userHandler.Update)
				userGroup.DELETE("/:id", userHandler.Delete)
			}

			// Role Routes
			roleGroup := protected.Group("/roles")
			{
				roleGroup.POST("", roleHandler.Create)
				roleGroup.GET("", roleHandler.List)
				roleGroup.GET("/:id", roleHandler.GetByID)
				roleGroup.PUT("/:id", roleHandler.Update)
				roleGroup.DELETE("/:id", roleHandler.Delete)
			}

			// Permission Routes
			permGroup := protected.Group("/permissions")
			{
				permGroup.POST("", permHandler.Create)
				permGroup.GET("", permHandler.List)
				permGroup.GET("/:id", permHandler.GetByID)
				permGroup.PUT("/:id", permHandler.Update)
				permGroup.DELETE("/:id", permHandler.Delete)
			}

			// Menu Routes
			menuGroup := protected.Group("/menus")
			{
				menuGroup.POST("", menuHandler.Create)
				menuGroup.GET("", menuHandler.List)
				menuGroup.GET("/:id", menuHandler.GetByID)
				menuGroup.PUT("/:id", menuHandler.Update)
				menuGroup.DELETE("/:id", menuHandler.Delete)
			}
		}
	}

	return r
}
