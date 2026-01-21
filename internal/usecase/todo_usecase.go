package usecase

import (
	"errors"
	"math"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"github.com/darron08/todolist-demo/internal/domain/repository"
	"github.com/darron08/todolist-demo/pkg/dto"
)

var (
	ErrTodoTitleRequired      = errors.New("todo title is required")
	ErrTodoTitleTooLong       = errors.New("todo title is too long")
	ErrTodoDescriptionTooLong = errors.New("todo description is too long")
	ErrInvalidStatus          = errors.New("invalid status")
	ErrInvalidPriority        = errors.New("invalid priority")
	ErrTodoNotFound           = errors.New("todo not found")
	ErrUnauthorized           = errors.New("unauthorized")
)

// TodoUseCase implements business logic for todos
type TodoUseCase struct {
	todoRepo repository.TodoRepository
}

// NewTodoUseCase creates a new todo use case
func NewTodoUseCase(todoRepo repository.TodoRepository) *TodoUseCase {
	return &TodoUseCase{
		todoRepo: todoRepo,
	}
}

// CreateTodo creates a new todo
func (uc *TodoUseCase) CreateTodo(userID int64, req *dto.CreateTodoRequest) (*dto.TodoResponse, error) {
	// Validate title
	if req.Title == "" {
		return nil, ErrTodoTitleRequired
	}
	if len(req.Title) > 255 {
		return nil, ErrTodoTitleTooLong
	}

	// Validate description
	if len(req.Description) > 5000 {
		return nil, ErrTodoDescriptionTooLong
	}

	// Set default priority if not provided
	priority := entity.TodoPriorityMedium
	if req.Priority != "" {
		switch req.Priority {
		case "low":
			priority = entity.TodoPriorityLow
		case "medium":
			priority = entity.TodoPriorityMedium
		case "high":
			priority = entity.TodoPriorityHigh
		default:
			return nil, ErrInvalidPriority
		}
	}

	// Create todo entity
	todo := &entity.Todo{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		Status:      entity.TodoStatusNotStarted,
		Priority:    priority,
	}

	// Save to database
	if err := uc.todoRepo.Create(todo); err != nil {
		return nil, err
	}

	// Convert to response
	response := dto.ToTodoResponse(todo)
	return &response, nil
}

// GetTodo retrieves a single todo by ID
func (uc *TodoUseCase) GetTodo(id int64, userID int64) (*dto.TodoResponse, error) {
	todo, err := uc.todoRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if todo.UserID != userID {
		return nil, ErrUnauthorized
	}

	response := dto.ToTodoResponse(todo)
	return &response, nil
}

// UpdateTodo updates an existing todo
func (uc *TodoUseCase) UpdateTodo(id int64, userID int64, req *dto.UpdateTodoRequest) (*dto.TodoResponse, error) {
	// Get existing todo
	todo, err := uc.todoRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if todo.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Update fields if provided
	if req.Title != nil {
		if *req.Title == "" {
			return nil, ErrTodoTitleRequired
		}
		if len(*req.Title) > 255 {
			return nil, ErrTodoTitleTooLong
		}
		todo.Title = *req.Title
	}

	if req.Description != nil {
		if len(*req.Description) > 5000 {
			return nil, ErrTodoDescriptionTooLong
		}
		todo.Description = *req.Description
	}

	if req.DueDate != nil {
		todo.DueDate = req.DueDate
	}

	if req.Status != nil {
		switch *req.Status {
		case "not_started":
			todo.Status = entity.TodoStatusNotStarted
		case "in_progress":
			todo.Status = entity.TodoStatusInProgress
		case "completed":
			todo.Status = entity.TodoStatusCompleted
		default:
			return nil, ErrInvalidStatus
		}
	}

	if req.Priority != nil {
		switch *req.Priority {
		case "low":
			todo.Priority = entity.TodoPriorityLow
		case "medium":
			todo.Priority = entity.TodoPriorityMedium
		case "high":
			todo.Priority = entity.TodoPriorityHigh
		default:
			return nil, ErrInvalidPriority
		}
	}

	// Save changes
	if err := uc.todoRepo.Update(todo); err != nil {
		return nil, err
	}

	response := dto.ToTodoResponse(todo)
	return &response, nil
}

// DeleteTodo deletes a todo
func (uc *TodoUseCase) DeleteTodo(id int64, userID int64) error {
	// Get existing todo
	todo, err := uc.todoRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Check ownership
	if todo.UserID != userID {
		return ErrUnauthorized
	}

	// Delete todo
	return uc.todoRepo.Delete(id)
}

// UpdateTodoStatus updates the status of a todo
func (uc *TodoUseCase) UpdateTodoStatus(id int64, userID int64, status string) (*dto.TodoResponse, error) {
	// Get existing todo
	todo, err := uc.todoRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if todo.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Update status
	switch status {
	case "not_started":
		todo.Status = entity.TodoStatusNotStarted
	case "in_progress":
		todo.Status = entity.TodoStatusInProgress
	case "completed":
		todo.Status = entity.TodoStatusCompleted
	default:
		return nil, ErrInvalidStatus
	}

	// Save changes
	if err := uc.todoRepo.Update(todo); err != nil {
		return nil, err
	}

	response := dto.ToTodoResponse(todo)
	return &response, nil
}

// ListTodos lists todos with pagination and filters
func (uc *TodoUseCase) ListTodos(userID int64, req *dto.ListTodosRequest) (*dto.TodoListResponse, error) {
	// Set default pagination values
	page := req.Page
	if page < 1 {
		page = 1
	}

	limit := req.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Prepare filters
	var statusFilter *string
	if req.Status != "" {
		statusFilter = &req.Status
	}

	var priorityFilter *string
	if req.Priority != "" {
		priorityFilter = &req.Priority
	}

	// Get todos with filters
	todos, total, err := uc.todoRepo.FindByUserIDAndFilters(userID, statusFilter, priorityFilter, req.DueDateFrom, req.DueDateTo, offset, limit)
	if err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Convert to response
	return &dto.TodoListResponse{
		Data:       dto.ToTodoResponseList(todos),
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}
