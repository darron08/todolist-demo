package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposedHeaders:   []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORS returns a CORS middleware
func CORS(cfg *CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowedOrigin := ""
		for _, allowed := range cfg.AllowedOrigins {
			if allowed == "*" || allowed == origin {
				allowedOrigin = allowed
				break
			}
		}

		if allowedOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", joinHeaders(cfg.AllowedHeaders))
		c.Writer.Header().Set("Access-Control-Allow-Methods", joinHeaders(cfg.AllowedMethods))
		c.Writer.Header().Set("Access-Control-Expose-Headers", joinHeaders(cfg.ExposedHeaders))
		c.Writer.Header().Set("Access-Control-Max-Age", string(cfg.MaxAge))

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func joinHeaders(headers []string) string {
	result := ""
	for i, header := range headers {
		if i > 0 {
			result += ", "
		}
		result += header
	}
	return result
}
