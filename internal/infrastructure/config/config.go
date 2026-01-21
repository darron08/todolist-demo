package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Logger    LoggerConfig    `mapstructure:"logger"`
	CORS      CORSConfig      `mapstructure:"cors"`
	Swagger   SwaggerConfig   `mapstructure:"swagger"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Mode         string `mapstructure:"mode"`
	Port         string `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`
	IdleTimeout  string `mapstructure:"idle_timeout"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	MySQL MySQLConfig `mapstructure:"mysql"`
	Redis RedisConfig `mapstructure:"redis"`
}

// MySQLConfig represents MySQL database configuration
type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            string `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parse_time"`
	Loc             string `mapstructure:"loc"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	Database     int    `mapstructure:"database"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	DialTimeout  string `mapstructure:"dial_timeout"`
	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret    string `mapstructure:"secret"`
	ExpiresIn string `mapstructure:"expires_in"`
	Issuer    string `mapstructure:"issuer"`
}

// LoggerConfig represents logger configuration
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// SwaggerConfig represents Swagger documentation configuration
type SwaggerConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Path        string `mapstructure:"path"`
	Title       string `mapstructure:"title"`
	Description string `mapstructure:"description"`
	Version     string `mapstructure:"version"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
	Burst             int  `mapstructure:"burst"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set environment variable prefix
	viper.SetEnvPrefix("TODOLIST")
	viper.AutomaticEnv()

	// Set default values
	setDefaults()

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; use defaults and environment variables
			fmt.Println("Config file not found, using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal configuration
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override with environment variables if they exist
	overrideWithEnv(&config)

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")

	// Database defaults
	viper.SetDefault("database.mysql.host", "localhost")
	viper.SetDefault("database.mysql.port", "3306")
	viper.SetDefault("database.mysql.charset", "utf8mb4")
	viper.SetDefault("database.mysql.parse_time", true)
	viper.SetDefault("database.mysql.loc", "Local")
	viper.SetDefault("database.mysql.max_idle_conns", 10)
	viper.SetDefault("database.mysql.max_open_conns", 100)
	viper.SetDefault("database.mysql.conn_max_lifetime", "3600s")

	viper.SetDefault("database.redis.host", "localhost")
	viper.SetDefault("database.redis.port", "6379")
	viper.SetDefault("database.redis.database", 0)
	viper.SetDefault("database.redis.pool_size", 10)
	viper.SetDefault("database.redis.min_idle_conns", 5)
	viper.SetDefault("database.redis.dial_timeout", "5s")
	viper.SetDefault("database.redis.read_timeout", "3s")
	viper.SetDefault("database.redis.write_timeout", "3s")

	// JWT defaults
	viper.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	viper.SetDefault("jwt.expires_in", "24h")
	viper.SetDefault("jwt.issuer", "todolist-demo")

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
	viper.SetDefault("logger.output", "stdout")

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", []string{"*"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 86400)

	// Swagger defaults
	viper.SetDefault("swagger.enabled", true)
	viper.SetDefault("swagger.path", "/swagger")
	viper.SetDefault("swagger.title", "Todo List API")
	viper.SetDefault("swagger.description", "A high-performance todo list microservice")
	viper.SetDefault("swagger.version", "1.0")

	// Rate limit defaults
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.requests_per_minute", 100)
	viper.SetDefault("rate_limit.burst", 200)
}

// overrideWithEnv overrides configuration with environment variables
func overrideWithEnv(config *Config) {
	if port := os.Getenv("PORT"); port != "" {
		config.Server.Port = port
	}

	if host := os.Getenv("MYSQL_HOST"); host != "" {
		config.Database.MySQL.Host = host
	}

	if port := os.Getenv("MYSQL_PORT"); port != "" {
		config.Database.MySQL.Port = port
	}

	if username := os.Getenv("MYSQL_USERNAME"); username != "" {
		config.Database.MySQL.Username = username
	}

	if password := os.Getenv("MYSQL_PASSWORD"); password != "" {
		config.Database.MySQL.Password = password
	}

	if database := os.Getenv("MYSQL_DATABASE"); database != "" {
		config.Database.MySQL.Database = database
	}

	if host := os.Getenv("REDIS_HOST"); host != "" {
		config.Database.Redis.Host = host
	}

	if port := os.Getenv("REDIS_PORT"); port != "" {
		config.Database.Redis.Port = port
	}

	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		config.Database.Redis.Password = password
	}

	if db := os.Getenv("REDIS_DATABASE"); db != "" {
		if database, err := strconv.Atoi(db); err == nil {
			config.Database.Redis.Database = database
		}
	}

	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.JWT.Secret = secret
	}

	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		config.CORS.AllowedOrigins = []string{frontendURL}
	}
}
