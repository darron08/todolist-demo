package repository

import (
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
func (r *TagRepositoryImpl) Create(tag *entity.Tag) error {
	// Check if tag name already exists
	var existingTag entity.Tag
	result := r.db.Where("name = ? AND deleted_at IS NULL", tag.Name).First(&existingTag)
	if result.Error == nil {
		return errors.New("tag with this name already exists")
	}
	if result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}

	return r.db.Create(tag).Error
}

// FindByID finds a tag by ID
func (r *TagRepositoryImpl) FindByID(id string) (*entity.Tag, error) {
	var tag entity.Tag
	result := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&tag)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrTagNotFound
		}
		return nil, result.Error
	}
	return &tag, nil
}

// FindByName finds a tag by name
func (r *TagRepositoryImpl) FindByName(name string) (*entity.Tag, error) {
	var tag entity.Tag
	result := r.db.Where("name = ? AND deleted_at IS NULL", name).First(&tag)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrTagNotFound
		}
		return nil, result.Error
	}
	return &tag, nil
}

// Update updates a tag
func (r *TagRepositoryImpl) Update(tag *entity.Tag) error {
	result := r.db.Save(tag)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTagNotFound
	}
	return nil
}

// Delete soft deletes a tag
func (r *TagRepositoryImpl) Delete(id string) error {
	result := r.db.Where("id = ?", id).Delete(&entity.Tag{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTagNotFound
	}
	return nil
}

// List lists all tags with pagination
func (r *TagRepositoryImpl) List(offset, limit int) ([]*entity.Tag, error) {
	var tags []*entity.Tag
	result := r.db.Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tags)

	if result.Error != nil {
		return nil, result.Error
	}
	return tags, nil
}

// FindByUserID finds tags associated with a user
func (r *TagRepositoryImpl) FindByUserID(userID string) ([]*entity.Tag, error) {
	var tags []*entity.Tag

	result := r.db.Table("tags").
		Select("tags.*").
		Joins("INNER JOIN user_tags ON user_tags.tag_id = tags.id").
		Where("user_tags.user_id = ? AND tags.deleted_at IS NULL", userID).
		Order("tags.created_at DESC").
		Find(&tags)

	if result.Error != nil {
		return nil, result.Error
	}
	return tags, nil
}
