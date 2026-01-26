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
// @Summary Create a new todo
// @Description Create a new todo item for the authenticated user
// @Tags Todos
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateTodoRequest true "Todo details"
// @Success 201 {object} dto.TodoResponse "Todo created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request format or validation error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /todos [post]
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

	todo, createErr := h.todoUseCase.CreateTodo(c.Request.Context(), userID, &req)
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
// @Summary Get a todo by ID
// @Description Retrieve a specific todo item by its ID (only own todos)
// @Tags Todos
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Todo ID"
// @Success 200 {object} dto.TodoResponse "Todo retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid todo ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Todo not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /todos/{id} [get]
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

	todo, usecaseErr := h.todoUseCase.GetTodo(c.Request.Context(), id, userID)
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
// @Summary Update a todo
// @Description Update an existing todo item by its ID (only own todos)
// @Tags Todos
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Todo ID"
// @Param request body dto.UpdateTodoRequest true "Updated todo details"
// @Success 200 {object} dto.TodoResponse "Todo updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request format or validation error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Todo not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /todos/{id} [put]
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

	todo, usecaseErr := h.todoUseCase.UpdateTodo(c.Request.Context(), id, userID, &req)
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
// @Summary Delete a todo
// @Description Delete a todo item by its ID (only own todos)
// @Tags Todos
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Todo ID"
// @Success 200 {object} response.SuccessResponse "Todo deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid todo ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Todo not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /todos/{id} [delete]
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

	err := h.todoUseCase.DeleteTodo(c.Request.Context(), id, userID)
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
// @Summary Update todo status
// @Description Update the status of a todo item by its ID (only own todos)
// @Tags Todos
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Todo ID"
// @Param request body dto.UpdateTodoStatusRequest true "New status"
// @Success 200 {object} dto.TodoResponse "Todo status updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request format or validation error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Todo not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /todos/{id}/status [patch]
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

	todo, usecaseErr := h.todoUseCase.UpdateTodoStatus(c.Request.Context(), id, userID, req.Status)
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
// @Summary List todos
// @Description Retrieve a paginated list of todos for the authenticated user with optional filters and sorting
// @Tags Todos
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param status query string false "Filter by status" Enums(not_started, in_progress, completed)
// @Param priority query string false "Filter by priority" Enums(low, medium, high)
// @Param search query string false "Search in title and description" maxlength(100)
// @Param due_date_from query string false "Filter todos due after this date (RFC3339 format)" format(date-time)
// @Param due_date_to query string false "Filter todos due before this date (RFC3339 format)" format(date-time)
// @Param sort_by query string false "Sort field" Enums(due_date, status, title) default(due_date)
// @Param sort_order query string false "Sort order" Enums(asc, desc) default(asc)
// @Success 200 {object} response.PaginatedResponse "Todos retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid user ID or request format"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /todos [get]
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

	todos, usecaseErr := h.todoUseCase.ListTodos(c.Request.Context(), userID, &req)
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
