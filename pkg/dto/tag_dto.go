package dto

// CreateTagRequest represents a create tag request
type CreateTagRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// UpdateTagRequest represents an update tag request
type UpdateTagRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// TagResponse represents a tag response
type TagResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
