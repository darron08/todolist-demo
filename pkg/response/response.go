package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents an error response structure for Swagger
type ErrorResponse struct {
	Code      int    `json:"code" example:"400"`
	Message   string `json:"message" example:"Bad request"`
	RequestID string `json:"request_id,omitempty" example:"abc-123"`
}

// SuccessResponse represents a success response structure for Swagger
type SuccessResponse struct {
	Code      int         `json:"code" example:"200"`
	Message   string      `json:"message" example:"Success"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty" example:"abc-123"`
}

// PaginatedResponse represents a paginated response structure for Swagger
type PaginatedResponse struct {
	Code       int         `json:"code" example:"200"`
	Message    string      `json:"message" example:"Success"`
	Data       interface{} `json:"data,omitempty"`
	RequestID  string      `json:"request_id,omitempty" example:"abc-123"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Response represents the standard API response structure
type Response struct {
	Code      int         `json:"code" example:"200"`
	Message   string      `json:"message" example:"Success"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty" example:"abc-123"`
}

// PaginationResponse represents a paginated response
type PaginationResponse struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination represents pagination information
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Constants for response codes
const (
	CodeSuccess             = 200
	CodeCreated             = 201
	CodeBadRequest          = 400
	CodeUnauthorized        = 401
	CodeForbidden           = 403
	CodeNotFound            = 404
	CodeConflict            = 409
	CodeInternalServerError = 500
	CodeServiceUnavailable  = 503
)

// Success returns a successful response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "Success",
		Data:    data,
	})
}

// Created returns a created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    CodeCreated,
		Message: "Created",
		Data:    data,
	})
}

// BadRequest returns a bad request response
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeBadRequest,
		Message: message,
	})
}

// Unauthorized returns an unauthorized response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: message,
	})
}

// Forbidden returns a forbidden response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: message,
	})
}

// NotFound returns a not found response
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeNotFound,
		Message: message,
	})
}

// Conflict returns a conflict response
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Code:    CodeConflict,
		Message: message,
	})
}

// InternalServerError returns an internal server error response
func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeInternalServerError,
		Message: message,
	})
}

// ServiceUnavailable returns a service unavailable response
func ServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, Response{
		Code:    CodeServiceUnavailable,
		Message: message,
	})
}

// SuccessWithPagination returns a successful response with pagination
func SuccessWithPagination(c *gin.Context, data interface{}, pagination *Pagination) {
	c.JSON(http.StatusOK, PaginationResponse{
		Code:       CodeSuccess,
		Message:    "Success",
		Data:       data,
		Pagination: pagination,
	})
}
