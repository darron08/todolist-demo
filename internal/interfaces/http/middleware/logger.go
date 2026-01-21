package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Logger returns a gin middleware for logging HTTP requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Log request details
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		// Log based on status code
		if status >= 500 {
			fmt.Printf("[ERROR] request_id=%s method=%s path=%s status=%d latency=%v client_ip=%s\n",
				requestID, method, path, status, latency, clientIP)
		} else if status >= 400 {
			fmt.Printf("[WARN] request_id=%s method=%s path=%s status=%d latency=%v client_ip=%s\n",
				requestID, method, path, status, latency, clientIP)
		} else {
			fmt.Printf("[INFO] request_id=%s method=%s path=%s status=%d latency=%v client_ip=%s\n",
				requestID, method, path, status, latency, clientIP)
		}
	}
}
