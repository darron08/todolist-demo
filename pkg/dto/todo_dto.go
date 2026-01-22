package dto

import (
	"time"

	"github.com/darron08/todolist-demo/internal/domain/entity"
)

// CreateTodoRequest represents a create todo request
type CreateTodoRequest struct {
	Title       string     `json:"title" binding:"required,min=1,max=255"`
	Description string     `json:"description" binding:"max=5000"`
	DueDate     *time.Time `json:"due_date"`
	Priority    string     `json:"priority" binding:"omitempty,oneof=low medium high"`
	Tags        []string   `json:"tags" binding:"omitempty,max=10"`
}

// UpdateTodoRequest represents an update todo request
type UpdateTodoRequest struct {
	Title       *string    `json:"title" binding:"omitempty,min=1,max=255"`
	Description *string    `json:"description" binding:"omitempty,max=5000"`
	DueDate     *time.Time `json:"due_date"`
	Status      *string    `json:"status" binding:"omitempty,oneof=not_started in_progress completed"`
	Priority    *string    `json:"priority" binding:"omitempty,oneof=low medium high"`
	Tags        []string   `json:"tags" binding:"omitempty,max=10"`
}

// UpdateTodoStatusRequest represents an update todo status request
type UpdateTodoStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=not_started in_progress completed"`
}

// ListTodosRequest represents a list todos request with filters
type ListTodosRequest struct {
	Page        int        `form:"page" binding:"min=1"`
	Limit       int        `form:"limit" binding:"min=1,max=100"`
	Status      string     `form:"status" binding:"omitempty,oneof=not_started in_progress completed"`
	Priority    string     `form:"priority" binding:"omitempty,oneof=low medium high"`
	Search      string     `form:"search" binding:"max=100"`
	DueDateFrom *time.Time `form:"due_date_from" binding:"omitempty"`
	DueDateTo   *time.Time `form:"due_date_to" binding:"omitempty"`
	SortBy      string     `form:"sort_by" binding:"omitempty,oneof=due_date status title"`
	SortOrder   string     `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// TodoResponse represents a todo response
type TodoResponse struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	Tags        []TagInfo  `json:"tags,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TagInfo represents tag information in todo response
type TagInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// TodoListResponse represents a paginated todo list response
type TodoListResponse struct {
	Data       []TodoResponse `json:"data"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	Total      int64          `json:"total"`
	TotalPages int            `json:"total_pages"`
}

// ToTodoResponse converts entity.Todo to TodoResponse
func ToTodoResponse(todo *entity.Todo) TodoResponse {
	return TodoResponse{
		ID:          todo.ID,
		UserID:      todo.UserID,
		Title:       todo.Title,
		Description: todo.Description,
		DueDate:     todo.DueDate,
		Status:      string(todo.Status),
		Priority:    string(todo.Priority),
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}
}

// ToTodoResponseWithTags converts entity.Todo with tags to TodoResponse
func ToTodoResponseWithTags(todo *entity.Todo, tags []*entity.Tag) TodoResponse {
	tagInfos := make([]TagInfo, len(tags))
	for i, tag := range tags {
		tagInfos[i] = TagInfo{
			ID:   tag.ID,
			Name: tag.Name,
		}
	}

	return TodoResponse{
		ID:          todo.ID,
		UserID:      todo.UserID,
		Title:       todo.Title,
		Description: todo.Description,
		DueDate:     todo.DueDate,
		Status:      string(todo.Status),
		Priority:    string(todo.Priority),
		Tags:        tagInfos,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}
}

// ToTodoResponseList converts []entity.Todo to []TodoResponse
func ToTodoResponseList(todos []*entity.Todo) []TodoResponse {
	responses := make([]TodoResponse, len(todos))
	for i, todo := range todos {
		responses[i] = ToTodoResponse(todo)
	}
	return responses
}

// ToTagInfoList converts []*entity.Tag to []TagInfo
func ToTagInfoList(tags []*entity.Tag) []TagInfo {
	tagInfos := make([]TagInfo, len(tags))
	for i, tag := range tags {
		tagInfos[i] = TagInfo{
			ID:   tag.ID,
			Name: tag.Name,
		}
	}
	return tagInfos
}
