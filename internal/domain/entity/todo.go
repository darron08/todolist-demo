package entity

import (
	"time"
)

// TodoStatus represents the status of a todo item
type TodoStatus string

const (
	TodoStatusNotStarted TodoStatus = "not_started"
	TodoStatusInProgress TodoStatus = "in_progress"
	TodoStatusCompleted  TodoStatus = "completed"
)

// TodoPriority represents the priority of a todo item
type TodoPriority string

const (
	TodoPriorityLow    TodoPriority = "low"
	TodoPriorityMedium TodoPriority = "medium"
	TodoPriorityHigh   TodoPriority = "high"
)

// Todo represents a todo entity in the domain layer
type Todo struct {
	ID          string       `json:"id"`
	UserID      string       `json:"user_id"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	Status      TodoStatus   `json:"status"`
	Priority    TodoPriority `json:"priority"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty"`
}

// TableName returns the table name for GORM
func (Todo) TableName() string {
	return "todos"
}
