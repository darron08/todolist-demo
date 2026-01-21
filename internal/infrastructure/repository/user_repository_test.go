package repository

import (
	"testing"
	"time"

	"github.com/darron08/todolist-demo/internal/domain/entity"
	"github.com/darron08/todolist-demo/internal/infrastructure/database/mysql"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *mysql.DB {
	mysqlConfig := &mysql.DBConfig{
		Host:            "localhost",
		Port:            "3307",
		Username:        "todolist_test_user",
		Password:        "todolist_test_password",
		Database:        "todolist_test",
		Charset:         "utf8mb4",
		ParseTime:       true,
		Loc:             "Local",
		MaxIdleConns:    5,
		MaxOpenConns:    20,
		ConnMaxLifetime: time.Hour,
		LogMode:         logger.Silent,
	}

	db, err := mysql.NewConnection(mysqlConfig)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Clean database
	gormDB := db.GetDB()
	gormDB.Exec("DELETE FROM user_tags")
	gormDB.Exec("DELETE FROM tags")
	gormDB.Exec("DELETE FROM todos")
	gormDB.Exec("DELETE FROM users")

	return db
}

func cleanTestDB(t *testing.T) {
	db := setupTestDB(t)
	gormDB := db.GetDB()
	gormDB.Exec("DELETE FROM user_tags")
	gormDB.Exec("DELETE FROM tags")
	gormDB.Exec("DELETE FROM todos")
	gormDB.Exec("DELETE FROM users")
	db.Close()
}

func createTestUser(t *testing.T, db *mysql.DB) *entity.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testPassword123"), bcrypt.DefaultCost)

	user := &entity.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Role:         entity.UserRoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	userRepo := NewUserRepository(db.GetDB())
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanTestDB(t)

	userRepo := NewUserRepository(db.GetDB())

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &entity.User{
		Username:     "newuser",
		Email:        "newuser@example.com",
		PasswordHash: string(hashedPassword),
		Role:         entity.UserRoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := userRepo.Create(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID)
}

func TestUserRepository_Create_DuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	defer cleanTestDB(t)

	userRepo := NewUserRepository(db.GetDB())

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user1 := &entity.User{
		Username:     "duplicateuser",
		Email:        "user1@example.com",
		PasswordHash: string(hashedPassword),
		Role:         entity.UserRoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := userRepo.Create(user1)
	assert.NoError(t, err)

	user2 := &entity.User{
		Username:     "duplicateuser",
		Email:        "user2@example.com",
		PasswordHash: string(hashedPassword),
		Role:         entity.UserRoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = userRepo.Create(user2)
	assert.Error(t, err)
	assert.Equal(t, "user already exists", err.Error())
}

func TestUserRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanTestDB(t)

	user := createTestUser(t, db)
	userRepo := NewUserRepository(db.GetDB())

	foundUser, err := userRepo.FindByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, user.Username, foundUser.Username)
	assert.Equal(t, user.Email, foundUser.Email)
}

func TestUserRepository_FindByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer cleanTestDB(t)

	user := createTestUser(t, db)
	userRepo := NewUserRepository(db.GetDB())

	foundUser, err := userRepo.FindByUsername(user.Username)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, user.Username, foundUser.Username)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer cleanTestDB(t)

	user := createTestUser(t, db)
	userRepo := NewUserRepository(db.GetDB())

	user.Email = "updated@example.com"
	err := userRepo.Update(user)
	assert.NoError(t, err)

	updatedUser, err := userRepo.FindByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated@example.com", updatedUser.Email)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer cleanTestDB(t)

	user := createTestUser(t, db)
	userRepo := NewUserRepository(db.GetDB())

	err := userRepo.Delete(user.ID)
	assert.NoError(t, err)

	_, err = userRepo.FindByID(user.ID)
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer cleanTestDB(t)

	userRepo := NewUserRepository(db.GetDB())

	for i := 0; i < 5; i++ {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		user := &entity.User{
			Username:     "user" + string(rune('a'+i)),
			Email:        "user" + string(rune('a'+i)) + "@example.com",
			PasswordHash: string(hashedPassword),
			Role:         entity.UserRoleUser,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err := userRepo.Create(user)
		assert.NoError(t, err)
	}

	users, err := userRepo.List(0, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 5)
}
