package entity

import (
	"time"
)

// Tag represents a tag entity in the domain layer
type Tag struct {
	ID        int64      `json:"id" gorm:"primaryKey;autoIncrement;type:bigint"`
	Name      string     `json:"name" gorm:"type:varchar(100);uniqueIndex"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// TodoTag represents the many-to-many relationship between todos and tags
type TodoTag struct {
	TodoID    int64     `json:"todo_id" gorm:"type:bigint;not null;primaryKey"`
	TagID     int64     `json:"tag_id" gorm:"type:bigint;not null;primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the table name for GORM
func (Tag) TableName() string {
	return "tags"
}

// TableName returns the table name for GORM
func (TodoTag) TableName() string {
	return "todo_tags"
}
