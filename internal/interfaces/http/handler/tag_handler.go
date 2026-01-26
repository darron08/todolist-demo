package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/darron08/todolist-demo/internal/usecase"
	"github.com/darron08/todolist-demo/pkg/dto"
	"github.com/darron08/todolist-demo/pkg/response"
)

// TagHandler handles HTTP requests for tags
type TagHandler struct {
	tagUseCase *usecase.TagUseCase
}

// NewTagHandler creates a new tag handler
func NewTagHandler(tagUseCase *usecase.TagUseCase) *TagHandler {
	return &TagHandler{
		tagUseCase: tagUseCase,
	}
}

// CreateTag handles POST /api/v1/tags
// @Summary Create a new tag
// @Description Create a new tag
// @Tags Tags
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateTagRequest true "Tag details"
// @Success 201 {object} dto.TagResponse "Tag created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request format or validation error"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /tags [post]
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req dto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tag, err := h.tagUseCase.CreateTag(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrTagNameRequired ||
			err == usecase.ErrTagNameTooLong {
			response.BadRequest(c, err.Error())
			return
		}
		if err.Error() == "tag with this name already exists" {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "failed to create tag")
		return
	}

	response.Created(c, tag)
}

// GetTag handles GET /api/v1/tags/:id
// @Summary Get a tag by ID
// @Description Retrieve a specific tag by its ID
// @Tags Tags
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Tag ID"
// @Success 200 {object} dto.TagResponse "Tag retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid tag ID"
// @Failure 404 {object} response.ErrorResponse "Tag not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /tags/{id} [get]
func (h *TagHandler) GetTag(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "tag id is required")
		return
	}

	// Convert id to int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid tag id")
		return
	}

	tag, err := h.tagUseCase.GetTag(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrTagNotFound {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "failed to get tag")
		return
	}

	response.Success(c, tag)
}

// UpdateTag handles PUT /api/v1/tags/:id
// @Summary Update a tag
// @Description Update an existing tag by its ID
// @Tags Tags
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Tag ID"
// @Param request body dto.UpdateTagRequest true "Updated tag details"
// @Success 200 {object} dto.TagResponse "Tag updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request format or validation error"
// @Failure 404 {object} response.ErrorResponse "Tag not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /tags/{id} [put]
func (h *TagHandler) UpdateTag(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "tag id is required")
		return
	}

	// Convert id to int64
	id, idErr := strconv.ParseInt(idStr, 10, 64)
	if idErr != nil {
		response.BadRequest(c, "invalid tag id")
		return
	}

	var req dto.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tag, err := h.tagUseCase.UpdateTag(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrTagNotFound ||
			err == usecase.ErrTagNameRequired ||
			err == usecase.ErrTagNameTooLong {
			response.BadRequest(c, err.Error())
			return
		}
		if err.Error() == "tag with this name already exists" {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "failed to update tag")
		return
	}

	response.Success(c, tag)
}

// DeleteTag handles DELETE /api/v1/tags/:id
// @Summary Delete a tag
// @Description Delete a tag by its ID
// @Tags Tags
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Tag ID"
// @Success 200 {object} response.SuccessResponse "Tag deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid tag ID"
// @Failure 404 {object} response.ErrorResponse "Tag not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /tags/{id} [delete]
func (h *TagHandler) DeleteTag(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "tag id is required")
		return
	}

	// Convert id to int64
	id, idErr := strconv.ParseInt(idStr, 10, 64)
	if idErr != nil {
		response.BadRequest(c, "invalid tag id")
		return
	}

	err := h.tagUseCase.DeleteTag(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrTagNotFound {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "failed to delete tag")
		return
	}

	response.Success(c, gin.H{"message": "tag deleted successfully"})
}

// ListTags handles GET /api/v1/tags
// @Summary List all tags
// @Description Retrieve a paginated list of all tags
// @Tags Tags
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Success 200 {object} response.PaginatedResponse "Tags retrieved successfully"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /tags [get]
func (h *TagHandler) ListTags(c *gin.Context) {
	var req dto.ListTagsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tags, err := h.tagUseCase.ListTags(c.Request.Context(), req.Page, req.Limit)
	if err != nil {
		response.InternalServerError(c, "failed to list tags")
		return
	}

	response.SuccessWithPagination(c, tags.Data, &response.Pagination{
		Page:       tags.Page,
		Limit:      tags.Limit,
		Total:      int(tags.Total),
		TotalPages: tags.TotalPages,
	})
}

// GetUserTags handles GET /api/v1/users/my-tags
// @Summary Get current user's tags
// @Description Retrieve all tags used by the authenticated user with todo counts
// @Tags Tags
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} []dto.TagResponse "User tags retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid user ID"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /users/my-tags [get]
func (h *TagHandler) GetUserTags(c *gin.Context) {
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

	tags, err := h.tagUseCase.GetTagsByUserID(c.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(c, "failed to get user tags")
		return
	}

	response.Success(c, tags)
}
