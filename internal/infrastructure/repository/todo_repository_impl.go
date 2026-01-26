package repository

import (
	"context"
	"errors"
	"fmt"
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
func (r *TodoRepositoryImpl) Create(ctx context.Context, todo *entity.Todo) error {
	return r.db.WithContext(ctx).Create(todo).Error
}

// FindByID finds a todo by ID
func (r *TodoRepositoryImpl) FindByID(ctx context.Context, id int64) (*entity.Todo, error) {
	var todo entity.Todo
	result := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&todo)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrTodoNotFound
		}
		return nil, result.Error
	}
	return &todo, nil
}

// FindByUserID finds todos by user ID with pagination
func (r *TodoRepositoryImpl) FindByUserID(ctx context.Context, userID int64, offset, limit int) ([]*entity.Todo, error) {
	var todos []*entity.Todo
	result := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).
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
func (r *TodoRepositoryImpl) Update(ctx context.Context, todo *entity.Todo) error {
	result := r.db.WithContext(ctx).Save(todo)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTodoNotFound
	}
	return nil
}

// Delete soft deletes a todo
func (r *TodoRepositoryImpl) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Todo{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTodoNotFound
	}
	return nil
}

// List lists all todos with pagination
func (r *TodoRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entity.Todo, error) {
	var todos []*entity.Todo
	result := r.db.WithContext(ctx).Where("deleted_at IS NULL").
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
func (r *TodoRepositoryImpl) FindByStatus(ctx context.Context, status entity.TodoStatus, offset, limit int) ([]*entity.Todo, error) {
	var todos []*entity.Todo
	result := r.db.WithContext(ctx).Where("status = ? AND deleted_at IS NULL", status).
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
func (r *TodoRepositoryImpl) FindByDueDate(ctx context.Context, startDate, endDate *time.Time, offset, limit int) ([]*entity.Todo, error) {
	var todos []*entity.Todo
	query := r.db.WithContext(ctx).Where("deleted_at IS NULL")

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
func (r *TodoRepositoryImpl) FindByUserIDAndFilters(ctx context.Context, userID int64, status *string, priority *string, dueDateFrom, dueDateTo *time.Time, sortBy, sortOrder string, offset, limit int) ([]*entity.Todo, int64, error) {
	var todos []*entity.Todo
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Todo{}).Where("user_id = ? AND deleted_at IS NULL", userID)

	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if priority != nil {
		query = query.Where("priority = ?", *priority)
	}
	if dueDateFrom != nil {
		query = query.Where("due_date >= ?", *dueDateFrom)
	}
	if dueDateTo != nil {
		query = query.Where("due_date <= ?", *dueDateTo)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Map sort_by to database column names
	var orderByColumn string
	switch sortBy {
	case "due_date":
		orderByColumn = "due_date"
	case "status":
		orderByColumn = "status"
	case "title":
		orderByColumn = "title"
	default:
		orderByColumn = "due_date"
	}

	// Build order by clause with direction
	orderClause := fmt.Sprintf("%s %s", orderByColumn, sortOrder)

	// Get paginated results with dynamic sorting
	result := query.Order(orderClause).
		Limit(limit).
		Offset(offset).
		Find(&todos)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return todos, total, nil
}

// FindByFilters finds todos with filters (without user ID restriction)
func (r *TodoRepositoryImpl) FindByFilters(ctx context.Context, status *string, priority *string, offset, limit int) ([]*entity.Todo, error) {
	var todos []*entity.Todo

	query := r.db.WithContext(ctx).Model(&entity.Todo{}).Where("deleted_at IS NULL")

	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if priority != nil {
		query = query.Where("priority = ?", *priority)
	}

	// Get paginated results
	result := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&todos)

	if result.Error != nil {
		return nil, result.Error
	}

	return todos, nil
}
