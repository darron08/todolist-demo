package repository

import (
	"github.com/darron08/todolist-demo/internal/domain/entity"
	"time"
)

// UserRepository defines the interface for user repository operations
type UserRepository interface {
	Create(user *entity.User) error
	FindByID(id int64) (*entity.User, error)
	FindByUsername(username string) (*entity.User, error)
	FindByEmail(email string) (*entity.User, error)
	Update(user *entity.User) error
	Delete(id int64) error
	List(offset, limit int) ([]*entity.User, error)
}

// TodoRepository defines the interface for todo repository operations
type TodoRepository interface {
	Create(todo *entity.Todo) error
	FindByID(id int64) (*entity.Todo, error)
	FindByUserID(userID int64, offset, limit int) ([]*entity.Todo, error)
	Update(todo *entity.Todo) error
	Delete(id int64) error
	List(offset, limit int) ([]*entity.Todo, error)
	FindByStatus(status entity.TodoStatus, offset, limit int) ([]*entity.Todo, error)
	FindByDueDate(startDate, endDate *time.Time, offset, limit int) ([]*entity.Todo, error)
	FindByUserIDAndFilters(userID int64, status *string, priority *string, dueDateFrom, dueDateTo *time.Time, sortBy, sortOrder string, offset, limit int) ([]*entity.Todo, int64, error)
	FindByFilters(status *string, priority *string, offset, limit int) ([]*entity.Todo, error)
}

// TagRepository defines the interface for tag repository operations
type TagRepository interface {
	Create(tag *entity.Tag) error
	FindByID(id int64) (*entity.Tag, error)
	FindByName(name string) (*entity.Tag, error)
	Update(tag *entity.Tag) error
	Delete(id int64) error
	List(offset, limit int) ([]*entity.Tag, error)
}

// TodoTagRepository defines the interface for todo-tag relationship operations
type TodoTagRepository interface {
	AddTagsToTodo(todoID int64, tagIDs []int64) error
	RemoveTagsFromTodo(todoID int64, tagIDs []int64) error
	ReplaceTagsForTodo(todoID int64, tagIDs []int64) error
	GetTagsByTodoID(todoID int64) ([]*entity.Tag, error)
	GetTodosByTagID(tagID int64, offset, limit int) ([]*entity.Todo, int64, error)
	GetTagStatsByUserID(userID int64) (map[int64]int64, error)
}
