package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Response is the standard API response structure
type Response struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorResponse is the error response structure
type ErrorResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp time.Time   `json:"timestamp"`
}

// Success returns a successful response
func Success(c *gin.Context, data interface{}) {
	requestID, _ := c.Get("request_id")
	c.JSON(200, Response{
		Code:      "SUCCESS",
		Message:   "Success",
		Data:      data,
		RequestID: requestID.(string),
		Timestamp: time.Now(),
	})
}

// Error returns an error response
func Error(c *gin.Context, statusCode int, code string, message string, details interface{}) {
	requestID, _ := c.Get("request_id")
	c.JSON(statusCode, ErrorResponse{
		Code:      code,
		Message:   message,
		Details:   details,
		RequestID: requestID.(string),
		Timestamp: time.Now(),
	})
}

// BadRequest returns a 400 error
func BadRequest(c *gin.Context, message string, details interface{}) {
	Error(c, 400, "BAD_REQUEST", message, details)
}

// Unauthorized returns a 401 error
func Unauthorized(c *gin.Context, message string) {
	Error(c, 401, "UNAUTHORIZED", message, nil)
}

// Forbidden returns a 403 error
func Forbidden(c *gin.Context, message string) {
	Error(c, 403, "FORBIDDEN", message, nil)
}

// NotFound returns a 404 error
func NotFound(c *gin.Context, message string) {
	Error(c, 404, "NOT_FOUND", message, nil)
}

// Conflict returns a 409 error
func Conflict(c *gin.Context, message string) {
	Error(c, 409, "CONFLICT", message, nil)
}

// InternalError returns a 500 error
func InternalError(c *gin.Context, message string) {
	Error(c, 500, "INTERNAL_ERROR", message, nil)
}
