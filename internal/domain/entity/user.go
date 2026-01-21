package entity

import (
	"time"
)

// User represents a user entity in the domain layer
type User struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // Don't include in JSON responses
	Role         UserRole   `json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

// TableName returns the table name for GORM
func (User) TableName() string {
	return "users"
}
