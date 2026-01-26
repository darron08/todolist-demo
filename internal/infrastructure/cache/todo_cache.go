package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"github.com/darron08/todolist-demo/internal/domain/repository"
	"github.com/darron08/todolist-demo/internal/infrastructure/database/redis"
	redisv8 "github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
)

// TodoCache manages caching for todos using Sorted Set + Hash
type TodoCache struct {
	redisClient *redis.Client
	todoRepo    repository.TodoRepository
	lockManager *LockManager

	// Singleflight groups
	todoFlight             singleflight.Group
	todoListFlight         singleflight.Group
	rebuildSortedSetFlight singleflight.Group

	// TTL configurations
	hashTTL        time.Duration
	sortedSetTTL   time.Duration
	queryCacheTTL  time.Duration
	lockTimeout    time.Duration
	lockRetry      int
	lockRetryDelay time.Duration
}

// NewTodoCache creates a new todo cache instance
func NewTodoCache(redisClient *redis.Client, todoRepo repository.TodoRepository, hashTTL, sortedSetTTL, queryCacheTTL time.Duration) *TodoCache {
	return &TodoCache{
		redisClient:    redisClient,
		todoRepo:       todoRepo,
		lockManager:    NewLockManager(redisClient),
		hashTTL:        hashTTL,
		sortedSetTTL:   sortedSetTTL,
		queryCacheTTL:  queryCacheTTL,
		lockTimeout:    10 * time.Second,
		lockRetry:      3,
		lockRetryDelay: 100 * time.Millisecond,
	}
}

// CreateTodo creates a todo and updates cache using pipeline
func (tc *TodoCache) CreateTodo(ctx context.Context, todo *entity.Todo) error {
	lock := NewLock(tc.redisClient, fmt.Sprintf("todo:user:%d", todo.UserID))

	return lock.WithLockRetry(ctx, tc.lockTimeout, tc.lockRetryDelay, tc.lockRetry, func() error {
		// Use pipeline for atomic operations
		pipe := tc.redisClient.Pipeline()

		// 1. Create/update hash cache
		hashKey := BuildTodoHashKey(todo.ID)
		hashFields := BuildPipelineTodoHash(todo)
		pipe.HSet(ctx, hashKey, hashFields)
		pipe.Expire(ctx, hashKey, tc.hashTTL)

		// 2. Try to update all sorted sets (if they exist)
		tc.updateSortedSetsWithPipeline(ctx, pipe, todo)

		// 3. Execute pipeline
		_, err := tc.redisClient.ExecPipeline(pipe)
		if err != nil {
			return fmt.Errorf("failed to execute pipeline: %w", err)
		}

		// 4. Delete query caches separately (pattern deletion)
		pattern := fmt.Sprintf("%s%d%s*", TodoQueryCachePrefix, todo.UserID, QueryCacheSuffix)
		_, err = tc.redisClient.DelPattern(ctx, pattern)
		if err != nil {
			log.Printf("Warning: failed to delete query caches: %v", err)
		}

		return nil
	})
}

// UpdateTodo updates a todo and updates cache using pipeline
func (tc *TodoCache) UpdateTodo(ctx context.Context, todo *entity.Todo) error {
	lock := NewLock(tc.redisClient, fmt.Sprintf("todo:user:%d", todo.UserID))

	return lock.WithLockRetry(ctx, tc.lockTimeout, tc.lockRetryDelay, tc.lockRetry, func() error {
		// Get old todo for comparison
		oldTodo, err := tc.todoRepo.FindByID(ctx, todo.ID)
		if err != nil {
			return err
		}

		// Use pipeline for atomic operations
		pipe := tc.redisClient.Pipeline()

		// 1. Update hash cache
		hashKey := BuildTodoHashKey(todo.ID)
		hashFields := BuildPipelineTodoHash(todo)
		pipe.HSet(ctx, hashKey, hashFields)
		pipe.Expire(ctx, hashKey, tc.hashTTL)

		// 2. Update sorted sets (remove from old positions, add to new positions)
		tc.updateSortedSetsWithPipeline(ctx, pipe, todo)

		// 3. If status changed, handle status-specific sorted sets
		if oldTodo.Status != todo.Status {
			tc.handleStatusChangeWithPipeline(ctx, pipe, todo.ID, todo.UserID, string(oldTodo.Status), string(todo.Status))
		}

		// Execute pipeline
		_, err = tc.redisClient.ExecPipeline(pipe)
		if err != nil {
			return fmt.Errorf("failed to execute pipeline: %w", err)
		}

		// 4. Delete query caches separately (pattern deletion)
		pattern := fmt.Sprintf("%s%d%s*", TodoQueryCachePrefix, todo.UserID, QueryCacheSuffix)
		_, err = tc.redisClient.DelPattern(ctx, pattern)
		if err != nil {
			log.Printf("Warning: failed to delete query caches: %v", err)
		}

		return nil
	})
}

