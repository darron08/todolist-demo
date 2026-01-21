package entity

import (
	"time"
)

// Tag represents a tag entity in the domain layer
type Tag struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// UserTag represents the many-to-many relationship between users and tags
type UserTag struct {
	UserID    string    `json:"user_id"`
	TagID     string    `json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName returns the table name for GORM
func (Tag) TableName() string {
	return "tags"
}

// TableName returns the table name for GORM
func (UserTag) TableName() string {
	return "user_tags"
}
