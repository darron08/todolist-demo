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
