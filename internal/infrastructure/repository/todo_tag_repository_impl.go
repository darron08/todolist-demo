package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"github.com/darron08/todolist-demo/internal/domain/repository"
)

// TodoTagRepositoryImpl implements repository.TodoTagRepository interface
type TodoTagRepositoryImpl struct {
	db *gorm.DB
}

// NewTodoTagRepository creates a new todo-tag repository
func NewTodoTagRepository(db *gorm.DB) repository.TodoTagRepository {
	return &TodoTagRepositoryImpl{db: db}
}

// AddTagsToTodo adds tags to a todo
func (r *TodoTagRepositoryImpl) AddTagsToTodo(ctx context.Context, todoID int64, tagIDs []int64) error {
	if len(tagIDs) == 0 {
		return nil
	}

	var todoTags []entity.TodoTag
	for _, tagID := range tagIDs {
		todoTags = append(todoTags, entity.TodoTag{
			TodoID: todoID,
			TagID:  tagID,
		})
	}

	return r.db.WithContext(ctx).Create(&todoTags).Error
}

// RemoveTagsFromTodo removes tags from a todo
func (r *TodoTagRepositoryImpl) RemoveTagsFromTodo(ctx context.Context, todoID int64, tagIDs []int64) error {
	if len(tagIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Where("todo_id = ? AND tag_id IN ?", todoID, tagIDs).
		Delete(&entity.TodoTag{}).
		Error
}

// ReplaceTagsForTodo replaces all tags for a todo
func (r *TodoTagRepositoryImpl) ReplaceTagsForTodo(ctx context.Context, todoID int64, tagIDs []int64) error {
	tx := r.db.WithContext(ctx).Begin()

	if err := tx.Where("todo_id = ?", todoID).Delete(&entity.TodoTag{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(tagIDs) > 0 {
		var todoTags []entity.TodoTag
		for _, tagID := range tagIDs {
			todoTags = append(todoTags, entity.TodoTag{
				TodoID: todoID,
				TagID:  tagID,
			})
		}
		if err := tx.Create(&todoTags).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetTagsByTodoID gets all tags for a todo
func (r *TodoTagRepositoryImpl) GetTagsByTodoID(ctx context.Context, todoID int64) ([]*entity.Tag, error) {
	var tags []*entity.Tag

	result := r.db.WithContext(ctx).Table("tags").
		Select("tags.*").
		Joins("INNER JOIN todo_tags ON todo_tags.tag_id = tags.id").
		Where("todo_tags.todo_id = ? AND tags.deleted_at IS NULL", todoID).
		Order("tags.name ASC").
		Find(&tags)

	if result.Error != nil {
		return nil, result.Error
	}
	return tags, nil
}

// GetTodosByTagID gets all todos for a tag
func (r *TodoTagRepositoryImpl) GetTodosByTagID(ctx context.Context, tagID int64, offset, limit int) ([]*entity.Todo, int64, error) {
	var todos []*entity.Todo
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Todo{}).
		Joins("INNER JOIN todo_tags ON todo_tags.todo_id = todos.id").
		Where("todo_tags.tag_id = ? AND todos.deleted_at IS NULL", tagID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	result := query.Order("todos.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&todos)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return todos, total, nil
}

// GetTagStatsByUserID gets tag statistics for a user
func (r *TodoTagRepositoryImpl) GetTagStatsByUserID(ctx context.Context, userID int64) (map[int64]int64, error) {
	type TagStat struct {
		TagID int64
		Count int64
	}

	var stats []TagStat

	result := r.db.WithContext(ctx).Table("todo_tags").
		Select("tag_id, COUNT(*) as count").
		Joins("INNER JOIN todos ON todos.id = todo_tags.todo_id").
		Where("todos.user_id = ?", userID).
		Group("tag_id").
		Find(&stats)

	if result.Error != nil {
		return nil, result.Error
	}

	tagStats := make(map[int64]int64)
	for _, stat := range stats {
		tagStats[stat.TagID] = stat.Count
	}

	return tagStats, nil
}
