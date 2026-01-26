package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"github.com/darron08/todolist-demo/internal/domain/repository"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
	ErrEmailExists  = errors.New("email already exists")
)

// UserRepositoryImpl implements repository.UserRepository interface
type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &UserRepositoryImpl{db: db}
}

// Create creates a new user
func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	// Check if username already exists
	var existingUser entity.User
	result := r.db.WithContext(ctx).Where("username = ?", user.Username).First(&existingUser)
	if result.Error == nil {
		return ErrUserExists
	}
	if result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}

	// Check if email already exists
	result = r.db.WithContext(ctx).Where("email = ?", user.Email).First(&existingUser)
	if result.Error == nil {
		return ErrEmailExists
	}
	if result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}

	return r.db.WithContext(ctx).Create(user).Error
}

// FindByID finds a user by ID
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id int64) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindByUsername finds a user by username
func (r *UserRepositoryImpl) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// Delete soft deletes a user
func (r *UserRepositoryImpl) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// List lists users with pagination
func (r *UserRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entity.User, error) {
	var users []*entity.User
	result := r.db.WithContext(ctx).Where("deleted_at IS NULL").Order("created_at DESC").Limit(limit).Offset(offset).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}
