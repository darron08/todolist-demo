package utils

import (
	"testing"
	"time"

	"github.com/darron08/todolist-demo/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestJWTManager_GenerateAccessToken(t *testing.T) {
	secret := "test-secret"
	issuer := "test-issuer"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour

	jwtManager := utils.NewJWTManager(secret, issuer, accessTokenExpiry, refreshTokenExpiry)

	token, err := jwtManager.GenerateAccessToken("user123", "testuser", "user")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTManager_ValidateAccessToken(t *testing.T) {
	secret := "test-secret"
	issuer := "test-issuer"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour

	jwtManager := utils.NewJWTManager(secret, issuer, accessTokenExpiry, refreshTokenExpiry)

	token, err := jwtManager.GenerateAccessToken("user123", "testuser", "user")
	assert.NoError(t, err)

	claims, err := jwtManager.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user", claims.Role)
	assert.Equal(t, "access", claims.TokenType)
}

func TestJWTManager_GenerateRefreshToken(t *testing.T) {
	secret := "test-secret"
	issuer := "test-issuer"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour

	jwtManager := utils.NewJWTManager(secret, issuer, accessTokenExpiry, refreshTokenExpiry)

	token, tokenID, err := jwtManager.GenerateRefreshToken("user123", "testuser", "user")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, tokenID)
}

func TestJWTManager_ValidateRefreshToken(t *testing.T) {
	secret := "test-secret"
	issuer := "test-issuer"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour

	jwtManager := utils.NewJWTManager(secret, issuer, accessTokenExpiry, refreshTokenExpiry)

	token, tokenID, err := jwtManager.GenerateRefreshToken("user123", "testuser", "user")
	assert.NoError(t, err)

	claims, err := jwtManager.ValidateRefreshToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user", claims.Role)
	assert.Equal(t, tokenID, claims.TokenID)
	assert.Equal(t, "refresh", claims.TokenType)
}

func TestJWTManager_InvalidToken(t *testing.T) {
	secret := "test-secret"
	issuer := "test-issuer"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour

	jwtManager := utils.NewJWTManager(secret, issuer, accessTokenExpiry, refreshTokenExpiry)

	_, err := jwtManager.ValidateAccessToken("invalid.token.here")
	assert.Error(t, err)
}

func TestJWTManager_GetUserIDFromToken(t *testing.T) {
	secret := "test-secret"
	issuer := "test-issuer"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour

	jwtManager := utils.NewJWTManager(secret, issuer, accessTokenExpiry, refreshTokenExpiry)

	token, err := jwtManager.GenerateAccessToken("user123", "testuser", "user")
	assert.NoError(t, err)

	userID, err := jwtManager.GetUserIDFromToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "user123", userID)
}
