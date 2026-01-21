package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"github.com/darron08/todolist-demo/internal/domain/repository"
)

var (
	ErrTodoNotFound = errors.New("todo not found")
)

// TodoRepositoryImpl implements repository.TodoRepository interface
type TodoRepositoryImpl struct {
	db *gorm.DB
}

// NewTodoRepository creates a new todo repository
func NewTodoRepository(db *gorm.DB) repository.TodoRepository {
	return &TodoRepositoryImpl{db: db}
}

// Create creates a new todo
func (r *TodoRepositoryImpl) Create(todo *entity.Todo) error {
	return r.db.Create(todo).Error
}

// FindByID finds a todo by ID
func (r *TodoRepositoryImpl) FindByID(id string) (*entity.Todo, error) {
	var todo entity.Todo
	result := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&todo)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrTodoNotFound
		}
		return nil, result.Error
	}
	return &todo, nil
}

// FindByUserID finds todos by user ID with pagination
func (r *TodoRepositoryImpl) FindByUserID(userID string, offset, limit int) ([]*entity.Todo, error) {
	var todos []*entity.Todo
	result := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&todos)

	if result.Error != nil {
		return nil, result.Error
	}
	return todos, nil
}

// Update updates a todo
func (r *TodoRepositoryImpl) Update(todo *entity.Todo) error {
	result := r.db.Save(todo)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTodoNotFound
	}
	return nil
}

// Delete soft deletes a todo
func (r *TodoRepositoryImpl) Delete(id string) error {
	result := r.db.Where("id = ?", id).Delete(&entity.Todo{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTodoNotFound
	}
	return nil
}

// List lists all todos with pagination
func (r *TodoRepositoryImpl) List(offset, limit int) ([]*entity.Todo, error) {
	var todos []*entity.Todo
	result := r.db.Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&todos)

	if result.Error != nil {
		return nil, result.Error
	}
	return todos, nil
}

// FindByStatus finds todos by status with pagination
func (r *TodoRepositoryImpl) FindByStatus(status entity.TodoStatus, offset, limit int) ([]*entity.Todo, error) {
	var todos []*entity.Todo
	result := r.db.Where("status = ? AND deleted_at IS NULL", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&todos)

	if result.Error != nil {
		return nil, result.Error
	}
	return todos, nil
}

// FindByDueDate finds todos within a date range with pagination
func (r *TodoRepositoryImpl) FindByDueDate(startDate, endDate *time.Time, offset, limit int) ([]*entity.Todo, error) {
	var todos []*entity.Todo
	query := r.db.Where("deleted_at IS NULL")

	if startDate != nil {
		query = query.Where("due_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("due_date <= ?", *endDate)
	}

	result := query.Order("due_date ASC").
		Limit(limit).
		Offset(offset).
		Find(&todos)

	if result.Error != nil {
		return nil, result.Error
	}
	return todos, nil
}

// FindByUserIDAndFilters finds todos by user ID with filters
func (r *TodoRepositoryImpl) FindByUserIDAndFilters(userID string, status *string, priority *string, offset, limit int) ([]*entity.Todo, int64, error) {
	var todos []*entity.Todo
	var total int64

	query := r.db.Model(&entity.Todo{}).Where("user_id = ? AND deleted_at IS NULL", userID)

	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if priority != nil {
		query = query.Where("priority = ?", *priority)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	result := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&todos)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return todos, total, nil
}
