package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Config represents Redis configuration
type Config struct {
	Host         string
	Port         string
	Password     string
	Database     int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Client represents the Redis client wrapper
type Client struct {
	*redis.Client
}

// NewConnection creates a new Redis connection
func NewConnection(cfg *Config) (*Client, error) {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.Database,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{Client: client}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.Client.Close()
}

// HealthCheck checks if the Redis connection is healthy
func (c *Client) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.Ping(ctx).Err()
}

// GetClient returns the underlying Redis client
func (c *Client) GetClient() *redis.Client {
	return c.Client
}

// Set sets a key-value pair with expiration
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Del deletes one or more keys
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.Client.Del(ctx, keys...).Err()
}

// Exists checks if one or more keys exist
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.Client.Exists(ctx, keys...).Result()
}

// Expire sets an expiration time for a key
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.Client.Expire(ctx, key, expiration).Err()
}

// TTL returns the remaining time to live of a key
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.Client.TTL(ctx, key).Result()
}
