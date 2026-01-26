package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"github.com/darron08/todolist-demo/internal/domain/repository"
)

var (
	ErrTagNotFound = errors.New("tag not found")
)

// TagRepositoryImpl implements repository.TagRepository interface
type TagRepositoryImpl struct {
	db *gorm.DB
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db *gorm.DB) repository.TagRepository {
	return &TagRepositoryImpl{db: db}
}

// Create creates a new tag
func (r *TagRepositoryImpl) Create(ctx context.Context, tag *entity.Tag) error {
	// Check if tag name already exists
	var existingTag entity.Tag
	result := r.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", tag.Name).First(&existingTag)
	if result.Error == nil {
		return errors.New("tag with this name already exists")
	}
	if result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}

	return r.db.WithContext(ctx).Create(tag).Error
}

// FindByID finds a tag by ID
func (r *TagRepositoryImpl) FindByID(ctx context.Context, id int64) (*entity.Tag, error) {
	var tag entity.Tag
	result := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&tag)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrTagNotFound
		}
		return nil, result.Error
	}
	return &tag, nil
}

// FindByName finds a tag by name
func (r *TagRepositoryImpl) FindByName(ctx context.Context, name string) (*entity.Tag, error) {
	var tag entity.Tag
	result := r.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name).First(&tag)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrTagNotFound
		}
		return nil, result.Error
	}
	return &tag, nil
}

// Update updates a tag
func (r *TagRepositoryImpl) Update(ctx context.Context, tag *entity.Tag) error {
	result := r.db.WithContext(ctx).Save(tag)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTagNotFound
	}
	return nil
}

// Delete soft deletes a tag
func (r *TagRepositoryImpl) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Tag{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTagNotFound
	}
	return nil
}

// List lists all tags with pagination
func (r *TagRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entity.Tag, error) {
	var tags []*entity.Tag
	result := r.db.WithContext(ctx).Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tags)

	if result.Error != nil {
		return nil, result.Error
	}
	return tags, nil
}
