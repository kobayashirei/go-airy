package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/kobayashirei/airy/internal/logger"
)

// Recovery is a middleware that recovers from panics and logs the error
// Validates: Requirements 20.2
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID
				requestID, _ := c.Get("request_id")

				// Get stack trace
				stack := string(debug.Stack())

				// Log error with context
				logger.Error("Panic recovered",
					zap.Any("request_id", requestID),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.Any("error", err),
					zap.String("stack_trace", stack),
					zap.String("client_ip", c.ClientIP()),
				)

				// Return error response
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":       "INTERNAL_ERROR",
					"message":    "Internal server error",
					"request_id": requestID,
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}

// ErrorLogger is a middleware that logs errors from handlers
// Validates: Requirements 20.2, 20.4
func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			requestID, _ := c.Get("request_id")

			for _, e := range c.Errors {
				logger.Error("Request error",
					zap.Any("request_id", requestID),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.Error(e.Err),
					zap.String("type", fmt.Sprintf("%v", e.Type)),
					zap.String("meta", fmt.Sprintf("%v", e.Meta)),
				)
			}
		}
	}
}
