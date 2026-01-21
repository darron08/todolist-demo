package utils

import (
	"testing"

	"github.com/darron08/todolist-demo/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidateEmail_Success(t *testing.T) {
	validator := utils.NewValidator()

	validEmails := []string{
		"user@example.com",
		"test.user@example.com",
		"user123@test-domain.co.uk",
	}

	for _, email := range validEmails {
		isValid := validator.ValidateEmail(email)
		assert.True(t, isValid, "Email should be valid: "+email)
	}
}

func TestValidateEmail_Invalid(t *testing.T) {
	validator := utils.NewValidator()

	invalidEmails := []string{
		"",
		"invalid",
		"@example.com",
		"user@",
		"user@example",
		"user@. com",
	}

	for _, email := range invalidEmails {
		isValid := validator.ValidateEmail(email)
		assert.False(t, isValid, "Email should be invalid: "+email)
	}
}

func TestValidatePassword_Strong(t *testing.T) {
	validator := utils.NewValidator()

	strongPasswords := []string{
		"StrongPass123!",
		"MyP@ssw0rd",
		"V3ry$ecur3#2024",
	}

	for _, password := range strongPasswords {
		err := validator.ValidatePassword(password)
		assert.NoError(t, err, "Password should be valid: "+password)
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	validator := utils.NewValidator()

	shortPasswords := []string{
		"Short1!",
		"Pass1",
		"Pw1!",
		"",
	}

	for _, password := range shortPasswords {
		err := validator.ValidatePassword(password)
		assert.Error(t, err, "Password should be too short: "+password)
		assert.Equal(t, "password must be at least 8 characters long", err.Error())
	}
}

func TestValidatePassword_MissingUpper(t *testing.T) {
	validator := utils.NewValidator()

	password := "lowercase123!"
	err := validator.ValidatePassword(password)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uppercase")
}

func TestValidatePassword_MissingLower(t *testing.T) {
	validator := utils.NewValidator()

	password := "UPPERCASE123!"
	err := validator.ValidatePassword(password)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lowercase")
}

func TestValidatePassword_MissingNumber(t *testing.T) {
	validator := utils.NewValidator()

	password := "NoNumbers!"
	err := validator.ValidatePassword(password)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "number")
}

func TestValidatePassword_MissingSpecial(t *testing.T) {
	validator := utils.NewValidator()

	password := "NoSpecialChars123"
	err := validator.ValidatePassword(password)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "special")
}

func TestValidateUsername_Valid(t *testing.T) {
	validator := utils.NewValidator()

	validUsernames := []string{
		"testuser",
		"test_user123",
		"test-user-2024",
		"user123",
	}

	for _, username := range validUsernames {
		err := validator.ValidateUsername(username)
		assert.NoError(t, err, "Username should be valid: "+username)
	}
}

func TestValidateUsername_TooShort(t *testing.T) {
	validator := utils.NewValidator()

	shortUsernames := []string{
		"ab",
		"a",
		"",
	}

	for _, username := range shortUsernames {
		err := validator.ValidateUsername(username)
		assert.Error(t, err, "Username should be too short: "+username)
		assert.Equal(t, "username must be at least 3 characters long", err.Error())
	}
}

func TestValidateUsername_TooLong(t *testing.T) {
	validator := utils.NewValidator()

	longUsername := "this_is_a_very_long_username_that_exceeds_50_characters"
	err := validator.ValidateUsername(longUsername)
	assert.Error(t, err)
	assert.Equal(t, "username must be at most 50 characters long", err.Error())
}

func TestValidateUsername_InvalidFormat(t *testing.T) {
	validator := utils.NewValidator()

	invalidUsernames := []string{
		"user@domain",
		"user name",
		"user#123",
		"user.test",
	}

	for _, username := range invalidUsernames {
		err := validator.ValidateUsername(username)
		assert.Error(t, err, "Username should be invalid: "+username)
	}
}
