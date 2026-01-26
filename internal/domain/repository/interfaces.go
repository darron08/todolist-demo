package repository

import (
	"context"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"time"
)

// UserRepository defines the interface for user repository operations
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id int64) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, offset, limit int) ([]*entity.User, error)
}

// TodoRepository defines the interface for todo repository operations
type TodoRepository interface {
	Create(ctx context.Context, todo *entity.Todo) error
	FindByID(ctx context.Context, id int64) (*entity.Todo, error)
	FindByUserID(ctx context.Context, userID int64, offset, limit int) ([]*entity.Todo, error)
	Update(ctx context.Context, todo *entity.Todo) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, offset, limit int) ([]*entity.Todo, error)
	FindByStatus(ctx context.Context, status entity.TodoStatus, offset, limit int) ([]*entity.Todo, error)
	FindByDueDate(ctx context.Context, startDate, endDate *time.Time, offset, limit int) ([]*entity.Todo, error)
	FindByUserIDAndFilters(ctx context.Context, userID int64, status *string, priority *string, dueDateFrom, dueDateTo *time.Time, sortBy, sortOrder string, offset, limit int) ([]*entity.Todo, int64, error)
	FindByFilters(ctx context.Context, status *string, priority *string, offset, limit int) ([]*entity.Todo, error)
}

// TagRepository defines the interface for tag repository operations
type TagRepository interface {
	Create(ctx context.Context, tag *entity.Tag) error
	FindByID(ctx context.Context, id int64) (*entity.Tag, error)
	FindByName(ctx context.Context, name string) (*entity.Tag, error)
	Update(ctx context.Context, tag *entity.Tag) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, offset, limit int) ([]*entity.Tag, error)
}

// TodoTagRepository defines the interface for todo-tag relationship operations
type TodoTagRepository interface {
	AddTagsToTodo(ctx context.Context, todoID int64, tagIDs []int64) error
	RemoveTagsFromTodo(ctx context.Context, todoID int64, tagIDs []int64) error
	ReplaceTagsForTodo(ctx context.Context, todoID int64, tagIDs []int64) error
	GetTagsByTodoID(ctx context.Context, todoID int64) ([]*entity.Tag, error)
	GetTodosByTagID(ctx context.Context, tagID int64, offset, limit int) ([]*entity.Todo, int64, error)
	GetTagStatsByUserID(ctx context.Context, userID int64) (map[int64]int64, error)
}
