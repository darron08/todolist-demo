package dto

import (
	"time"

	"github.com/darron08/todolist-demo/internal/domain/entity"
)

// CreateTagRequest represents a create tag request
type CreateTagRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

// UpdateTagRequest represents an update tag request
type UpdateTagRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

// ListTagsRequest represents a list tags request
type ListTagsRequest struct {
	Page  int `form:"page" binding:"min=1"`
	Limit int `form:"limit" binding:"min=1,max=100"`
}

// TagResponse represents a tag response
type TagResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	TodoCount int64     `json:"todo_count,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TagListResponse represents a paginated tag list response
type TagListResponse struct {
	Data       []TagResponse `json:"data"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	Total      int64         `json:"total"`
	TotalPages int           `json:"total_pages"`
}

// ToTagResponse converts entity.Tag to TagResponse
func ToTagResponse(tag *entity.Tag) TagResponse {
	return TagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}
}

// ToTagResponseWithCount converts entity.Tag to TagResponse with todo count
func ToTagResponseWithCount(tag *entity.Tag, todoCount int64) TagResponse {
	return TagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		TodoCount: todoCount,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}
}

// ToTagResponseList converts []*entity.Tag to []TagResponse
func ToTagResponseList(tags []*entity.Tag) []TagResponse {
	responses := make([]TagResponse, len(tags))
	for i, tag := range tags {
		responses[i] = ToTagResponse(tag)
	}
	return responses
}
