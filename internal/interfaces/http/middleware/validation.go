package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationError represents a validation error response
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Errors []ValidationError `json:"errors"`
}

// RequestValidation returns a middleware for request validation
func RequestValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are validation errors
		if len(c.Errors) > 0 {
			validationErrors := make([]ValidationError, 0)

			for _, err := range c.Errors {
				// Check if it's a validation error
				if validationErr, ok := err.Err.(validator.ValidationErrors); ok {
					for _, e := range validationErr {
						validationErrors = append(validationErrors, ValidationError{
							Field:   e.Field(),
							Message: getValidationErrorMessage(e),
						})
					}
				}
			}

			if len(validationErrors) > 0 {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Errors: validationErrors,
				})
				c.Abort()
				return
			}
		}
	}
}

// getValidationErrorMessage returns a user-friendly error message
func getValidationErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Please enter a valid email address"
	case "min":
		return "This field must be at least " + e.Param() + " characters long"
	case "max":
		return "This field must be at most " + e.Param() + " characters long"
	case "oneof":
		return "This field must be one of: " + e.Param()
	default:
		return "Invalid value for this field"
	}
}
