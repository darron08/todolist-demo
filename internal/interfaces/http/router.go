package http

import (
	"github.com/darron08/todolist-demo/internal/infrastructure/config"
	"github.com/darron08/todolist-demo/internal/infrastructure/redis"
	httpHandler "github.com/darron08/todolist-demo/internal/interfaces/http/handler"
	"github.com/darron08/todolist-demo/internal/interfaces/http/middleware"
	"github.com/darron08/todolist-demo/pkg/utils"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures Gin router with all routes and middleware
func SetupRouter(
	cfg *config.Config,
	jwtManager *utils.JWTManager,
	tokenStore *redis.TokenStore,
	userHandler *httpHandler.UserHandler,
	todoHandler *httpHandler.TodoHandler,
) *gin.Engine {
	r := gin.New()

	// Add global middleware
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.RequestValidation())

	// CORS middleware
	corsConfig := &middleware.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		ExposedHeaders:   cfg.CORS.ExposedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}
	r.Use(middleware.CORS(corsConfig))

	// Health check endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	r.GET("/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ready",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
			auth.POST("/logout", userHandler.Logout)
		}

		// User routes (require authentication)
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtManager))
		{
			users.GET("/profile", userHandler.GetProfile)
		}

		// Todo routes (require authentication)
		todos := v1.Group("/todos")
		todos.Use(middleware.AuthMiddleware(jwtManager))
		{
			todos.POST("", todoHandler.CreateTodo)
			todos.GET("", todoHandler.ListTodos)
			todos.GET("/:id", todoHandler.GetTodo)
			todos.PUT("/:id", todoHandler.UpdateTodo)
			todos.DELETE("/:id", todoHandler.DeleteTodo)
			todos.PATCH("/:id/status", todoHandler.UpdateTodoStatus)
		}

		// Tag routes (TODO: implement in phase 3)
		// tags := v1.Group("/tags")
		// tags.Use(middleware.AuthMiddleware(jwtManager))
		// {
		//     tags.GET("", tagHandler.ListTags)
		//     tags.POST("", tagHandler.CreateTag)
		//     tags.GET("/:id", tagHandler.GetTag)
		//     tags.PUT("/:id", tagHandler.UpdateTag)
		//     tags.DELETE("/:id", tagHandler.DeleteTag)
		// }
	}

	// Swagger documentation (if enabled)
	if cfg.Swagger.Enabled {
		r.Static(cfg.Swagger.Path, "./docs/swagger")
	}

	return r
}
