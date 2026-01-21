package utils

import (
	"strings"
	"unicode"
)

// Validator provides common validation functions
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateEmail validates email format
func (v *Validator) ValidateEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}

	at := strings.Index(email, "@")
	if at == -1 || at == 0 || at == len(email)-1 {
		return false
	}

	dot := strings.LastIndex(email, ".")
	if dot == -1 || dot < at || dot == len(email)-1 {
		return false
	}

	return true
}

// ValidatePassword validates password strength
func (v *Validator) ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	if len(password) > 128 {
		return ErrPasswordTooLong
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return ErrPasswordMissingUpper
	}
	if !hasLower {
		return ErrPasswordMissingLower
	}
	if !hasNumber {
		return ErrPasswordMissingNumber
	}
	if !hasSpecial {
		return ErrPasswordMissingSpecial
	}

	return nil
}

// ValidateUsername validates username format
func (v *Validator) ValidateUsername(username string) error {
	if len(username) < 3 {
		return ErrUsernameTooShort
	}

	if len(username) > 50 {
		return ErrUsernameTooLong
	}

	for _, char := range username {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) && char != '_' && char != '-' {
			return ErrUsernameInvalidFormat
		}
	}

	return nil
}

// SanitizeString sanitizes a string by removing extra whitespace
func (v *Validator) SanitizeString(input string) string {
	return strings.TrimSpace(input)
}

// Validation errors
var (
	ErrPasswordTooShort       = NewValidationError("password must be at least 8 characters long")
	ErrPasswordTooLong        = NewValidationError("password must be at most 128 characters long")
	ErrPasswordMissingUpper   = NewValidationError("password must contain at least one uppercase letter")
	ErrPasswordMissingLower   = NewValidationError("password must contain at least one lowercase letter")
	ErrPasswordMissingNumber  = NewValidationError("password must contain at least one number")
	ErrPasswordMissingSpecial = NewValidationError("password must contain at least one special character")

	ErrUsernameTooShort      = NewValidationError("username must be at least 3 characters long")
	ErrUsernameTooLong       = NewValidationError("username must be at most 50 characters long")
	ErrUsernameInvalidFormat = NewValidationError("username can only contain letters, numbers, underscores and hyphens")

	ErrInvalidEmail = NewValidationError("invalid email format")
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.Message
}
