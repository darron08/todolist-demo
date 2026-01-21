package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/darron08/todolist-demo/internal/usecase"
	"github.com/darron08/todolist-demo/pkg/dto"
	"github.com/darron08/todolist-demo/pkg/response"
)

// AdminHandler handles HTTP requests for admin operations
type AdminHandler struct {
	adminUseCase *usecase.AdminUseCase
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminUseCase *usecase.AdminUseCase) *AdminHandler {
	return &AdminHandler{
		adminUseCase: adminUseCase,
	}
}

// CreateUser handles POST /api/v1/admin/users (admin only)
// @Summary Create a user (Admin)
// @Description Create a new user account (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.AdminCreateUserRequest true "User creation details"
// @Success 201 {object} dto.UserResponse "User created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request format or validation error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - admin role required"
// @Failure 409 {object} response.ErrorResponse "Username or email already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users [post]
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req dto.AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userIDStr := c.GetString("UserID")
	if userIDStr == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	user, createErr := h.adminUseCase.CreateUser(userID, &req)
	if createErr != nil {
		if createErr == usecase.ErrUsernameExists || createErr == usecase.ErrEmailExists {
			response.Conflict(c, createErr.Error())
			return
		}
		if createErr == usecase.ErrInvalidUsername || createErr == usecase.ErrInvalidEmail || createErr == usecase.ErrInvalidPassword {
			response.BadRequest(c, createErr.Error())
			return
		}
		response.InternalServerError(c, "failed to create user")
		return
	}

	response.Created(c, user)
}

// ListAllUsers handles GET /api/v1/admin/users (admin only)
// @Summary List all users (Admin)
// @Description Retrieve a paginated list of all users (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Success 200 {object} response.PaginatedResponse "Users retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - admin role required"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users [get]
func (h *AdminHandler) ListAllUsers(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")

	pageInt := 1
	if page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageInt = p
		}
	}

	limitInt := 20
	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			limitInt = l
		}
	}

	offset := (pageInt - 1) * limitInt

	users, err := h.adminUseCase.ListAllUsers(offset, limitInt)
	if err != nil {
		response.InternalServerError(c, "failed to list users")
		return
	}

	response.SuccessWithPagination(c, users.Data, &response.Pagination{
		Page:       users.Page,
		Limit:      users.Limit,
		Total:      int(users.Total),
		TotalPages: users.TotalPages,
	})
}

// GetUser handles GET /api/v1/admin/users/:id (admin only)
// @Summary Get user by ID (Admin)
// @Description Retrieve a specific user by ID (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "User ID"
// @Success 200 {object} dto.UserResponse "User retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid user ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - admin role required"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{id} [get]
func (h *AdminHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "user id is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	user, usecaseErr := h.adminUseCase.GetUser(id)
	if usecaseErr != nil {
		response.NotFound(c, usecaseErr.Error())
		return
	}

	response.Success(c, user)
}

// DeleteUser handles DELETE /api/v1/admin/users/:id (admin only)
// @Summary Delete user (Admin)
// @Description Delete a user by ID (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "User ID"
// @Success 200 {object} response.SuccessResponse "User deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid user ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - admin role required"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "user id is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := h.adminUseCase.DeleteUser(id); err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "user deleted successfully"})
}

// ListAllTodos handles GET /api/v1/admin/todos (admin only)
// @Summary List all todos (Admin)
// @Description Retrieve a paginated list of all todos with optional filters (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param status query string false "Filter by status" Enums(not_started, in_progress, completed)
// @Param priority query string false "Filter by priority" Enums(low, medium, high)
// @Success 200 {object} response.PaginatedResponse "Todos retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - admin role required"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/todos [get]
func (h *AdminHandler) ListAllTodos(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	status := c.Query("status")
	priority := c.Query("priority")

	pageInt := 1
	if page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageInt = p
		}
	}

	limitInt := 20
	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			limitInt = l
		}
	}

	var statusFilter *string
	if status != "" {
		statusFilter = &status
	}

	var priorityFilter *string
	if priority != "" {
		priorityFilter = &priority
	}

	todos, err := h.adminUseCase.ListAllTodos(pageInt, limitInt, statusFilter, priorityFilter)
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

// DeleteAnyTodo handles DELETE /api/v1/admin/todos/:id (admin only)
// @Summary Delete any todo (Admin)
// @Description Delete any todo by ID, regardless of ownership (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Todo ID"
// @Success 200 {object} response.SuccessResponse "Todo deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid todo ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - admin role required"
// @Failure 404 {object} response.ErrorResponse "Todo not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/todos/{id} [delete]
func (h *AdminHandler) DeleteAnyTodo(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "todo id is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid todo id")
		return
	}

	if err := h.adminUseCase.DeleteAnyTodo(id); err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "todo deleted successfully"})
}
