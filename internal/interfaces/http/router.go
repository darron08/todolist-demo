package http

import (
	"github.com/darron08/todolist-demo/internal/infrastructure/config"
	httpHandler "github.com/darron08/todolist-demo/internal/interfaces/http/handler"
	"github.com/darron08/todolist-demo/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures Gin router with all routes and middleware
func SetupRouter(cfg *config.Config) *gin.Engine {
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
		// Todo routes (without authentication for MVP)
		// TODO: Add authentication middleware in phase 3
		todoHandler := httpHandler.NewTodoHandler(nil)

		todos := v1.Group("/todos")
		{
			todos.POST("", todoHandler.CreateTodo)                   // Create todo
			todos.GET("", todoHandler.ListTodos)                     // List todos
			todos.GET("/:id", todoHandler.GetTodo)                   // Get todo by ID
			todos.PUT("/:id", todoHandler.UpdateTodo)                // Update todo
			todos.DELETE("/:id", todoHandler.DeleteTodo)             // Delete todo
			todos.PATCH("/:id/status", todoHandler.UpdateTodoStatus) // Update todo status
		}

		// User routes (TODO: implement in phase 3)
		// users := v1.Group("/users")
		// {
		//     users.POST("/register", userHandler.Register)
		//     users.POST("/login", userHandler.Login)
		// }

		// Authentication routes (TODO: implement in phase 3)
		// auth := v1.Group("/auth")
		// {
		//     auth.POST("/register", authHandler.Register)
		//     auth.POST("/login", authHandler.Login)
		//     auth.POST("/refresh", authHandler.RefreshToken)
		// }

		// Tag routes (TODO: implement in phase 3)
		// tags := v1.Group("/tags")
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
