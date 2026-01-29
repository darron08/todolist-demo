package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"github.com/darron08/todolist-demo/internal/domain/repository"
	"github.com/darron08/todolist-demo/internal/infrastructure/redis"
	"github.com/darron08/todolist-demo/pkg/dto"
	"github.com/darron08/todolist-demo/pkg/utils"
)

var (
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidUsername    = errors.New("invalid username format")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidPassword    = errors.New("password does not meet security requirements")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid or expired refresh token")
)

// Repository errors (imported from repository layer)
var (
	repoErrUserNotFound = errors.New("user not found")
)

// UserUseCase implements business logic for users
type UserUseCase struct {
	userRepo          repository.UserRepository
	jwtManager        *utils.JWTManager
	tokenStore        *redis.TokenStore
	passwordValidator *utils.Validator
	usernameValidator *utils.Validator
	emailValidator    *utils.Validator
	hashedPassword    string
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(
	userRepo repository.UserRepository,
	jwtManager *utils.JWTManager,
	tokenStore *redis.TokenStore,
) *UserUseCase {
	validator := utils.NewValidator()

	return &UserUseCase{
		userRepo:          userRepo,
		jwtManager:        jwtManager,
		tokenStore:        tokenStore,
		passwordValidator: validator,
		usernameValidator: validator,
		emailValidator:    validator,
	}
}

// Register registers a new user
func (uc *UserUseCase) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// Validate username
	if err := uc.usernameValidator.ValidateUsername(req.Username); err != nil {
		return nil, ErrInvalidUsername
	}

	// Validate email
	if !uc.emailValidator.ValidateEmail(req.Email) {
		return nil, ErrInvalidEmail
	}

	// Validate password
	if err := uc.passwordValidator.ValidatePassword(req.Password); err != nil {
		return nil, ErrInvalidPassword
	}

	// Check if username already exists
	if _, err := uc.userRepo.FindByUsername(ctx, req.Username); err == nil {
		return nil, ErrUsernameExists
	}

	// Check if email already exists
	if _, err := uc.userRepo.FindByEmail(ctx, req.Email); err == nil {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	now := time.Now()

	role := entity.UserRoleUser
	if req.Role == "admin" {
		role = entity.UserRoleAdmin
	}

	user := &entity.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Save to database
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Return response
	return &dto.RegisterResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}, nil
}

// Login authenticates a user and returns tokens
func (uc *UserUseCase) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Find user by username
	user, err := uc.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if err != nil && err.Error() == "user not found" {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Verify password
	if !utils.VerifyPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Generate access token (15 minutes)
	accessToken, err := uc.jwtManager.GenerateAccessToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token (7 days)
	refreshToken, tokenID, err := uc.jwtManager.GenerateRefreshToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in Redis
	if err := uc.tokenStore.StoreRefreshToken(ctx, user.ID, tokenID, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Calculate expiry time in seconds
	expiresIn := int64(15 * 60) // 15 minutes in seconds

	// Return response
	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User: dto.UserResponse{
			UserID:   user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     string(user.Role),
		},
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (uc *UserUseCase) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error) {
	// Validate refresh token
	claims, err := uc.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Check if refresh token exists in Redis
	valid, err := uc.tokenStore.ValidateRefreshToken(ctx, claims.UserID, claims.TokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}
	if !valid {
		return nil, ErrInvalidToken
	}

	// Generate new access token
	accessToken, err := uc.jwtManager.GenerateAccessToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate new refresh token
	newRefreshToken, newTokenID, err := uc.jwtManager.GenerateRefreshToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Delete old refresh token
	if err := uc.tokenStore.DeleteRefreshToken(ctx, claims.UserID, claims.TokenID); err != nil {
		// Log error but don't fail the operation
		_ = err
	}

	// Store new refresh token
	if err := uc.tokenStore.StoreRefreshToken(ctx, claims.UserID, newTokenID, newRefreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Calculate expiry time in seconds
	expiresIn := int64(15 * 60) // 15 minutes in seconds

	// Return response
	return &dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

// Logout logs out a user by invalidating their refresh token
func (uc *UserUseCase) Logout(ctx context.Context, userID int64, refreshTokenID string) error {
	// If refresh token ID is provided, validate and delete it
	if refreshTokenID != "" {
		valid, err := uc.tokenStore.ValidateRefreshToken(ctx, userID, refreshTokenID)
		if err != nil {
			return fmt.Errorf("failed to validate refresh token: %w", err)
		}

		if !valid {
			return ErrInvalidToken
		}

		// Delete refresh token from Redis
		if err := uc.tokenStore.DeleteRefreshToken(ctx, userID, refreshTokenID); err != nil {
			return fmt.Errorf("failed to delete refresh token: %w", err)
		}
	} else {
		// If no specific token ID, delete all refresh tokens for this user
		if err := uc.tokenStore.DeleteAllUserTokens(ctx, userID); err != nil {
			return fmt.Errorf("failed to delete all user refresh tokens: %w", err)
		}
	}

	return nil
}

// GetProfile retrieves user profile by ID
func (uc *UserUseCase) GetProfile(ctx context.Context, userID int64) (*dto.UserResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		if err != nil && err.Error() == "user not found" {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &dto.UserResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}, nil
}
