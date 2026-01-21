package database

import (
	"time"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"github.com/darron08/todolist-demo/internal/infrastructure/config"
	"github.com/darron08/todolist-demo/internal/infrastructure/database/mysql"
	"github.com/darron08/todolist-demo/internal/infrastructure/database/redis"
	"gorm.io/gorm/logger"
)

// Database holds database connections
type Database struct {
	MySQL *mysql.DB
	Redis *redis.Client
}

// InitializeDatabases initializes MySQL and Redis connections
func InitializeDatabases(cfg *config.Config) (*Database, error) {
	// Initialize MySQL
	mysqlDB, err := initializeMySQL(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize Redis
	redisClient, err := initializeRedis(cfg)
	if err != nil {
		return nil, err
	}

	// Auto migrate tables
	if err := autoMigrate(mysqlDB); err != nil {
		return nil, err
	}

	return &Database{
		MySQL: mysqlDB,
		Redis: redisClient,
	}, nil
}

// initializeMySQL initializes MySQL connection
func initializeMySQL(cfg *config.Config) (*mysql.DB, error) {
	logLevel := logger.Silent
	if cfg.Logger.Level == "debug" {
		logLevel = logger.Info
	}

	mysqlConfig := &mysql.DBConfig{
		Host:            cfg.Database.MySQL.Host,
		Port:            cfg.Database.MySQL.Port,
		Username:        cfg.Database.MySQL.Username,
		Password:        cfg.Database.MySQL.Password,
		Database:        cfg.Database.MySQL.Database,
		Charset:         cfg.Database.MySQL.Charset,
		ParseTime:       cfg.Database.MySQL.ParseTime,
		Loc:             cfg.Database.MySQL.Loc,
		MaxIdleConns:    cfg.Database.MySQL.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MySQL.MaxOpenConns,
		ConnMaxLifetime: time.Hour,
		LogMode:         logLevel,
	}

	return mysql.NewConnection(mysqlConfig)
}

// initializeRedis initializes Redis connection
func initializeRedis(cfg *config.Config) (*redis.Client, error) {
	dialTimeout, _ := time.ParseDuration(cfg.Database.Redis.DialTimeout)
	readTimeout, _ := time.ParseDuration(cfg.Database.Redis.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(cfg.Database.Redis.WriteTimeout)

	redisConfig := &redis.Config{
		Host:         cfg.Database.Redis.Host,
		Port:         cfg.Database.Redis.Port,
		Password:     cfg.Database.Redis.Password,
		Database:     cfg.Database.Redis.Database,
		PoolSize:     cfg.Database.Redis.PoolSize,
		MinIdleConns: cfg.Database.Redis.MinIdleConns,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return redis.NewConnection(redisConfig)
}

// autoMigrate runs auto migration for all entities
func autoMigrate(db *mysql.DB) error {
	gormDB := db.GetDB()

	// Auto migrate all entities
	return gormDB.AutoMigrate(
		&entity.User{},
		&entity.Todo{},
		&entity.Tag{},
		&entity.UserTag{},
	)
}