// DeleteTodo deletes a todo and cleans up cache using pipeline
func (tc *TodoCache) DeleteTodo(ctx context.Context, todoID, userID int64) error {
	lock := NewLock(tc.redisClient, fmt.Sprintf("todo:user:%d", userID))

	return lock.WithLockRetry(ctx, tc.lockTimeout, tc.lockRetryDelay, tc.lockRetry, func() error {
		// Use pipeline for atomic operations
		pipe := tc.redisClient.Pipeline()

		// 1. Delete hash cache
		hashKey := BuildTodoHashKey(todoID)
		pipe.Del(ctx, hashKey)

		// 2. Remove from all sorted sets
		sortedSetKeys := GetAllSortedSetKeys(userID)
		for _, key := range sortedSetKeys {
			pipe.ZRem(ctx, key, todoID)
		}

		// Execute pipeline
		_, err := tc.redisClient.ExecPipeline(pipe)
		if err != nil {
			return fmt.Errorf("failed to execute pipeline: %w", err)
		}

		// 3. Delete query caches separately (pattern deletion)
		pattern := fmt.Sprintf("%s%d%s*", TodoQueryCachePrefix, userID, QueryCacheSuffix)
		_, err = tc.redisClient.DelPattern(ctx, pattern)
		if err != nil {
			log.Printf("Warning: failed to delete query caches: %v", err)
		}

		return nil
	})
}

// UpdateTodoStatus updates a todo's status and updates cache
func (tc *TodoCache) UpdateTodoStatus(ctx context.Context, todoID, userID int64, newStatus string) error {
	lock := NewLock(tc.redisClient, fmt.Sprintf("todo:user:%d", userID))

	return lock.WithLockRetry(ctx, tc.lockTimeout, tc.lockRetryDelay, tc.lockRetry, func() error {
		// Get current todo
		todo, err := tc.todoRepo.FindByID(ctx, todoID)
		if err != nil {
			return err
		}

		oldStatus := string(todo.Status)

		// Use pipeline for atomic operations
		pipe := tc.redisClient.Pipeline()

		// 1. Update hash cache
		hashKey := BuildTodoHashKey(todoID)
		pipe.HSet(ctx, hashKey, "status", newStatus)

		// 2. Handle status change in sorted sets
		tc.handleStatusChangeWithPipeline(ctx, pipe, todoID, userID, oldStatus, newStatus)

		// 3. Update other sorted sets (if other fields changed)
		todo.Status = entity.TodoStatus(newStatus)
		tc.updateSortedSetsWithPipeline(ctx, pipe, todo)

		// Execute pipeline
		_, err = tc.redisClient.ExecPipeline(pipe)
		if err != nil {
			return fmt.Errorf("failed to execute pipeline: %w", err)
		}

		// 4. Delete query caches separately (pattern deletion)
		pattern := fmt.Sprintf("%s%d%s*", TodoQueryCachePrefix, userID, QueryCacheSuffix)
		_, err = tc.redisClient.DelPattern(ctx, pattern)
		if err != nil {
			log.Printf("Warning: failed to delete query caches: %v", err)
		}

		return nil
	})
}

