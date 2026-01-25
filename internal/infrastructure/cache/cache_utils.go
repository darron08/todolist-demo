package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
	"time"

	"github.com/darron08/todolist-demo/internal/domain/entity"
)

// ListFilter represents filter parameters for todo list queries
type ListFilter struct {
	Status      *string
	Priority    *string
	Search      string
	DueDateFrom *time.Time
	DueDateTo   *time.Time
}

// Key prefix constants
const (
	// Todo cache keys
	TodoHashKeyPrefix    = "cache:todo:"
	TodoSortedSetPrefix  = "cache:todos:user:"
	TodoQueryCachePrefix = "cache:todos:user:"
	QueryCacheSuffix     = ":query:"

	// Tag cache keys
	TagStringKeyPrefix   = "cache:tag:"
	TagListKeyPrefix     = "cache:tags:"
	TagUserTagsKeyPrefix = "cache:tags:my-tags:"

	// Lock keys
	LockKeyPrefix = "lock:"
)

// buildSortedSetKey builds a sorted set key for todos
func BuildSortedSetKey(userID int64, filters *ListFilter, sortBy, sortOrder string) string {
	base := fmt.Sprintf("%s%d:sorted:", TodoSortedSetPrefix, userID)

	if filters != nil {
		if filters.Status != nil {
			base += fmt.Sprintf("status:%s:", *filters.Status)
		}
		if filters.Priority != nil {
			base += fmt.Sprintf("priority:%s:", *filters.Priority)
		}
	}

	base += fmt.Sprintf("%s:%s", sortBy, sortOrder)
	return base
}

// buildQueryCacheKey builds a query cache key for complex todo queries
func BuildQueryCacheKey(userID int64, filters *ListFilter, sortBy, sortOrder string, page, limit int) string {
	queryData := map[string]interface{}{
		"user_id":    userID,
		"status":     filters,
		"priority":   filters,
		"search":     filters.Search,
		"due_from":   filters.DueDateFrom,
		"due_to":     filters.DueDateTo,
		"sort_by":    sortBy,
		"sort_order": sortOrder,
		"page":       page,
		"limit":      limit,
	}

	hash := CalculateHash(queryData)
	return fmt.Sprintf("%s%d%s%s", TodoQueryCachePrefix, userID, QueryCacheSuffix, hash)
}

// buildTodoHashKey builds a hash key for a single todo
func BuildTodoHashKey(todoID int64) string {
	return fmt.Sprintf("%s%d", TodoHashKeyPrefix, todoID)
}

// buildTagStringKey builds a string key for a single tag
func BuildTagStringKey(tagID int64) string {
	return fmt.Sprintf("%s%d", TagStringKeyPrefix, tagID)
}

// buildTagListKey builds a string key for tag list
func BuildTagListKey(page, limit int) string {
	return fmt.Sprintf("%spage:%d:limit:%d", TagListKeyPrefix, page, limit)
}

// buildUserTagsKey builds a string key for user tags
func BuildUserTagsKey(userID int64) string {
	return fmt.Sprintf("%s%d", TagUserTagsKeyPrefix, userID)
}

// buildLockKey builds a lock key
func BuildLockKey(resource string) string {
	return fmt.Sprintf("%s%s", LockKeyPrefix, resource)
}

// getTodoScore calculates the score for a todo in a sorted set
func GetTodoScore(todo *entity.Todo, sortBy, sortOrder string) float64 {
	switch sortBy {
	case "due_date":
		return getDueDateScore(todo.DueDate, sortOrder)

	case "created_at":
		return getCreatedAtScore(todo.CreatedAt, sortOrder)

	case "title":
		return getTitleScore(todo.Title)

	default:
		// Default to created_at descending
		return getCreatedAtScore(todo.CreatedAt, "desc")
	}
}

// getDueDateScore calculates score for due_date sorting
func getDueDateScore(dueDate *time.Time, sortOrder string) float64 {
	if dueDate == nil {
		// No due date - put at the end or beginning based on sort order
		if sortOrder == "desc" {
			return math.Inf(-1) // Negative infinity (first in descending)
		}
		return math.Inf(1) // Positive infinity (last in ascending)
	}

	timestamp := dueDate.Unix()
	if sortOrder == "desc" {
		return float64(-timestamp)
	}
	return float64(timestamp)
}

// getCreatedAtScore calculates score for created_at sorting
func getCreatedAtScore(createdAt time.Time, sortOrder string) float64 {
	timestamp := createdAt.Unix()
	if sortOrder == "desc" {
		return float64(-timestamp)
	}
	return float64(timestamp)
}

// getTitleScore converts title to a numeric score using FNV hash
func getTitleScore(title string) float64 {
	h := fnv.New32a()
	h.Write([]byte(title))
	return float64(h.Sum32())
}

// getAllSortedSetKeys returns all sorted set keys for a user
func GetAllSortedSetKeys(userID int64) []string {
	base := fmt.Sprintf("%s%d:sorted:", TodoSortedSetPrefix, userID)

	keys := []string{
		base + "due_date:asc",
		base + "due_date:desc",
		base + "created_at:desc",
		base + "title:asc",
		base + "status:not_started:due_date:asc",
		base + "status:in_progress:due_date:asc",
		base + "status:completed:due_date:asc",
		base + "priority:high:due_date:asc",
	}

	return keys
}

