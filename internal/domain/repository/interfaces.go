package repository

import (
	"github.com/darron08/todolist-demo/internal/domain/entity"
	"time"
)

// UserRepository defines the interface for user repository operations
type UserRepository interface {
	Create(user *entity.User) error
	FindByID(id string) (*entity.User, error)
	FindByUsername(username string) (*entity.User, error)
	FindByEmail(email string) (*entity.User, error)
	Update(user *entity.User) error
	Delete(id string) error
	List(offset, limit int) ([]*entity.User, error)
}

// TodoRepository defines the interface for todo repository operations
type TodoRepository interface {
	Create(todo *entity.Todo) error
	FindByID(id string) (*entity.Todo, error)
	FindByUserID(userID string, offset, limit int) ([]*entity.Todo, error)
	Update(todo *entity.Todo) error
	Delete(id string) error
	List(offset, limit int) ([]*entity.Todo, error)
	FindByStatus(status entity.TodoStatus, offset, limit int) ([]*entity.Todo, error)
	FindByDueDate(startDate, endDate *time.Time, offset, limit int) ([]*entity.Todo, error)
	FindByUserIDAndFilters(userID string, status *string, priority *string, offset, limit int) ([]*entity.Todo, int64, error)
}

// TagRepository defines the interface for tag repository operations
type TagRepository interface {
	Create(tag *entity.Tag) error
	FindByID(id string) (*entity.Tag, error)
	FindByName(name string) (*entity.Tag, error)
	Update(tag *entity.Tag) error
	Delete(id string) error
	List(offset, limit int) ([]*entity.Tag, error)
	FindByUserID(userID string) ([]*entity.Tag, error)
}
