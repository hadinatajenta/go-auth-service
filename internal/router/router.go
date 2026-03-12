package router

import (
	"auth-service/internal/config"
	"auth-service/internal/middleware"
	"auth-service/internal/module/auth"
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

	// Initialize services
	authService := auth.NewService(userRepo, authRepo, cfg)
	userService := user.NewService(userRepo)

	// Initialize handlers
	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

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
			protected.GET("/me", userHandler.GetProfile) // Example protected route
		}
	}

	return r
}
