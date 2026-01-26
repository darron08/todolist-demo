package handler

import (
	"strconv"

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
// @Summary Register a new user
// @Description Register a new user account with username, email, and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User registration details"
// @Success 201 {object} dto.RegisterResponse "User successfully registered"
// @Failure 400 {object} response.ErrorResponse "Invalid request format or validation error"
// @Failure 409 {object} response.ErrorResponse "Username or email already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
		return
	}

	// Register user
	user, err := h.userUseCase.Register(c.Request.Context(), &req)
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
// @Summary User login
// @Description Authenticate user and return access and refresh tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "User login credentials"
// @Success 200 {object} dto.LoginResponse "Login successful"
// @Failure 400 {object} response.ErrorResponse "Invalid request format"
// @Failure 401 {object} response.ErrorResponse "Invalid username or password"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
		return
	}

	// Login user
	loginResponse, err := h.userUseCase.Login(c.Request.Context(), &req)
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
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.RefreshTokenResponse "Token successfully refreshed"
// @Failure 400 {object} response.ErrorResponse "Invalid request format"
// @Failure 401 {object} response.ErrorResponse "Invalid or expired refresh token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
		return
	}

	// Refresh token
	refreshResponse, err := h.userUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
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
// @Summary User logout
// @Description Logout current user and invalidate the refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} response.SuccessResponse "Logout successful"
// @Failure 400 {object} response.ErrorResponse "Invalid authorization header"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	// Get user ID and token ID from context
	userIDStr := c.GetString("UserID")
	tokenID := c.GetString("TokenID")

	if userIDStr == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// Convert user ID to int64
	userID, parseErr := strconv.ParseInt(userIDStr, 10, 64)
	if parseErr != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	// For logout, we need to token ID from JWT claims
	// We'll extract it from the Authorization header and validate it
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.BadRequest(c, "authorization header is required")
		return
	}

	// Note: We should pass to actual token ID from to JWT
	// For now, we'll use a simple implementation
	err := h.userUseCase.Logout(c.Request.Context(), userID, tokenID)
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
// @Summary Get current user profile
// @Description Retrieve the profile information of the currently authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.UserResponse "User profile retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid user ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
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

	// Get user profile
	profile, err := h.userUseCase.GetProfile(c.Request.Context(), userID)
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