// GetTodo retrieves a single todo from cache or database
func (tc *TodoCache) GetTodo(ctx context.Context, todoID int64) (*entity.Todo, error) {
	// 1. Try to get from hash cache
	hashKey := BuildTodoHashKey(todoID)
	hashFields, err := tc.redisClient.HGetAll(ctx, hashKey)
	if err == nil && len(hashFields) > 0 {
		todo, err := ParseTodoFromHash(hashFields)
		if err == nil {
			return todo, nil
		}
	}

	// 2. Use singleflight to prevent thundering herd
	result, err, _ := tc.todoFlight.Do(fmt.Sprintf("get-todo:%d", todoID), func() (interface{}, error) {
		// Cache miss, get from database
		todo, err := tc.todoRepo.FindByID(ctx, todoID)
		if err != nil {
			return nil, err
		}

		// Update cache (synchronous)
		_ = tc.updateHashCache(ctx, todo)

		return todo, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*entity.Todo), nil
}

// GetTodoList retrieves a paginated list of todos using sorted set or query cache
func (tc *TodoCache) GetTodoList(ctx context.Context, userID int64, filters *ListFilter, sortBy, sortOrder string, page, limit int) ([]*entity.Todo, int64, error) {
	// Check if we should use sorted set
	useSortedSet := ShouldUseSortedSet(filters, sortBy)

	if useSortedSet {
		return tc.getTodoListFromSortedSet(ctx, userID, filters, sortBy, sortOrder, page, limit)
	}

	// Use query cache for complex queries
	return tc.getTodoListFromQueryCache(ctx, userID, filters, sortBy, sortOrder, page, limit)
}

// getTodoListFromSortedSet retrieves todos using sorted set
func (tc *TodoCache) getTodoListFromSortedSet(ctx context.Context, userID int64, filters *ListFilter, sortBy, sortOrder string, page, limit int) ([]*entity.Todo, int64, error) {
	offset := (page - 1) * limit

	// Build sorted set key
	sortedSetKey := BuildSortedSetKey(userID, filters, sortBy, sortOrder)

	// Use singleflight to prevent concurrent rebuild
	result, err, _ := tc.todoListFlight.Do(sortedSetKey, func() (interface{}, error) {
		// Check if sorted set exists
		exists, err := tc.redisClient.Exists(ctx, sortedSetKey)
		if err != nil {
			return nil, err
		}

		// Rebuild sorted set if it doesn't exist (lazy loading)
		if exists == 0 {
			_, _, err = tc.rebuildSortedSetWithFlight(ctx, userID, filters, sortBy, sortOrder)
			if err != nil {
				return nil, err
			}
		}

		// Get IDs from sorted set with pagination
		start := int64(offset)
		stop := int64(offset + limit - 1)
		ids, err := tc.redisClient.ZRangeByID(ctx, sortedSetKey, start, stop)
		if err != nil {
			return nil, err
		}

		// Get total count
		total, err := tc.redisClient.ZCard(ctx, sortedSetKey)
		if err != nil {
			return nil, err
		}

		// Batch get todos from hash or database
		todos := make([]*entity.Todo, 0, len(ids))
		for _, id := range ids {
			todo, err := tc.getTodoFromCacheOrDB(ctx, id)
			if err == nil {
				todos = append(todos, todo)
			}
		}

		return &todoListResult{todos, total}, nil
	})

	if err != nil {
		return nil, 0, err
	}

	r := result.(*todoListResult)
	return r.todos, r.total, nil
}

// getTodoListFromQueryCache retrieves todos using query cache (for complex queries)
func (tc *TodoCache) getTodoListFromQueryCache(ctx context.Context, userID int64, filters *ListFilter, sortBy, sortOrder string, page, limit int) ([]*entity.Todo, int64, error) {
	// Build query cache key
	cacheKey := BuildQueryCacheKey(userID, filters, sortBy, sortOrder, page, limit)

	// Try to get from cache
	cached, err := tc.redisClient.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var response struct {
			Data  []*entity.Todo `json:"data"`
			Total int64          `json:"total"`
		}
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			return response.Data, response.Total, nil
		}
	}

	// Use singleflight to prevent thundering herd
	result, err, _ := tc.todoListFlight.Do(cacheKey, func() (interface{}, error) {
		// Cache miss, query database
		offset := (page - 1) * limit
		todos, total, err := tc.todoRepo.FindByUserIDAndFilters(
			ctx,
			userID,
			filters.Status,
			filters.Priority,
			filters.DueDateFrom,
			filters.DueDateTo,
			sortBy,
			sortOrder,
			offset,
			limit,
		)

		if err != nil {
			return nil, err
		}

		// Cache result (synchronous)
		response := struct {
			Data  []*entity.Todo `json:"data"`
			Total int64          `json:"total"`
		}{
			Data:  todos,
			Total: total,
		}

		jsonBytes, err := json.Marshal(response)
		if err == nil {
			_ = tc.redisClient.Set(ctx, cacheKey, string(jsonBytes), tc.queryCacheTTL)
		}

		return &todoListResult{todos, total}, nil
	})

	if err != nil {
		return nil, 0, err
	}

	r := result.(*todoListResult)
	return r.todos, r.total, nil
}

