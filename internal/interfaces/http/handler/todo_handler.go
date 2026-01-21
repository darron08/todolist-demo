package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/darron08/todolist-demo/internal/usecase"
	"github.com/darron08/todolist-demo/pkg/dto"
	"github.com/darron08/todolist-demo/pkg/response"
)

// TodoHandler handles HTTP requests for todos
type TodoHandler struct {
	todoUseCase *usecase.TodoUseCase
}

// NewTodoHandler creates a new todo handler
func NewTodoHandler(todoUseCase *usecase.TodoUseCase) *TodoHandler {
	return &TodoHandler{
		todoUseCase: todoUseCase,
	}
}

// CreateTodo handles POST /api/v1/todos
func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var req dto.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDStr := c.GetString("UserID")
	if userIDStr == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// Convert user ID to int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	todo, createErr := h.todoUseCase.CreateTodo(userID, &req)
	if createErr != nil {
		if createErr == usecase.ErrTodoTitleRequired ||
			createErr == usecase.ErrTodoTitleTooLong ||
			createErr == usecase.ErrTodoDescriptionTooLong ||
			createErr == usecase.ErrInvalidPriority {
			response.BadRequest(c, createErr.Error())
			return
		}
		response.InternalServerError(c, "failed to create todo")
		return
	}

	response.Created(c, todo)
}

// GetTodo handles GET /api/v1/todos/:id
func (h *TodoHandler) GetTodo(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "todo id is required")
		return
	}

	// Convert id to int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid todo id")
		return
	}

	// TODO: Remove this temporary user ID when authentication is implemented
	userIDStr := c.GetString("UserID")
	if userIDStr == "" {
		userIDStr = "temp-user-1"
	}

	// Convert user ID to int64
	userID, parseErr := strconv.ParseInt(userIDStr, 10, 64)
	if parseErr != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	todo, usecaseErr := h.todoUseCase.GetTodo(id, userID)
	if usecaseErr != nil {
		if usecaseErr == usecase.ErrTodoNotFound {
			response.NotFound(c, usecaseErr.Error())
			return
		}
		if usecaseErr == usecase.ErrUnauthorized {
			response.Unauthorized(c, usecaseErr.Error())
			return
		}
		response.InternalServerError(c, "failed to get todo")
		return
	}

	response.Success(c, todo)
}

// UpdateTodo handles PUT /api/v1/todos/:id
func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "todo id is required")
		return
	}

	// Convert id to int64
	id, idErr := strconv.ParseInt(idStr, 10, 64)
	if idErr != nil {
		response.BadRequest(c, "invalid todo id")
		return
	}

	var req dto.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// TODO: Remove this temporary user ID when authentication is implemented
	userIDStr := c.GetString("UserID")
	if userIDStr == "" {
		userIDStr = "temp-user-1"
	}

	// Convert user ID to int64
	userID, userErr := strconv.ParseInt(userIDStr, 10, 64)
	if userErr != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	todo, usecaseErr := h.todoUseCase.UpdateTodo(id, userID, &req)
	if usecaseErr != nil {
		if usecaseErr == usecase.ErrTodoNotFound {
			response.NotFound(c, usecaseErr.Error())
			return
		}
		if usecaseErr == usecase.ErrUnauthorized {
			response.Unauthorized(c, usecaseErr.Error())
			return
		}
		if usecaseErr == usecase.ErrTodoTitleRequired ||
			usecaseErr == usecase.ErrTodoTitleTooLong ||
			usecaseErr == usecase.ErrTodoDescriptionTooLong ||
			usecaseErr == usecase.ErrInvalidStatus ||
			usecaseErr == usecase.ErrInvalidPriority {
			response.BadRequest(c, usecaseErr.Error())
			return
		}
		response.InternalServerError(c, "failed to update todo")
		return
	}

	response.Success(c, todo)
}

// DeleteTodo handles DELETE /api/v1/todos/:id
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "todo id is required")
		return
	}

	// Convert id to int64
	id, idErr := strconv.ParseInt(idStr, 10, 64)
	if idErr != nil {
		response.BadRequest(c, "invalid todo id")
		return
	}

	// TODO: Remove this temporary user ID when authentication is implemented
	userIDStr := c.GetString("UserID")
	if userIDStr == "" {
		userIDStr = "temp-user-1"
	}

	// Convert user ID to int64
	userID, userErr := strconv.ParseInt(userIDStr, 10, 64)
	if userErr != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	err := h.todoUseCase.DeleteTodo(id, userID)
	if err != nil {
		if err == usecase.ErrTodoNotFound {
			response.NotFound(c, err.Error())
			return
		}
		if err == usecase.ErrUnauthorized {
			response.Unauthorized(c, err.Error())
			return
		}
		response.InternalServerError(c, "failed to delete todo")
		return
	}

	response.Success(c, gin.H{"message": "todo deleted successfully"})
}

// UpdateTodoStatus handles PATCH /api/v1/todos/:id/status
func (h *TodoHandler) UpdateTodoStatus(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "todo id is required")
		return
	}

	// Convert id to int64
	id, idErr := strconv.ParseInt(idStr, 10, 64)
	if idErr != nil {
		response.BadRequest(c, "invalid todo id")
		return
	}

	var req dto.UpdateTodoStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// TODO: Remove this temporary user ID when authentication is implemented
	userIDStr := c.GetString("UserID")
	if userIDStr == "" {
		userIDStr = "temp-user-1"
	}

	// Convert user ID to int64
	userID, userErr := strconv.ParseInt(userIDStr, 10, 64)
	if userErr != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	todo, usecaseErr := h.todoUseCase.UpdateTodoStatus(id, userID, req.Status)
	if usecaseErr != nil {
		if usecaseErr == usecase.ErrTodoNotFound {
			response.NotFound(c, usecaseErr.Error())
			return
		}
		if usecaseErr == usecase.ErrUnauthorized {
			response.Unauthorized(c, usecaseErr.Error())
			return
		}
		if usecaseErr == usecase.ErrInvalidStatus {
			response.BadRequest(c, usecaseErr.Error())
			return
		}
		response.InternalServerError(c, "failed to update todo status")
		return
	}

	response.Success(c, todo)
}

// ListTodos handles GET /api/v1/todos
func (h *TodoHandler) ListTodos(c *gin.Context) {
	// TODO: Remove this temporary user ID when authentication is implemented
	userIDStr := c.GetString("UserID")
	if userIDStr == "" {
		userIDStr = "temp-user-1"
	}

	// Convert user ID to int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	var req dto.ListTodosRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	todos, usecaseErr := h.todoUseCase.ListTodos(userID, &req)
	if usecaseErr != nil {
		response.InternalServerError(c, "failed to list todos")
		return
	}

	response.SuccessWithPagination(c, todos.Data, &response.Pagination{
		Page:       todos.Page,
		Limit:      todos.Limit,
		Total:      int(todos.Total),
		TotalPages: todos.TotalPages,
	})
}
