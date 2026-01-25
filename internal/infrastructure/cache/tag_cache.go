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
)

// TagCache manages caching for tags using String
type TagCache struct {
	redisClient *redis.Client
	tagRepo     repository.TagRepository
	lockManager *LockManager

	// TTL configurations
	tagTTL         time.Duration
	lockTimeout    time.Duration
	lockRetry      int
	lockRetryDelay time.Duration
}

// NewTagCache creates a new tag cache instance
func NewTagCache(redisClient *redis.Client, tagRepo repository.TagRepository, tagTTL time.Duration) *TagCache {
	return &TagCache{
		redisClient:    redisClient,
		tagRepo:        tagRepo,
		lockManager:    NewLockManager(redisClient),
		tagTTL:         tagTTL,
		lockTimeout:    10 * time.Second,
		lockRetry:      3,
		lockRetryDelay: 100 * time.Millisecond,
	}
}

// CreateTag creates a tag and deletes tag list caches
func (tc *TagCache) CreateTag(ctx context.Context, tag *entity.Tag) error {
	// Delete all tag list caches
	return tc.deleteAllTagCaches(ctx)
}

// UpdateTag updates a tag and deletes tag list caches
func (tc *TagCache) UpdateTag(ctx context.Context, tag *entity.Tag) error {
	// Delete all tag list caches
	return tc.deleteAllTagCaches(ctx)
}

// DeleteTag deletes a tag and deletes tag list caches
func (tc *TagCache) DeleteTag(ctx context.Context, tagID int64) error {
	// Delete tag cache
	tagKey := BuildTagStringKey(tagID)
	_ = tc.redisClient.Del(ctx, tagKey)

	// Delete all tag list caches
	return tc.deleteAllTagCaches(ctx)
}

// GetTag retrieves a single tag from string cache or database
func (tc *TagCache) GetTag(ctx context.Context, tagID int64) (*entity.Tag, error) {
	// 1. Try to get from string cache
	tagKey := BuildTagStringKey(tagID)
	cached, err := tc.redisClient.Get(ctx, tagKey)
	if err == nil && cached != "" {
		var cachedTag entity.Tag
		if unmarshalErr := json.Unmarshal([]byte(cached), &cachedTag); unmarshalErr == nil {
			return &cachedTag, nil
		}
	}

	// 2. Cache miss, get from database
	tagResult, err := tc.tagRepo.FindByID(tagID)
	if err != nil {
		return nil, err
	}

	// 3. Update cache (fire and forget)
	go tc.updateTagCache(context.Background(), tagResult)

	return tagResult, nil
}

// GetTagList retrieves a paginated list of tags from string cache or database
func (tc *TagCache) GetTagList(ctx context.Context, page, limit int) ([]*entity.Tag, int64, error) {
	cacheKey := BuildTagListKey(page, limit)

	// 1. Try to get from cache
	cached, err := tc.redisClient.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var response struct {
			Data  []*entity.Tag `json:"data"`
			Total int64         `json:"total"`
		}
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			return response.Data, response.Total, nil
		}
	}

	// 2. Cache miss, query database
	offset := (page - 1) * limit
	tags, err := tc.tagRepo.List(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	allTags, err := tc.tagRepo.List(0, 10000)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(allTags))

	// 3. Cache result
	response := struct {
		Data  []*entity.Tag `json:"data"`
		Total int64         `json:"total"`
	}{
		Data:  tags,
		Total: total,
	}

	jsonBytes, err := json.Marshal(response)
	if err == nil {
		go tc.redisClient.Set(context.Background(), cacheKey, string(jsonBytes), tc.tagTTL)
	}

	return tags, total, nil
}

// GetUserTags retrieves user tags with todo counts from string cache or database
func (tc *TagCache) GetUserTags(ctx context.Context, userID int64) ([]*entity.Tag, error) {
	cacheKey := BuildUserTagsKey(userID)

	// 1. Try to get from cache
	cached, err := tc.redisClient.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var tags []*entity.Tag
		if err := json.Unmarshal([]byte(cached), &tags); err == nil {
			return tags, nil
		}
	}

	// 2. Cache miss, query database (this method might need to be implemented in repository)
	// For now, we'll implement it using existing methods
	allTags, err := tc.tagRepo.List(0, 10000)
	if err != nil {
		return nil, err
	}

	// Note: The todo count logic is handled in the use case layer
	// Here we just return the tags

	// 3. Cache result
	jsonBytes, err := json.Marshal(allTags)
	if err == nil {
		go tc.redisClient.Set(context.Background(), cacheKey, string(jsonBytes), tc.tagTTL)
	}

	return allTags, nil
}

// Helper methods

// updateTagCache updates string cache for a tag
func (tc *TagCache) updateTagCache(ctx context.Context, tag *entity.Tag) error {
	tagKey := BuildTagStringKey(tag.ID)

	jsonBytes, err := json.Marshal(tag)
	if err != nil {
		return err
	}

	return tc.redisClient.Set(ctx, tagKey, string(jsonBytes), tc.tagTTL)
}

// deleteAllTagCaches deletes all tag-related caches
func (tc *TagCache) deleteAllTagCaches(ctx context.Context) error {
	lock := NewLock(tc.redisClient, "tags:all")

	return lock.WithLockRetry(ctx, tc.lockTimeout, tc.lockRetryDelay, tc.lockRetry, func() error {
		// Delete tag list caches (pattern: cache:tags:page:*:limit:*)
		pattern1 := fmt.Sprintf("%spage:*:limit:*", TagListKeyPrefix)
		_, err := tc.redisClient.DelPattern(ctx, pattern1)
		if err != nil {
			log.Printf("Warning: failed to delete tag list caches: %v", err)
		}

		// Delete user tags caches (pattern: cache:tags:my-tags:*)
		pattern2 := fmt.Sprintf("%s*", TagUserTagsKeyPrefix)
		_, err = tc.redisClient.DelPattern(ctx, pattern2)
		if err != nil {
			log.Printf("Warning: failed to delete user tags caches: %v", err)
		}

		return nil
	})
}

// InvalidateByUserID invalidates caches for a specific user
func (tc *TagCache) InvalidateByUserID(ctx context.Context, userID int64) error {
	userTagsKey := BuildUserTagsKey(userID)
	return tc.redisClient.Del(ctx, userTagsKey)
}