// parseTodoFromHash parses a todo entity from hash fields
func ParseTodoFromHash(fields map[string]string) (*entity.Todo, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("empty hash fields")
	}

	todo := &entity.Todo{}

	// Parse ID
	if idStr, ok := fields["id"]; ok {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse id: %w", err)
		}
		todo.ID = id
	}

	// Parse UserID
	if userIDStr, ok := fields["user_id"]; ok {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user_id: %w", err)
		}
		todo.UserID = userID
	}

	// Parse basic fields
	if title, ok := fields["title"]; ok {
		todo.Title = title
	}

	if description, ok := fields["description"]; ok {
		todo.Description = description
	}

	if status, ok := fields["status"]; ok {
		todo.Status = entity.TodoStatus(status)
	}

	if priority, ok := fields["priority"]; ok {
		todo.Priority = entity.TodoPriority(priority)
	}

	// Parse DueDate
	if dueDateStr, ok := fields["due_date"]; ok && dueDateStr != "" {
		timestamp, err := strconv.ParseInt(dueDateStr, 10, 64)
		if err == nil {
			dueDate := time.Unix(timestamp, 0)
			todo.DueDate = &dueDate
		}
	}

	// Parse CreatedAt
	if createdAtStr, ok := fields["created_at"]; ok && createdAtStr != "" {
		timestamp, err := strconv.ParseInt(createdAtStr, 10, 64)
		if err == nil {
			todo.CreatedAt = time.Unix(timestamp, 0)
		}
	}

	// Parse UpdatedAt
	if updatedAtStr, ok := fields["updated_at"]; ok && updatedAtStr != "" {
		timestamp, err := strconv.ParseInt(updatedAtStr, 10, 64)
		if err == nil {
			todo.UpdatedAt = time.Unix(timestamp, 0)
		}
	}

	return todo, nil
}

// buildPipelineTodoHash builds a hash fields map from a todo entity
func BuildPipelineTodoHash(todo *entity.Todo) map[string]interface{} {
	fields := map[string]interface{}{
		"id":         todo.ID,
		"user_id":    todo.UserID,
		"title":      todo.Title,
		"status":     string(todo.Status),
		"priority":   string(todo.Priority),
		"created_at": todo.CreatedAt.Unix(),
		"updated_at": todo.UpdatedAt.Unix(),
	}

	if todo.Description != "" {
		fields["description"] = todo.Description
	}

	if todo.DueDate != nil {
		fields["due_date"] = todo.DueDate.Unix()
	}

	return fields
}

// parseTagFromHash parses a tag entity from hash fields
func ParseTagFromHash(fields map[string]string) (*entity.Tag, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("empty hash fields")
	}

	tag := &entity.Tag{}

	// Parse ID
	if idStr, ok := fields["id"]; ok {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse id: %w", err)
		}
		tag.ID = id
	}

	// Parse Name
	if name, ok := fields["name"]; ok {
		tag.Name = name
	}

	// Parse CreatedAt
	if createdAtStr, ok := fields["created_at"]; ok && createdAtStr != "" {
		timestamp, err := strconv.ParseInt(createdAtStr, 10, 64)
		if err == nil {
			tag.CreatedAt = time.Unix(timestamp, 0)
		}
	}

	// Parse UpdatedAt
	if updatedAtStr, ok := fields["updated_at"]; ok && updatedAtStr != "" {
		timestamp, err := strconv.ParseInt(updatedAtStr, 10, 64)
		if err == nil {
			tag.UpdatedAt = time.Unix(timestamp, 0)
		}
	}

	return tag, nil
}

// buildPipelineTagHash builds a hash fields map from a tag entity
func BuildPipelineTagHash(tag *entity.Tag) map[string]interface{} {
	return map[string]interface{}{
		"id":         tag.ID,
		"name":       tag.Name,
		"created_at": tag.CreatedAt.Unix(),
		"updated_at": tag.UpdatedAt.Unix(),
	}
}

// shouldUseSortedSet determines if a query should use sorted set
func ShouldUseSortedSet(filters *ListFilter, sortBy string) bool {
	// Complex filters should use string cache instead
	if filters != nil {
		// Date range filtering
		if filters.DueDateFrom != nil || filters.DueDateTo != nil {
			return false
		}

		// Search functionality
		if filters.Search != "" {
			return false
		}

		// Multiple filters (status + priority together)
		if filters.Status != nil && filters.Priority != nil {
			return false
		}
	}

	// Valid sort fields
	validSortFields := map[string]bool{
		"due_date":   true,
		"created_at": true,
		"title":      true,
	}

	return validSortFields[sortBy]
}

// calculateHash calculates MD5 hash of query parameters
func CalculateHash(data interface{}) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	hash := md5.Sum(jsonBytes)
	return hex.EncodeToString(hash[:])
}

// buildStatusChangeKeys returns old and new sorted set keys for status change
func BuildStatusChangeKeys(userID int64, oldStatus, newStatus string) []string {
	oldKey := fmt.Sprintf("%s%d:sorted:status:%s:due_date:asc", TodoSortedSetPrefix, userID, oldStatus)
	newKey := fmt.Sprintf("%s%d:sorted:status:%s:due_date:asc", TodoSortedSetPrefix, userID, newStatus)

	return []string{oldKey, newKey}
}

// buildPrioritySortedSetKey builds a sorted set key with priority filter
func BuildPrioritySortedSetKey(userID int64, priority, sortBy, sortOrder string) string {
	base := fmt.Sprintf("%s%d:sorted:priority:%s:", TodoSortedSetPrefix, userID, priority)
	return base + fmt.Sprintf("%s:%s", sortBy, sortOrder)
}
