package usecase

import (
	"errors"
	"math"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"github.com/darron08/todolist-demo/internal/domain/repository"
	"github.com/darron08/todolist-demo/pkg/dto"
)

var (
	ErrTagNameRequired = errors.New("tag name is required")
	ErrTagNameTooLong  = errors.New("tag name is too long")
)

// TagUseCase implements business logic for tags
type TagUseCase struct {
	tagRepo     repository.TagRepository
	todoTagRepo repository.TodoTagRepository
}

// NewTagUseCase creates a new tag use case
func NewTagUseCase(tagRepo repository.TagRepository, todoTagRepo repository.TodoTagRepository) *TagUseCase {
	return &TagUseCase{
		tagRepo:     tagRepo,
		todoTagRepo: todoTagRepo,
	}
}

// CreateTag creates a new tag
func (uc *TagUseCase) CreateTag(req *dto.CreateTagRequest) (*dto.TagResponse, error) {
	// Validate name
	if req.Name == "" {
		return nil, ErrTagNameRequired
	}
	if len(req.Name) > 100 {
		return nil, ErrTagNameTooLong
	}

	// Check if tag already exists
	_, err := uc.tagRepo.FindByName(req.Name)
	if err == nil {
		return nil, errors.New("tag with this name already exists")
	}

	// Create tag entity
	tag := &entity.Tag{
		Name: req.Name,
	}

	// Save to database
	if err := uc.tagRepo.Create(tag); err != nil {
		return nil, err
	}

	response := dto.ToTagResponse(tag)
	return &response, nil
}

// GetTag retrieves a single tag by ID
func (uc *TagUseCase) GetTag(id int64) (*dto.TagResponse, error) {
	tag, err := uc.tagRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	response := dto.ToTagResponse(tag)
	return &response, nil
}

// UpdateTag updates an existing tag
func (uc *TagUseCase) UpdateTag(id int64, req *dto.UpdateTagRequest) (*dto.TagResponse, error) {
	// Get existing tag
	tag, err := uc.tagRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Validate name
	if req.Name == "" {
		return nil, ErrTagNameRequired
	}
	if len(req.Name) > 100 {
		return nil, ErrTagNameTooLong
	}

	// Check if another tag with the same name exists
	existingTag, err := uc.tagRepo.FindByName(req.Name)
	if err == nil && existingTag.ID != id {
		return nil, errors.New("tag with this name already exists")
	}

	// Update tag
	tag.Name = req.Name

	// Save changes
	if err := uc.tagRepo.Update(tag); err != nil {
		return nil, err
	}

	response := dto.ToTagResponse(tag)
	return &response, nil
}

// DeleteTag deletes a tag
func (uc *TagUseCase) DeleteTag(id int64) error {
	return uc.tagRepo.Delete(id)
}

// ListTags lists all tags with pagination
func (uc *TagUseCase) ListTags(page, limit int) (*dto.TagListResponse, error) {
	// Set default pagination values
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get tags
	tags, err := uc.tagRepo.List(offset, limit)
	if err != nil {
		return nil, err
	}

	// Get all tags for counting total
	allTags, err := uc.tagRepo.List(0, 10000)
	if err != nil {
		return nil, err
	}
	total := int64(len(allTags))

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &dto.TagListResponse{
		Data:       dto.ToTagResponseList(tags),
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// GetTagsByUserID gets tags used by a specific user with todo counts
func (uc *TagUseCase) GetTagsByUserID(userID int64) ([]*dto.TagResponse, error) {
	tags, err := uc.tagRepo.List(0, 10000)
	if err != nil {
		return nil, err
	}

	tagStats, err := uc.todoTagRepo.GetTagStatsByUserID(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.TagResponse, len(tags))
	for i, tag := range tags {
		todoCount, exists := tagStats[tag.ID]
		if !exists {
			todoCount = 0
		}
		tagWithCount := dto.ToTagResponseWithCount(tag, todoCount)
		responses[i] = &tagWithCount
	}

	return responses, nil
}
