package main

import (
	"log"
	"os"
	"time"

	"github.com/darron08/todolist-demo/internal/infrastructure/config"
	"github.com/darron08/todolist-demo/internal/infrastructure/database"
	"github.com/darron08/todolist-demo/internal/infrastructure/redis"
	"github.com/darron08/todolist-demo/internal/infrastructure/repository"
	"github.com/darron08/todolist-demo/internal/interfaces/http"
	httpHandler "github.com/darron08/todolist-demo/internal/interfaces/http/handler"
	"github.com/darron08/todolist-demo/internal/usecase"
	"github.com/darron08/todolist-demo/pkg/utils"
	"github.com/gin-gonic/gin"
)

// @title Todo List API
// @version 1.0
// @description A high-performance todo list microservice built with Go and Clean Architecture principles.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
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

	// Initialize databases
	databases, err := database.InitializeDatabases(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize databases: %v", err)
	}
	defer closeDatabases(databases)

	// Initialize repositories
	userRepo := repository.NewUserRepository(databases.MySQL.GetDB())
	todoRepo := repository.NewTodoRepository(databases.MySQL.GetDB())
	tagRepo := repository.NewTagRepository(databases.MySQL.GetDB())
	todoTagRepo := repository.NewTodoTagRepository(databases.MySQL.GetDB())

	// Initialize token store
	tokenStore := redis.NewTokenStore(databases.Redis)

	// Initialize JWT manager
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	jwtManager := utils.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Issuer, accessTokenExpiry, refreshTokenExpiry)

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(userRepo, jwtManager, tokenStore)
	todoUseCase := usecase.NewTodoUseCase(todoRepo, tagRepo, todoTagRepo)
	adminUseCase := usecase.NewAdminUseCase(userRepo, todoRepo)
	tagUseCase := usecase.NewTagUseCase(tagRepo, todoTagRepo)

	// Initialize handlers
	userHandler := httpHandler.NewUserHandler(userUseCase)
	todoHandler := httpHandler.NewTodoHandler(todoUseCase)
	adminHandler := httpHandler.NewAdminHandler(adminUseCase)
	tagHandler := httpHandler.NewTagHandler(tagUseCase)

	// Initialize router
	router := http.SetupRouter(cfg, jwtManager, tokenStore, userHandler, todoHandler, adminHandler, tagHandler)

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

// closeDatabases closes all database connections
func closeDatabases(dbs *database.Database) {
	if dbs.MySQL != nil {
		if err := dbs.MySQL.Close(); err != nil {
			log.Printf("Failed to close MySQL connection: %v", err)
		}
	}
	if dbs.Redis != nil {
		if err := dbs.Redis.Close(); err != nil {
			log.Printf("Failed to close Redis connection: %v", err)
		}
	}
}
