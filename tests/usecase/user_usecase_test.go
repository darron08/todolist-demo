package usecase_test

import (
	"errors"
	"testing"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"github.com/darron08/todolist-demo/internal/usecase"
	"github.com/darron08/todolist-demo/pkg/dto"
	"github.com/darron08/todolist-demo/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"time"
)

// MockUserRepository is a mock for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *entity.User) error {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(id string) (*entity.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(username string) (*entity.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*entity.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *entity.User) error {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

func (m *MockUserRepository) List(offset, limit int) ([]*entity.User, error) {
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

// MockTokenStore is a mock for testing
type MockTokenStore struct {
	mock.Mock
}

func (m *MockTokenStore) StoreRefreshToken(ctx interface{}, userID, tokenID, token string) error {
	args := m.Called(ctx, userID, tokenID, token)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

func (m *MockTokenStore) ValidateRefreshToken(ctx interface{}, userID, tokenID string) (bool, error) {
	args := m.Called(ctx, userID, tokenID)
	if args.Get(0) == nil {
		return false, args.Error(1)
	}
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockTokenStore) DeleteRefreshToken(ctx interface{}, userID, tokenID string) error {
	args := m.Called(ctx, userID, tokenID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

func TestUserUseCase_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTokenStore := new(MockTokenStore)
	jwtManager := utils.NewJWTManager("test-secret", "test", 15*time.Minute, 7*24*time.Hour)

	userUseCase := usecase.NewUserUseCase(mockRepo, jwtManager, mockTokenStore)

	req := &dto.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "StrongPass123!",
	}

	// Setup mock expectations
	mockRepo.On("FindByUsername", req.Username).Return(nil, errors.New("user not found"))
	mockRepo.On("FindByEmail", req.Email).Return(nil, errors.New("user not found"))
	mockRepo.On("Create", mock.AnythingOfType("*entity.User")).Return(nil)

	response, err := userUseCase.Register(req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "testuser", response.Username)
	assert.Equal(t, "test@example.com", response.Email)
	assert.Equal(t, "user", response.Role)
}

func TestUserUseCase_Register_UsernameExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTokenStore := new(MockTokenStore)
	jwtManager := utils.NewJWTManager("test-secret", "test", 15*time.Minute, 7*24*time.Hour)

	userUseCase := usecase.NewUserUseCase(mockRepo, jwtManager, mockTokenStore)

	req := &dto.RegisterRequest{
		Username: "existinguser",
		Email:    "test@example.com",
		Password: "StrongPass123!",
	}

	// Setup mock expectations
	mockRepo.On("FindByUsername", req.Username).Return(&entity.User{ID: "123"}, nil)

	_, err := userUseCase.Register(req)

	assert.Error(t, err)
	assert.Equal(t, usecase.ErrUsernameExists, err)
}

func TestUserUseCase_Register_EmailExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTokenStore := new(MockTokenStore)
	jwtManager := utils.NewJWTManager("test-secret", "test", 15*time.Minute, 7*24*time.Hour)

	userUseCase := usecase.NewUserUseCase(mockRepo, jwtManager, mockTokenStore)

	req := &dto.RegisterRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "StrongPass123!",
	}

	// Setup mock expectations
	mockRepo.On("FindByUsername", req.Username).Return(nil, errors.New("user not found"))
	mockRepo.On("FindByEmail", req.Email).Return(&entity.User{ID: "123"}, nil)

	_, err := userUseCase.Register(req)

	assert.Error(t, err)
	assert.Equal(t, usecase.ErrEmailExists, err)
}

func TestUserUseCase_Register_WeakPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTokenStore := new(MockTokenStore)
	jwtManager := utils.NewJWTManager("test-secret", "test", 15*time.Minute, 7*24*time.Hour)

	userUseCase := usecase.NewUserUseCase(mockRepo, jwtManager, mockTokenStore)

	req := &dto.RegisterRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "weak",
	}

	_, err := userUseCase.Register(req)

	assert.Error(t, err)
	assert.Equal(t, usecase.ErrInvalidPassword, err)
}

func TestUserUseCase_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTokenStore := new(MockTokenStore)
	jwtManager := utils.NewJWTManager("test-secret", "test", 15*time.Minute, 7*24*time.Hour)

	userUseCase := usecase.NewUserUseCase(mockRepo, jwtManager, mockTokenStore)

	req := &dto.LoginRequest{
		Username: "testuser",
		Password: "StrongPass123!",
	}

	hashedPassword, _ := utils.HashPassword(req.Password)

	// Setup mock expectations
	mockRepo.On("FindByUsername", req.Username).Return(&entity.User{
		ID:           "123",
		Username:     req.Username,
		PasswordHash: hashedPassword,
		Role:         entity.UserRoleUser,
	}, nil)
	mockTokenStore.On("StoreRefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	response, err := userUseCase.Login(req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, int64(900), response.ExpiresIn)
	assert.Equal(t, "testuser", response.User.Username)
}

func TestUserUseCase_Login_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTokenStore := new(MockTokenStore)
	jwtManager := utils.NewJWTManager("test-secret", "test", 15*time.Minute, 7*24*time.Hour)

	userUseCase := usecase.NewUserUseCase(mockRepo, jwtManager, mockTokenStore)

	req := &dto.LoginRequest{
		Username: "testuser",
		Password: "WrongPassword456!",
	}

	// Setup mock expectations
	mockRepo.On("FindByUsername", req.Username).Return(&entity.User{
		Username:     req.Username,
		PasswordHash: "hashed_password",
		Role:         entity.UserRoleUser,
	}, nil)

	_, err := userUseCase.Login(req)

	assert.Error(t, err)
	assert.Equal(t, usecase.ErrInvalidCredentials, err)
}

func TestUserUseCase_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTokenStore := new(MockTokenStore)
	jwtManager := utils.NewJWTManager("test-secret", "test", 15*time.Minute, 7*24*time.Hour)

	userUseCase := usecase.NewUserUseCase(mockRepo, jwtManager, mockTokenStore)

	req := &dto.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}

	// Setup mock expectations
	mockRepo.On("FindByUsername", req.Username).Return(nil, errors.New("user not found"))

	_, err := userUseCase.Login(req)

	assert.Error(t, err)
	assert.Equal(t, usecase.ErrInvalidCredentials, err)
}