// Helper methods

type todoListResult struct {
	todos []*entity.Todo
	total int64
}

// rebuildSortedSetWithFlight rebuilds a single sorted set from database with singleflight
func (tc *TodoCache) rebuildSortedSetWithFlight(ctx context.Context, userID int64, filters *ListFilter, sortBy, sortOrder string) ([]*entity.Todo, int64, error) {
	key := BuildSortedSetKey(userID, filters, sortBy, sortOrder)

	result, err, _ := tc.rebuildSortedSetFlight.Do(key, func() (interface{}, error) {
		// Load all todos for this user (or with filters)
		todos, _, err := tc.todoRepo.FindByUserIDAndFilters(
			ctx,
			userID,
			filters.Status,
			filters.Priority,
			nil, nil,
			sortBy,
			sortOrder,
			0, 10000,
		)

		if err != nil {
			return nil, err
		}

		// Delete old sorted set
		pipe := tc.redisClient.Pipeline()
		pipe.Del(ctx, key)

		// Add all todos to sorted set
		for _, todoItem := range todos {
			score := GetTodoScore(todoItem, sortBy, sortOrder)
			members := []redisv8.Z{{Score: score, Member: todoItem.ID}}
			ptrMembers := make([]*redisv8.Z, len(members))
			for i := range members {
				ptrMembers[i] = &members[i]
			}
			pipe.ZAdd(ctx, key, ptrMembers...)
		}

		// Set expiration
		pipe.Expire(ctx, key, tc.sortedSetTTL)

		// Execute pipeline
		_, err = tc.redisClient.ExecPipeline(pipe)
		if err != nil {
			return nil, err
		}

		return todos, nil
	})

	if err != nil {
		return nil, 0, err
	}

	return result.([]*entity.Todo), int64(len(result.([]*entity.Todo))), nil
}

