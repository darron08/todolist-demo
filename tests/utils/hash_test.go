package utils

import (
	"testing"

	"github.com/darron08/todolist-demo/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "testPassword123"
	hash, err := utils.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
	assert.True(t, len(hash) > 50)
}

func TestVerifyPassword_Success(t *testing.T) {
	password := "testPassword123"
	hash, _ := utils.HashPassword(password)

	isValid := utils.VerifyPassword(password, hash)
	assert.True(t, isValid)
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	password1 := "testPassword123"
	password2 := "wrongPassword456"
	hash, _ := utils.HashPassword(password1)

	isValid := utils.VerifyPassword(password2, hash)
	assert.False(t, isValid)
}

func TestHash_DifferentResults(t *testing.T) {
	password := "testPassword123"

	hash1, _ := utils.HashPassword(password)
	hash2, _ := utils.HashPassword(password)

	assert.NotEqual(t, hash1, hash2, "same password should generate different hashes due to salt")
}
