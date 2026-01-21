package usecase

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"github.com/darron08/todolist-demo/internal/domain/repository"
	"github.com/darron08/todolist-demo/pkg/dto"
	"github.com/darron08/todolist-demo/pkg/utils"
)

var (
	ErrAdminPermission = errors.New("insufficient permissions: admin role required")
)

// AdminUseCase implements business logic for admin operations
type AdminUseCase struct {
	userRepo          repository.UserRepository
	todoRepo          repository.TodoRepository
	passwordValidator *utils.Validator
}

// NewAdminUseCase creates a new admin use case
func NewAdminUseCase(
	userRepo repository.UserRepository,
	todoRepo repository.TodoRepository,
) *AdminUseCase {
	validator := utils.NewValidator()
	return &AdminUseCase{
		userRepo:          userRepo,
		todoRepo:          todoRepo,
		passwordValidator: validator,
	}
}

// CreateUser creates a new user with specified role (admin only)
func (uc *AdminUseCase) CreateUser(userID int64, req *dto.AdminCreateUserRequest) (*dto.RegisterResponse, error) {
	role := entity.UserRoleUser
	if req.Role == "admin" {
		role = entity.UserRoleAdmin
	}

	// Validate username
	if err := uc.passwordValidator.ValidateUsername(req.Username); err != nil {
		return nil, ErrInvalidUsername
	}

	// Validate email
	if !uc.passwordValidator.ValidateEmail(req.Email) {
		return nil, ErrInvalidEmail
	}

	// Validate password
	if err := uc.passwordValidator.ValidatePassword(req.Password); err != nil {
		return nil, ErrInvalidPassword
	}

	// Check if username already exists
	if _, err := uc.userRepo.FindByUsername(req.Username); err == nil {
		return nil, ErrUsernameExists
	}

	// Check if email already exists
	if _, err := uc.userRepo.FindByEmail(req.Email); err == nil {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	now := time.Now()
	user := &entity.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Save to database
	if err := uc.userRepo.Create(user); err != nil {
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

// ListAllUsers lists all users with pagination (admin only)
func (uc *AdminUseCase) ListAllUsers(offset, limit int) (*dto.UserListResponse, error) {
	users, err := uc.userRepo.List(offset, limit)
	if err != nil {
		return nil, err
	}

	// Calculate total
	total, err := uc.countTotalUsers()
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Convert to response
	return &dto.UserListResponse{
		Data:       dto.ToUserResponseList(users),
		Page:       (offset / limit) + 1,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// GetUser gets a specific user by ID (admin only)
func (uc *AdminUseCase) GetUser(id int64) (*dto.UserResponse, error) {
	user, err := uc.userRepo.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &dto.UserResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}, nil
}

// DeleteUser deletes a user by ID (admin only)
func (uc *AdminUseCase) DeleteUser(id int64) error {
	return uc.userRepo.Delete(id)
}

// ListAllTodos lists all todos from all users with pagination and filters (admin only)
func (uc *AdminUseCase) ListAllTodos(page, limit int, status *string, priority *string) (*dto.TodoListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get todos with filters (without user ID filter)
	todos, err := uc.todoRepo.FindByFilters(status, priority, offset, limit)
	if err != nil {
		return nil, err
	}

	// Calculate total
	total, err := uc.countTotalTodos()
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Convert to response
	return &dto.TodoListResponse{
		Data:       dto.ToTodoResponseList(todos),
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// DeleteAnyTodo deletes any todo by ID regardless of ownership (admin only)
func (uc *AdminUseCase) DeleteAnyTodo(id int64) error {
	return uc.todoRepo.Delete(id)
}

// countTotalUsers counts total users
func (uc *AdminUseCase) countTotalUsers() (int64, error) {
	users, err := uc.userRepo.List(0, 0)
	if err != nil {
		return 0, err
	}
	return int64(len(users)), nil
}

// countTotalTodos counts total todos
func (uc *AdminUseCase) countTotalTodos() (int64, error) {
	todos, err := uc.todoRepo.List(0, 0)
	if err != nil {
		return 0, err
	}
	return int64(len(todos)), nil
}
