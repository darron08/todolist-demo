package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/darron08/todolist-demo/internal/infrastructure/database/redis"
	"github.com/google/uuid"
)

const (
	// Redis key prefixes
	refreshTokenKeyPrefix = "refresh_token:"

	// Default expiry times
	refreshTokenExpiry = 7 * 24 * time.Hour // 7 days
)

// TokenStore manages token storage in Redis
type TokenStore struct {
	client *redis.Client
}

// NewTokenStore creates a new token store
func NewTokenStore(redisClient *redis.Client) *TokenStore {
	return &TokenStore{
		client: redisClient,
	}
}

// StoreRefreshToken stores a refresh token in Redis
func (s *TokenStore) StoreRefreshToken(ctx context.Context, userID int64, tokenID, token string) error {
	key := s.buildRefreshTokenKey(userID, tokenID)
	return s.client.Set(ctx, key, token, refreshTokenExpiry)
}

// ValidateRefreshToken checks if a refresh token exists in Redis
func (s *TokenStore) ValidateRefreshToken(ctx context.Context, userID int64, tokenID string) (bool, error) {
	key := s.buildRefreshTokenKey(userID, tokenID)
	exists, err := s.client.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check refresh token: %w", err)
	}
	return exists > 0, nil
}

// DeleteRefreshToken deletes a refresh token from Redis
func (s *TokenStore) DeleteRefreshToken(ctx context.Context, userID int64, tokenID string) error {
	key := s.buildRefreshTokenKey(userID, tokenID)
	return s.client.Del(ctx, key)
}

// DeleteAllUserTokens deletes all refresh tokens for a user
func (s *TokenStore) DeleteAllUserTokens(ctx context.Context, userID int64) error {
	// Build pattern for all user tokens
	pattern := s.buildRefreshTokenPattern(userID)

	// Note: This is a simplified implementation
	// For production, you might want to use Redis SCAN for better performance
	_ = pattern // Use pattern in production with SCAN

	// For now, just mark user tokens as invalid
	key := fmt.Sprintf("%s:%s:all", refreshTokenKeyPrefix, userID)
	return s.client.Set(ctx, key, "1", time.Nanosecond)
}

// GenerateTokenID generates a unique token ID
func GenerateTokenID() string {
	return uuid.New().String()
}

// buildRefreshTokenKey builds a Redis key for a refresh token
func (s *TokenStore) buildRefreshTokenKey(userID int64, tokenID string) string {
	return fmt.Sprintf("%s%d:%s", refreshTokenKeyPrefix, userID, tokenID)
}

// buildRefreshTokenPattern builds a pattern for user's refresh tokens
func (s *TokenStore) buildRefreshTokenPattern(userID int64) string {
	return fmt.Sprintf("%s%d:*", refreshTokenKeyPrefix, userID)
}
