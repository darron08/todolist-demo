package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/darron08/todolist-demo/pkg/utils"
)

var (
	ErrMissingAuthHeader = errors.New("authorization header is missing")
	ErrInvalidAuthHeader = errors.New("authorization header format is invalid")
	ErrInvalidToken      = errors.New("invalid or expired token")
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "authorization header is missing",
			})
			c.Abort()
			return
		}

		// Check Bearer format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "authorization header format is invalid",
			})
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid or expired token",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("UserID", fmt.Sprintf("%d", claims.UserID))
		c.Set("Username", claims.Username)
		c.Set("Role", claims.Role)

		c.Next()
	}
}

// OptionalAuthMiddleware optionally validates JWT tokens
func OptionalAuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header provided, continue without setting user context
			c.Next()
			return
		}

		// Check Bearer format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "authorization header format is invalid",
			})
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			// Invalid token, but don't fail the request
			c.Next()
			return
		}

		// Set user context if token is valid
		c.Set("UserID", fmt.Sprintf("%d", claims.UserID))
		c.Set("Username", claims.Username)
		c.Set("Role", claims.Role)

		c.Next()
	}
}

// RequireRole checks if user has required role
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context
		role, exists := c.Get("Role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "user not authenticated",
			})
			c.Abort()
			return
		}

		// Check if user has required role
		userRole, ok := role.(string)
		if !ok || userRole != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
