package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/security"
)

// SecurityConfig holds configuration for security middleware
type SecurityConfig struct {
	// EnableXSSProtection enables XSS protection headers
	EnableXSSProtection bool
	// EnableContentTypeNosniff enables X-Content-Type-Options: nosniff
	EnableContentTypeNosniff bool
	// EnableFrameDeny enables X-Frame-Options: DENY
	EnableFrameDeny bool
	// EnableHSTS enables HTTP Strict Transport Security
	EnableHSTS bool
	// HSTSMaxAge is the max-age for HSTS in seconds
	HSTSMaxAge int
	// EnableInputValidation enables input validation middleware
	EnableInputValidation bool
	// MaxBodySize is the maximum request body size in bytes
	MaxBodySize int64
}

// DefaultSecurityConfig returns the default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		EnableXSSProtection:      true,
		EnableContentTypeNosniff: true,
		EnableFrameDeny:          true,
		EnableHSTS:               true,
		HSTSMaxAge:               31536000, // 1 year
		EnableInputValidation:    true,
		MaxBodySize:              10 * 1024 * 1024, // 10MB
	}
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders(config SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// XSS Protection
		if config.EnableXSSProtection {
			c.Header("X-XSS-Protection", "1; mode=block")
		}

		// Content Type Options
		if config.EnableContentTypeNosniff {
			c.Header("X-Content-Type-Options", "nosniff")
		}

		// Frame Options
		if config.EnableFrameDeny {
			c.Header("X-Frame-Options", "DENY")
		}

		// HSTS
		if config.EnableHSTS {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// InputValidation validates and sanitizes request inputs
func InputValidation() gin.HandlerFunc {
	validator := security.NewValidator()

	return func(c *gin.Context) {
		// Validate query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if err := validator.CheckSQLInjection(value); err != nil {
					response.BadRequest(c, "Invalid query parameter: "+key, "potential injection detected")
					c.Abort()
					return
				}
				if err := validator.CheckXSS(value); err != nil {
					response.BadRequest(c, "Invalid query parameter: "+key, "potential XSS detected")
					c.Abort()
					return
				}
			}
		}

		// For POST/PUT/PATCH requests, validate JSON body
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/json") {
				// Read body
				body, err := io.ReadAll(c.Request.Body)
				if err != nil {
					response.BadRequest(c, "Failed to read request body", nil)
					c.Abort()
					return
				}

				// Restore body for later use
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

				// Parse JSON and validate string fields
				var jsonData map[string]interface{}
				if err := json.Unmarshal(body, &jsonData); err == nil {
					if err := validateJSONFields(jsonData, validator); err != nil {
						response.BadRequest(c, "Invalid request body", err.Error())
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}

// validateJSONFields recursively validates string fields in JSON data
func validateJSONFields(data map[string]interface{}, validator *security.Validator) error {
	for _, value := range data {
		switch v := value.(type) {
		case string:
			if err := validator.CheckSQLInjection(v); err != nil {
				return err
			}
			if err := validator.CheckXSS(v); err != nil {
				return err
			}
		case map[string]interface{}:
			if err := validateJSONFields(v, validator); err != nil {
				return err
			}
		case []interface{}:
			for _, item := range v {
				if str, ok := item.(string); ok {
					if err := validator.CheckSQLInjection(str); err != nil {
						return err
					}
					if err := validator.CheckXSS(str); err != nil {
						return err
					}
				} else if nested, ok := item.(map[string]interface{}); ok {
					if err := validateJSONFields(nested, validator); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// MaxBodySize limits the request body size
func MaxBodySize(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}
