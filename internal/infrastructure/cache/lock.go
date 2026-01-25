package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/darron08/todolist-demo/internal/infrastructure/database/redis"
)

// RedisLock implements a distributed lock using Redis
type RedisLock struct {
	client     *redis.Client
	key        string
	locked     bool
	identifier string
}

// NewLock creates a new distributed lock instance
func NewLock(redisClient *redis.Client, resource string) *RedisLock {
	return &RedisLock{
		client:     redisClient,
		key:        BuildLockKey(resource),
		identifier: generateIdentifier(),
		locked:     false,
	}
}

// TryLock attempts to acquire the lock with a timeout
func (l *RedisLock) TryLock(ctx context.Context, expiration time.Duration) (bool, error) {
	if l.locked {
		return false, fmt.Errorf("lock is already held")
	}

	acquired, err := l.client.Lock(ctx, l.key, expiration)
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if acquired {
		l.locked = true
		return true, nil
	}

	return false, nil
}

// TryLockWithRetry attempts to acquire the lock with retries
func (l *RedisLock) TryLockWithRetry(ctx context.Context, expiration, retryInterval time.Duration, maxRetries int) (bool, error) {
	for i := 0; i < maxRetries; i++ {
		acquired, err := l.TryLock(ctx, expiration)
		if err != nil {
			return false, err
		}

		if acquired {
			return true, nil
		}

		// Wait before retrying
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

// Unlock releases the lock
func (l *RedisLock) Unlock(ctx context.Context) error {
	if !l.locked {
		return fmt.Errorf("lock is not held")
	}

	err := l.client.Unlock(ctx, l.key)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	l.locked = false
	return nil
}

// IsLocked returns whether the lock is currently held
func (l *RedisLock) IsLocked() bool {
	return l.locked
}

// WithLock executes a function while holding the lock
func (l *RedisLock) WithLock(ctx context.Context, expiration time.Duration, fn func() error) error {
	// Try to acquire lock
	acquired, err := l.TryLock(ctx, expiration)
	if err != nil {
		return err
	}

	if !acquired {
		return fmt.Errorf("failed to acquire lock")
	}

	// Ensure lock is released even if function panics
	defer func() {
		if err := l.Unlock(ctx); err != nil {
			// Log error but don't panic
			// In production, use proper logging
		}
	}()

	// Execute the function
	return fn()
}

// WithLockRetry executes a function while holding the lock with retries
func (l *RedisLock) WithLockRetry(ctx context.Context, expiration, retryInterval time.Duration, maxRetries int, fn func() error) error {
	// Try to acquire lock with retries
	acquired, err := l.TryLockWithRetry(ctx, expiration, retryInterval, maxRetries)
	if err != nil {
		return err
	}

	if !acquired {
		return fmt.Errorf("failed to acquire lock after %d retries", maxRetries)
	}

	// Ensure lock is released even if function panics
	defer func() {
		if err := l.Unlock(ctx); err != nil {
			// Log error but don't panic
			// In production, use proper logging
		}
	}()

	// Execute the function
	return fn()
}

// generateIdentifier generates a unique identifier for the lock holder
func generateIdentifier() string {
	// In a more sophisticated implementation, you might use UUID
	// For simplicity, we use a timestamp and random number
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// LockManager manages multiple locks and prevents deadlocks
type LockManager struct {
	client *redis.Client
}

// NewLockManager creates a new lock manager
func NewLockManager(redisClient *redis.Client) *LockManager {
	return &LockManager{
		client: redisClient,
	}
}

// ExecuteWithLocks executes a function while holding multiple locks in order
// to prevent deadlocks (locks are acquired in sorted order by key)
func (lm *LockManager) ExecuteWithLocks(ctx context.Context, expiration time.Duration, resources []string, fn func() error) error {
	// Sort resources to prevent deadlocks
	sortedResources := make([]string, len(resources))
	copy(sortedResources, resources)

	// Simple sort (in production, use proper sorting)
	for i := 0; i < len(sortedResources); i++ {
		for j := i + 1; j < len(sortedResources); j++ {
			if sortedResources[i] > sortedResources[j] {
				sortedResources[i], sortedResources[j] = sortedResources[j], sortedResources[i]
			}
		}
	}

	// Acquire all locks
	locks := make([]*RedisLock, len(sortedResources))
	for i, resource := range sortedResources {
		locks[i] = NewLock(lm.client, resource)
		acquired, err := locks[i].TryLock(ctx, expiration)
		if err != nil {
			// Release already acquired locks
			lm.releaseLocks(ctx, locks[:i])
			return err
		}

		if !acquired {
			// Release already acquired locks
			lm.releaseLocks(ctx, locks[:i])
			return fmt.Errorf("failed to acquire lock for resource: %s", resource)
		}
	}

	// Ensure all locks are released
	defer lm.releaseLocks(ctx, locks)

	// Execute the function
	return fn()
}

// releaseLocks releases multiple locks
func (lm *LockManager) releaseLocks(ctx context.Context, locks []*RedisLock) {
	for _, lock := range locks {
		if lock.IsLocked() {
			_ = lock.Unlock(ctx)
		}
	}
}
