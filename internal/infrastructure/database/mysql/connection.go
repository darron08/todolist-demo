package mysql

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConfig represents MySQL database configuration
type DBConfig struct {
	Host            string
	Port            string
	Username        string
	Password        string
	Database        string
	Charset         string
	ParseTime       bool
	Loc             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	LogMode         logger.LogLevel
}

// DB represents the database connection wrapper
type DB struct {
	*gorm.DB
}

// NewConnection creates a new MySQL database connection
func NewConnection(cfg *DBConfig) (*DB, error) {
	// Build DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
		cfg.ParseTime,
		cfg.Loc,
	)

	// GORM configuration
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(cfg.LogMode),
	}

	// Open database connection
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Close()
}

// HealthCheck checks if the database connection is healthy
func (db *DB) HealthCheck() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Ping()
}

// GetDB returns the underlying GORM DB instance
func (db *DB) GetDB() *gorm.DB {
	return db.DB
}
