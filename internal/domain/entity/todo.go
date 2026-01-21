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
	ID          int64        `json:"id" gorm:"primaryKey;autoIncrement;type:bigint"`
	UserID      int64        `json:"user_id" gorm:"type:bigint;not null;index"`
	Title       string       `json:"title" gorm:"type:varchar(255);not null"`
	Description string       `json:"description,omitempty" gorm:"type:text"`
	DueDate     *time.Time   `json:"due_date,omitempty" gorm:"type:datetime"`
	Status      TodoStatus   `json:"status" gorm:"type:varchar(20);not null;default:'not_started'"`
	Priority    TodoPriority `json:"priority" gorm:"type:varchar(20);not null;default:'medium'"`
	CreatedAt   time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for GORM
func (Todo) TableName() string {
	return "todos"
}
