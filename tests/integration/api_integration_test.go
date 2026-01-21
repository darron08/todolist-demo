package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darron08/todolist-demo/internal/infrastructure/config"
	httpRouter "github.com/darron08/todolist-demo/internal/interfaces/http"
	"github.com/darron08/todolist-demo/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAPI_HealthCheck(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "test",
		},
		Swagger: config.SwaggerConfig{
			Enabled: false,
		},
	}

	router := httpRouter.SetupRouter(cfg, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestAPI_ReadyCheck(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "test",
		},
		Swagger: config.SwaggerConfig{
			Enabled: false,
		},
	}

	router := httpRouter.SetupRouter(cfg, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ready", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "ready")
}

func TestAPI_JWTTokenGeneration(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
			Issuer: "test",
		},
	}

	jwtManager := utils.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Issuer, 15*60*1000000000, 7*24*60*60*1000000000)

	token, err := jwtManager.GenerateAccessToken("user123", "testuser", "user")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAPI_JWTTokenValidation(t *testing.T) {
	jwtManager := utils.NewJWTManager("test-secret", "test", 15*60*1000000000, 7*24*60*60*1000000000)

	token, err := jwtManager.GenerateAccessToken("user123", "testuser", "user")
	assert.NoError(t, err)

	claims, err := jwtManager.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user", claims.Role)
}

func TestAPI_ResponseFormat(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "test",
		},
		Swagger: config.SwaggerConfig{
			Enabled: false,
		},
	}

	router := httpRouter.SetupRouter(cfg, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), `"status"`)
	assert.Contains(t, w.Body.String(), `"healthy"`)
}

func TestAPI_BasicRouting(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "test",
		},
		Swagger: config.SwaggerConfig{
			Enabled: false,
		},
		JWT: config.JWTConfig{
			Secret: "test-secret",
			Issuer: "test",
		},
	}

	router := httpRouter.SetupRouter(cfg, nil, nil, nil, nil)

	routes := router.Routes()

	// Check that auth routes are registered
	authRoutes := 0
	healthRoutes := 0
	readyRoutes := 0

	for _, route := range routes {
		switch route.Path {
		case "/health":
			healthRoutes++
		case "/ready":
			readyRoutes++
		case "/api/v1/auth/register":
		case "/api/v1/auth/login":
		case "/api/v1/auth/refresh":
		case "/api/v1/auth/logout":
			authRoutes++
		case "/api/v1/todos":
			// Todo routes exist
		}
	}

	assert.Greater(t, healthRoutes, 0, "Health route should be registered")
	assert.Greater(t, readyRoutes, 0, "Ready route should be registered")
	assert.Greater(t, authRoutes, 0, "Auth routes should be registered")
	assert.NotEqual(t, 0, len(routes), "Routes should be registered")
}