// getTodoFromCacheOrDB retrieves a todo from hash cache or database
func (tc *TodoCache) getTodoFromCacheOrDB(ctx context.Context, todoID int64) (*entity.Todo, error) {
	hashKey := BuildTodoHashKey(todoID)
	hashFields, err := tc.redisClient.HGetAll(ctx, hashKey)

	if err == nil && len(hashFields) > 0 {
		todo, err := ParseTodoFromHash(hashFields)
		if err == nil {
			return todo, nil
		}
	}

	// Use singleflight to prevent thundering herd
	result, err, _ := tc.todoFlight.Do(fmt.Sprintf("get-todo:%d", todoID), func() (interface{}, error) {
		// Get from database
		todo, err := tc.todoRepo.FindByID(ctx, todoID)
		if err != nil {
			return nil, err
		}

		// Update cache (synchronous)
		_ = tc.updateHashCache(ctx, todo)

		return todo, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*entity.Todo), nil
}

// updateHashCache updates the hash cache for a todo
func (tc *TodoCache) updateHashCache(ctx context.Context, todo *entity.Todo) error {
	hashKey := BuildTodoHashKey(todo.ID)
	hashFields := BuildPipelineTodoHash(todo)

	err := tc.redisClient.HSetAll(ctx, hashKey, hashFields)
	if err != nil {
		return err
	}

	return tc.redisClient.Expire(ctx, hashKey, tc.hashTTL)
}

// updateSortedSetsWithPipeline updates all relevant sorted sets in a pipeline
func (tc *TodoCache) updateSortedSetsWithPipeline(ctx context.Context, pipe redisv8.Pipeliner, todo *entity.Todo) {
	// Define all sorted set variations
	sortedSetConfigs := []struct {
		filters   *ListFilter
		sortBy    string
		sortOrder string
	}{
		// Base sorted sets
		{nil, "due_date", "asc"},
		{nil, "due_date", "desc"},
		{nil, "created_at", "desc"},
		{nil, "title", "asc"},

		// Status filtered sorted sets
		{&ListFilter{Status: strPtr("not_started")}, "due_date", "asc"},
		{&ListFilter{Status: strPtr("in_progress")}, "due_date", "asc"},
		{&ListFilter{Status: strPtr("completed")}, "due_date", "asc"},

		// Priority filtered sorted sets
		{&ListFilter{Priority: strPtr("high")}, "due_date", "asc"},
	}

	for _, config := range sortedSetConfigs {
		key := BuildSortedSetKey(todo.UserID, config.filters, config.sortBy, config.sortOrder)

		// Remove old (if exists)
		pipe.ZRem(ctx, key, todo.ID)

		// Add new
		score := GetTodoScore(todo, config.sortBy, config.sortOrder)
		members := []redisv8.Z{{Score: score, Member: todo.ID}}
		// Convert []Z to []*Z
		ptrMembers := make([]*redisv8.Z, len(members))
		for i := range members {
			ptrMembers[i] = &members[i]
		}
		pipe.ZAdd(ctx, key, ptrMembers...)
		pipe.Expire(ctx, key, tc.sortedSetTTL)
	}
}

// handleStatusChangeWithPipeline handles status change in sorted sets
func (tc *TodoCache) handleStatusChangeWithPipeline(ctx context.Context, pipe redisv8.Pipeliner, todoID int64, userID int64, oldStatus, newStatus string) {
	keys := BuildStatusChangeKeys(userID, oldStatus, newStatus)

	// Remove from old status sorted set
	if len(keys) > 0 {
		pipe.ZRem(ctx, keys[0], todoID)
	}

	// Add to new status sorted set (get score from todo)
	if len(keys) > 1 {
		todoInfo, err := tc.todoRepo.FindByID(ctx, todoID)
		if err == nil {
			score := GetTodoScore(todoInfo, "due_date", "asc")
			members := []redisv8.Z{{Score: score, Member: todoID}}
			// Convert []Z to []*Z
			ptrMembers := make([]*redisv8.Z, len(members))
			for i := range members {
				ptrMembers[i] = &members[i]
			}
			pipe.ZAdd(ctx, keys[1], ptrMembers...)
			pipe.Expire(ctx, keys[1], tc.sortedSetTTL)
		}
	}
}

// rebuildSortedSet rebuilds a single sorted set from database
func (tc *TodoCache) rebuildSortedSet(ctx context.Context, userID int64, filters *ListFilter, sortBy, sortOrder string) error {
	key := BuildSortedSetKey(userID, filters, sortBy, sortOrder)

	// Load all todos for this user (or with filters)
	todos, _, err := tc.todoRepo.FindByUserIDAndFilters(
		ctx,
		userID,
		filters.Status,
		filters.Priority,
		nil, nil,
		sortBy,
		sortOrder,
		0, 10000,
	)

	if err != nil {
		return err
	}

	// Delete old sorted set
	pipe := tc.redisClient.Pipeline()
	pipe.Del(ctx, key)

	// Add all todos to sorted set
	for _, todoItem := range todos {
		score := GetTodoScore(todoItem, sortBy, sortOrder)
		members := []redisv8.Z{{Score: score, Member: todoItem.ID}}
		// Convert []Z to []*Z
		ptrMembers := make([]*redisv8.Z, len(members))
		for i := range members {
			ptrMembers[i] = &members[i]
		}
		pipe.ZAdd(ctx, key, ptrMembers...)
	}

	// Set expiration
	pipe.Expire(ctx, key, tc.sortedSetTTL)

	// Execute pipeline
	_, err = tc.redisClient.ExecPipeline(pipe)
	return err
}

// Utility functions

func strPtr(s string) *string {
	return &s
}
