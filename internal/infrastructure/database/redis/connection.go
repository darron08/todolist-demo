package redis

import (
	"context"
	"fmt"
	"strconv"
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

// Hash operations

// HSetAll sets multiple fields in a hash
func (c *Client) HSetAll(ctx context.Context, key string, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	return c.Client.HSet(ctx, key, fields).Err()
}

// HGetAll gets all fields and values from a hash
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.Client.HGetAll(ctx, key).Result()
}

// HGet gets a specific field from a hash
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.Client.HGet(ctx, key, field).Result()
}

// HSet sets a specific field in a hash
func (c *Client) HSet(ctx context.Context, key, field string, value interface{}) error {
	return c.Client.HSet(ctx, key, field, value).Err()
}

// HDel deletes one or more fields from a hash
func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	if len(fields) == 0 {
		return nil
	}
	return c.Client.HDel(ctx, key, fields...).Err()
}

// HExists checks if a field exists in a hash
func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	return c.Client.HExists(ctx, key, field).Result()
}

// Sorted Set operations

// ZAddByID adds an ID with a score to a sorted set
func (c *Client) ZAddByID(ctx context.Context, key string, score float64, id int64) error {
	return c.Client.ZAdd(ctx, key, &redis.Z{Score: score, Member: id}).Err()
}

// ZAddByScore adds multiple IDs with scores to a sorted set
func (c *Client) ZAddByScore(ctx context.Context, key string, members []redis.Z) error {
	if len(members) == 0 {
		return nil
	}
	// Convert []redis.Z to []*redis.Z for the ZAdd method
	ptrMembers := make([]*redis.Z, len(members))
	for i := range members {
		ptrMembers[i] = &members[i]
	}
	return c.Client.ZAdd(ctx, key, ptrMembers...).Err()
}

// ZRemByID removes an ID from a sorted set
func (c *Client) ZRemByID(ctx context.Context, key string, id int64) error {
	return c.Client.ZRem(ctx, key, id).Err()
}

// ZRangeByID gets IDs from a sorted set by range
func (c *Client) ZRangeByID(ctx context.Context, key string, start, stop int64) ([]int64, error) {
	results, err := c.Client.ZRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(results))
	for _, r := range results {
		id, err := strconv.ParseInt(r, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// ZRangeByScoreWithIDs gets IDs from a sorted set by score range
func (c *Client) ZRangeByScoreWithIDs(ctx context.Context, key string, opt *redis.ZRangeBy) ([]int64, error) {
	results, err := c.Client.ZRangeByScore(ctx, key, opt).Result()
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(results))
	for _, r := range results {
		id, err := strconv.ParseInt(r, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// ZScoreByID gets the score of an ID in a sorted set
func (c *Client) ZScoreByID(ctx context.Context, key string, id int64) (float64, error) {
	return c.Client.ZScore(ctx, key, strconv.FormatInt(id, 10)).Result()
}

// ZCount gets the number of members in a sorted set
func (c *Client) ZCount(ctx context.Context, key string, min, max string) (int64, error) {
	return c.Client.ZCount(ctx, key, min, max).Result()
}

// ZCard gets the number of members in a sorted set
func (c *Client) ZCard(ctx context.Context, key string) (int64, error) {
	return c.Client.ZCard(ctx, key).Result()
}

// Pipeline operations

// Pipeline returns a new pipeline
func (c *Client) Pipeline() redis.Pipeliner {
	return c.Client.Pipeline()
}

// ExecPipeline executes a pipeline
func (c *Client) ExecPipeline(pipeline redis.Pipeliner) ([]redis.Cmder, error) {
	return pipeline.Exec(context.Background())
}

// Pattern deletion operations

// DelPattern deletes keys matching a pattern using SCAN
func (c *Client) DelPattern(ctx context.Context, pattern string) (int64, error) {
	var keys []string
	var cursor uint64

	for {
		var scanKeys []string
		var err error

		scanKeys, cursor, err = c.Client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return 0, err
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return 0, nil
	}

	count, err := c.Client.Del(ctx, keys...).Result()
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Distributed lock operations

// Lock acquires a distributed lock
func (c *Client) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("lock:%s", key)

	result, err := c.Client.SetNX(ctx, lockKey, "1", expiration).Result()
	if err != nil {
		return false, err
	}

	return result, nil
}

// Unlock releases a distributed lock
func (c *Client) Unlock(ctx context.Context, key string) error {
	lockKey := fmt.Sprintf("lock:%s", key)
	return c.Client.Del(ctx, lockKey).Err()
}

// TryLockWithRetry tries to acquire a lock with retries
func (c *Client) TryLockWithRetry(ctx context.Context, key string, expiration, retryInterval time.Duration, maxRetries int) (bool, error) {
	for i := 0; i < maxRetries; i++ {
		acquired, err := c.Lock(ctx, key, expiration)
		if err != nil {
			return false, err
		}

		if acquired {
			return true, nil
		}

		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			case <-time.After(retryInterval):
			}
		}
	}

	return false, fmt.Errorf("failed to acquire lock after %d retries", maxRetries)
}
