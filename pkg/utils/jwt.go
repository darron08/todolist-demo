package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTManager manages JWT token generation and validation
type JWTManager struct {
	secret             string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	issuer             string
}

// Claims represents JWT claims
type Claims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenID   string `json:"token_id,omitempty"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret, issuer string, accessTokenExpiry, refreshTokenExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secret:             secret,
		issuer:             issuer,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// GenerateAccessToken generates an access token
func (j *JWTManager) GenerateAccessToken(userID int64, username, role string) (string, error) {
	return j.generateToken(userID, username, role, "", "access", j.accessTokenExpiry)
}

// GenerateRefreshToken generates a refresh token
func (j *JWTManager) GenerateRefreshToken(userID int64, username, role string) (string, string, error) {
	tokenID := uuid.New().String()
	token, err := j.generateToken(userID, username, role, tokenID, "refresh", j.refreshTokenExpiry)
	return token, tokenID, err
}

// generateToken generates a JWT token
func (j *JWTManager) generateToken(userID int64, username, role, tokenID, tokenType string, expiry time.Duration) (string, error) {
	now := time.Now()

	claims := &Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenID:   tokenID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

// ValidateToken validates a JWT token and returns claims
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateAccessToken validates an access token
func (j *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, errors.New("invalid token type: expected access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (j *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type: expected refresh token")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a refresh token
func (j *JWTManager) RefreshAccessToken(refreshTokenString string) (string, string, error) {
	claims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", err
	}

	newToken, newTokenID, err := j.GenerateRefreshToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		return "", "", err
	}

	return newToken, newTokenID, nil
}

// GetUserIDFromToken extracts user ID from token
func (j *JWTManager) GetUserIDFromToken(tokenString string) (int64, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
