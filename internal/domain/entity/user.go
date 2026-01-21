package entity

import (
	"time"
)

// User represents a user entity in the domain layer
type User struct {
	ID           string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Username     string     `json:"username" gorm:"type:varchar(255);uniqueIndex"`
	Email        string     `json:"email" gorm:"type:varchar(255);uniqueIndex"`
	PasswordHash string     `json:"-" gorm:"type:varchar(255);not null"` // Don't include in JSON responses
	Role         UserRole   `json:"role" gorm:"type:varchar(20);not null;default:'user'"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
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
