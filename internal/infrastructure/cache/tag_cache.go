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
	"golang.org/x/sync/singleflight"
)

// TagCache manages caching for tags using String
type TagCache struct {
	redisClient *redis.Client
	tagRepo     repository.TagRepository
	lockManager *LockManager

	// Singleflight groups
	tagFlight      singleflight.Group
	tagListFlight  singleflight.Group
	userTagsFlight singleflight.Group

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

	// 2. Use singleflight to prevent thundering herd
	result, err, _ := tc.tagFlight.Do(fmt.Sprintf("get-tag:%d", tagID), func() (interface{}, error) {
		// Cache miss, get from database
		tagResult, err := tc.tagRepo.FindByID(ctx, tagID)
		if err != nil {
			return nil, err
		}

		// Update cache (synchronous)
		_ = tc.updateTagCache(ctx, tagResult)

		return tagResult, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*entity.Tag), nil
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

	// 2. Use singleflight to prevent thundering herd
	result, err, _ := tc.tagListFlight.Do(fmt.Sprintf("taglist:%d:%d", page, limit), func() (interface{}, error) {
		// Cache miss, query database
		offset := (page - 1) * limit
		tags, err := tc.tagRepo.List(ctx, offset, limit)
		if err != nil {
			return nil, err
		}

		// Get total count
		allTags, err := tc.tagRepo.List(ctx, 0, 10000)
		if err != nil {
			return nil, err
		}
		total := int64(len(allTags))

		// Cache result (synchronous)
		response := struct {
			Data  []*entity.Tag `json:"data"`
			Total int64         `json:"total"`
		}{
			Data:  tags,
			Total: total,
		}

		jsonBytes, err := json.Marshal(response)
		if err == nil {
			_ = tc.redisClient.Set(ctx, cacheKey, string(jsonBytes), tc.tagTTL)
		}

		return &tagListResult{tags, total}, nil
	})

	if err != nil {
		return nil, 0, err
	}

	r := result.(*tagListResult)
	return r.tags, r.total, nil
}

type tagListResult struct {
	tags  []*entity.Tag
	total int64
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

	// 2. Use singleflight to prevent thundering herd
	result, err, _ := tc.userTagsFlight.Do(fmt.Sprintf("usertags:%d", userID), func() (interface{}, error) {
		// Cache miss, query database
		allTags, err := tc.tagRepo.List(ctx, 0, 10000)
		if err != nil {
			return nil, err
		}

		// Cache result (synchronous)
		jsonBytes, err := json.Marshal(allTags)
		if err == nil {
			_ = tc.redisClient.Set(ctx, cacheKey, string(jsonBytes), tc.tagTTL)
		}

		return allTags, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]*entity.Tag), nil
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
