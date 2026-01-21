package main

import (
	"log"
	"os"

	"github.com/darron08/todolist-demo/internal/infrastructure/config"
	"github.com/darron08/todolist-demo/internal/interfaces/http"
	"github.com/gin-gonic/gin"
)

// @title Todo List API
// @version 1.0
// @description A high-performance todo list microservice
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := http.SetupRouter(cfg)

	// Get port from environment or config
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
