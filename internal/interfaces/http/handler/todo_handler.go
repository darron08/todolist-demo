package handler

import (
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

	// Get user ID from context (will be set by auth middleware)
	// TODO: Remove this temporary user ID when authentication is implemented
	userID := c.GetString("UserID")
	if userID == "" {
		// For MVP without authentication, use a temporary user ID
		userID = "temp-user-1"
	}

	todo, err := h.todoUseCase.CreateTodo(userID, &req)
	if err != nil {
		if err == usecase.ErrTodoTitleRequired ||
			err == usecase.ErrTodoTitleTooLong ||
			err == usecase.ErrTodoDescriptionTooLong ||
			err == usecase.ErrInvalidPriority {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "failed to create todo")
		return
	}

	response.Created(c, todo)
}

// GetTodo handles GET /api/v1/todos/:id
func (h *TodoHandler) GetTodo(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "todo id is required")
		return
	}

	// TODO: Remove this temporary user ID when authentication is implemented
	userID := c.GetString("UserID")
	if userID == "" {
		userID = "temp-user-1"
	}

	todo, err := h.todoUseCase.GetTodo(id, userID)
	if err != nil {
		if err == usecase.ErrTodoNotFound {
			response.NotFound(c, err.Error())
			return
		}
		if err == usecase.ErrUnauthorized {
			response.Unauthorized(c, err.Error())
			return
		}
		response.InternalServerError(c, "failed to get todo")
		return
	}

	response.Success(c, todo)
}

// UpdateTodo handles PUT /api/v1/todos/:id
func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "todo id is required")
		return
	}

	var req dto.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// TODO: Remove this temporary user ID when authentication is implemented
	userID := c.GetString("UserID")
	if userID == "" {
		userID = "temp-user-1"
	}

	todo, err := h.todoUseCase.UpdateTodo(id, userID, &req)
	if err != nil {
		if err == usecase.ErrTodoNotFound {
			response.NotFound(c, err.Error())
			return
		}
		if err == usecase.ErrUnauthorized {
			response.Unauthorized(c, err.Error())
			return
		}
		if err == usecase.ErrTodoTitleRequired ||
			err == usecase.ErrTodoTitleTooLong ||
			err == usecase.ErrTodoDescriptionTooLong ||
			err == usecase.ErrInvalidStatus ||
			err == usecase.ErrInvalidPriority {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "failed to update todo")
		return
	}

	response.Success(c, todo)
}

// DeleteTodo handles DELETE /api/v1/todos/:id
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "todo id is required")
		return
	}

	// TODO: Remove this temporary user ID when authentication is implemented
	userID := c.GetString("UserID")
	if userID == "" {
		userID = "temp-user-1"
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
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "todo id is required")
		return
	}

	var req dto.UpdateTodoStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// TODO: Remove this temporary user ID when authentication is implemented
	userID := c.GetString("UserID")
	if userID == "" {
		userID = "temp-user-1"
	}

	todo, err := h.todoUseCase.UpdateTodoStatus(id, userID, req.Status)
	if err != nil {
		if err == usecase.ErrTodoNotFound {
			response.NotFound(c, err.Error())
			return
		}
		if err == usecase.ErrUnauthorized {
			response.Unauthorized(c, err.Error())
			return
		}
		if err == usecase.ErrInvalidStatus {
			response.BadRequest(c, err.Error())
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
	userID := c.GetString("UserID")
	if userID == "" {
		userID = "temp-user-1"
	}

	var req dto.ListTodosRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	todos, err := h.todoUseCase.ListTodos(userID, &req)
	if err != nil {
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
