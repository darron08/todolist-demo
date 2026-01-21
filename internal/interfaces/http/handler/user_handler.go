package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/darron08/todolist-demo/internal/usecase"
	"github.com/darron08/todolist-demo/pkg/dto"
	"github.com/darron08/todolist-demo/pkg/response"
)

// UserHandler handles HTTP requests for user authentication and profile
type UserHandler struct {
	userUseCase *usecase.UserUseCase
}

// NewUserHandler creates a new user handler
func NewUserHandler(userUseCase *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

// Register handles POST /api/v1/auth/register
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
		return
	}

	// Register user
	user, err := h.userUseCase.Register(&req)
	if err != nil {
		if err == usecase.ErrUsernameExists {
			response.Conflict(c, err.Error())
			return
		}
		if err == usecase.ErrEmailExists {
			response.Conflict(c, err.Error())
			return
		}
		if err == usecase.ErrInvalidUsername {
			response.BadRequest(c, err.Error())
			return
		}
		if err == usecase.ErrInvalidEmail {
			response.BadRequest(c, err.Error())
			return
		}
		if err == usecase.ErrInvalidPassword {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "failed to register user")
		return
	}

	response.Created(c, user)
}

// Login handles POST /api/v1/auth/login
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
		return
	}

	// Login user
	loginResponse, err := h.userUseCase.Login(&req)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			response.Unauthorized(c, err.Error())
			return
		}
		response.InternalServerError(c, "failed to login")
		return
	}

	response.Success(c, loginResponse)
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
		return
	}

	// Refresh token
	refreshResponse, err := h.userUseCase.RefreshToken(req.RefreshToken)
	if err != nil {
		if err == usecase.ErrInvalidToken {
			response.Unauthorized(c, err.Error())
			return
		}
		response.InternalServerError(c, "failed to refresh token")
		return
	}

	response.Success(c, refreshResponse)
}

// Logout handles POST /api/v1/auth/logout
func (h *UserHandler) Logout(c *gin.Context) {
	// Get user ID and token ID from context
	userID := c.GetString("UserID")
	tokenID := c.GetString("TokenID")

	if userID == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// For logout, we need the token ID from JWT claims
	// We'll extract it from the Authorization header and validate it
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.BadRequest(c, "authorization header is required")
		return
	}

	// Note: We should pass the actual token ID from the JWT
	// For now, we'll use a simple implementation
	err := h.userUseCase.Logout(userID, tokenID)
	if err != nil {
		if err == usecase.ErrInvalidToken {
			response.Unauthorized(c, err.Error())
			return
		}
		response.InternalServerError(c, "failed to logout")
		return
	}

	response.Success(c, gin.H{"message": "logged out successfully"})
}

// GetProfile handles GET /api/v1/users/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := c.GetString("UserID")
	if userID == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// Get user profile
	profile, err := h.userUseCase.GetProfile(userID)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "failed to get user profile")
		return
	}

	response.Success(c, profile)
}
