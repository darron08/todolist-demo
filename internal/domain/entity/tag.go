package entity

import (
	"time"
)

// Tag represents a tag entity in the domain layer
type Tag struct {
	ID        string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name      string     `json:"name" gorm:"type:varchar(100);uniqueIndex"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type UserTag struct {
	UserID    string    `json:"user_id" gorm:"type:varchar(36);not null;primaryKey"`
	TagID     string    `json:"tag_id" gorm:"type:varchar(36);not null;primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the table name for GORM
func (Tag) TableName() string {
	return "tags"
}

// TableName returns the table name for GORM
func (UserTag) TableName() string {
	return "user_tags"
}
