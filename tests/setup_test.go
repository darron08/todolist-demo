package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"github.com/darron08/todolist-demo/internal/infrastructure/database/mysql"
	"github.com/darron08/todolist-demo/internal/infrastructure/database/redis"
	"github.com/darron08/todolist-demo/internal/infrastructure/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/logger"
)

var (
	TestDB    *mysql.DB
	TestRedis *redis.Client
)

// TestConfig holds test configuration
type TestConfig struct {
	MySQLHost     string
	MySQLPort     string
	MySQLUser     string
	MySQLPassword string
	MySQLDatabase string
	RedisHost     string
	RedisPort     string
}

// SetupTest initializes test databases and repositories
func SetupTest(t *testing.T) (*TestConfig, error) {
	// Load test configuration or use defaults
	cfg := &TestConfig{
		MySQLHost:     getEnv("TEST_MYSQL_HOST", "localhost"),
		MySQLPort:     getEnv("TEST_MYSQL_PORT", "3307"),
		MySQLUser:     getEnv("TEST_MYSQL_USER", "todolist_test_user"),
		MySQLPassword: getEnv("TEST_MYSQL_PASSWORD", "todolist_test_password"),
		MySQLDatabase: getEnv("TEST_MYSQL_DATABASE", "todolist_test"),
		RedisHost:     getEnv("TEST_REDIS_HOST", "localhost"),
		RedisPort:     getEnv("TEST_REDIS_PORT", "6380"),
	}

	// Initialize MySQL test database
	if err := setupMySQLTestDB(cfg); err != nil {
		return nil, fmt.Errorf("failed to setup MySQL: %w", err)
	}

	// Initialize Redis test connection
	if err := setupRedisTestConnection(cfg); err != nil {
		return nil, fmt.Errorf("failed to setup Redis: %w", err)
	}

	// Auto migrate tables
	if err := autoMigrateTestDB(); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	return cfg, nil
}

// TeardownTest cleans up test databases
func TeardownTest(t *testing.T) {
	// Close MySQL connection
	if TestDB != nil {
		if err := TestDB.Close(); err != nil {
			t.Logf("Failed to close MySQL: %v", err)
		}
		TestDB = nil
	}

	// Close Redis connection
	if TestRedis != nil {
		if err := TestRedis.Close(); err != nil {
			t.Logf("Failed to close Redis: %v", err)
		}
		TestRedis = nil
	}
}

// setupMySQLTestDB initializes MySQL test database
func setupMySQLTestDB(cfg *TestConfig) error {
	mysqlConfig := &mysql.DBConfig{
		Host:            cfg.MySQLHost,
		Port:            cfg.MySQLPort,
		Username:        cfg.MySQLUser,
		Password:        cfg.MySQLPassword,
		Database:        cfg.MySQLDatabase,
		Charset:         "utf8mb4",
		ParseTime:       true,
		Loc:             "Local",
		MaxIdleConns:    5,
		MaxOpenConns:    20,
		ConnMaxLifetime: time.Hour,
		LogMode:         logger.Silent,
	}

	var err error
	TestDB, err = mysql.NewConnection(mysqlConfig)
	if err != nil {
		return err
	}

	// Clean database before tests
	if err := cleanTestDatabase(); err != nil {
		return err
	}

	return nil
}

// setupRedisTestConnection initializes Redis test connection
func setupRedisTestConnection(cfg *TestConfig) error {
	dialTimeout := 5 * time.Second
	readTimeout := 3 * time.Second
	writeTimeout := 3 * time.Second

	redisConfig := &redis.Config{
		Host:         cfg.RedisHost,
		Port:         cfg.RedisPort,
		Password:     "",
		Database:     1,
		PoolSize:     5,
		MinIdleConns: 2,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	var err error
	TestRedis, err = redis.NewConnection(redisConfig)
	return err
}

// autoMigrateTestDB runs auto migration for all entities
func autoMigrateTestDB() error {
	gormDB := TestDB.GetDB()

	return gormDB.AutoMigrate(
		&entity.User{},
		&entity.Todo{},
		&entity.Tag{},
		&entity.TodoTag{},
	)
}

// cleanTestDatabase cleans all data from test database
func cleanTestDatabase() error {
	gormDB := TestDB.GetDB()

	// Delete all data in correct order due to foreign keys
	gormDB.Exec("DELETE FROM todo_tags")
	gormDB.Exec("DELETE FROM tags")
	gormDB.Exec("DELETE FROM todos")
	gormDB.Exec("DELETE FROM users")

	return nil
}

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T) *entity.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testPassword123"), bcrypt.DefaultCost)

	user := &entity.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Role:         entity.UserRoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	userRepo := repository.NewUserRepository(TestDB.GetDB())
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// CreateTestTodo creates a test todo in the database
func CreateTestTodo(t *testing.T, userID int64) *entity.Todo {
	todo := &entity.Todo{
		UserID:      userID,
		Title:       "Test Todo",
		Description: "This is a test todo",
		Status:      entity.TodoStatusNotStarted,
		Priority:    entity.TodoPriorityMedium,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	todoRepo := repository.NewTodoRepository(TestDB.GetDB())
	if err := todoRepo.Create(todo); err != nil {
		t.Fatalf("Failed to create test todo: %v", err)
	}

	return todo
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
