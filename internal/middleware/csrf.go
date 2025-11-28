package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/security"
)

// CSRFMiddleware provides CSRF protection for state-changing requests
func CSRFMiddleware(config security.CSRFConfig) gin.HandlerFunc {
	manager := security.NewCSRFManager(config)

	return func(c *gin.Context) {
		// Skip CSRF check for safe methods (GET, HEAD, OPTIONS, TRACE)
		if isSafeMethod(c.Request.Method) {
			// For safe methods, generate a new token if not present
			cookieToken, err := c.Cookie(config.CookieName)
			if err != nil || cookieToken == "" {
				token, err := manager.GenerateToken()
				if err != nil {
					response.InternalError(c, "Failed to generate CSRF token")
					c.Abort()
					return
				}
				setCSRFCookie(c, config, token.Token)
			}
			// Set token in context for templates
			c.Set("csrf_token", cookieToken)
			c.Next()
			return
		}

		// For state-changing methods, validate the token
		cookieToken, err := c.Cookie(config.CookieName)
		if err != nil || cookieToken == "" {
			response.Error(c, http.StatusForbidden, "CSRF_ERROR", "Missing CSRF token", nil)
			c.Abort()
			return
		}

		// Get token from header or form
		requestToken := c.GetHeader(config.HeaderName)
		if requestToken == "" {
			requestToken = c.PostForm(config.FormFieldName)
		}

		// Validate token pair
		if err := manager.ValidateTokenPair(cookieToken, requestToken); err != nil {
			response.Error(c, http.StatusForbidden, "CSRF_ERROR", "Invalid CSRF token", nil)
			c.Abort()
			return
		}

		// Generate new token for next request (token rotation)
		newToken, err := manager.GenerateToken()
		if err != nil {
			response.InternalError(c, "Failed to generate CSRF token")
			c.Abort()
			return
		}

		// Revoke old token
		manager.RevokeToken(cookieToken)

		// Set new token in cookie
		setCSRFCookie(c, config, newToken.Token)

		// Set token in context
		c.Set("csrf_token", newToken.Token)

		c.Next()
	}
}

// CSRFMiddlewareWithSkipPaths creates CSRF middleware that skips certain paths
func CSRFMiddlewareWithSkipPaths(config security.CSRFConfig, skipPaths []string) gin.HandlerFunc {
	csrfMiddleware := CSRFMiddleware(config)

	return func(c *gin.Context) {
		// Check if path should be skipped
		path := c.Request.URL.Path
		for _, skipPath := range skipPaths {
			if strings.HasPrefix(path, skipPath) {
				c.Next()
				return
			}
		}

		csrfMiddleware(c)
	}
}

// CSRFTokenHandler returns the current CSRF token
// GET /api/v1/csrf-token
func CSRFTokenHandler(config security.CSRFConfig) gin.HandlerFunc {
	manager := security.NewCSRFManager(config)

	return func(c *gin.Context) {
		// Generate new token
		token, err := manager.GenerateToken()
		if err != nil {
			response.InternalError(c, "Failed to generate CSRF token")
			return
		}

		// Set cookie
		setCSRFCookie(c, config, token.Token)

		response.Success(c, gin.H{
			"csrf_token": token.Token,
		})
	}
}

// isSafeMethod returns true if the HTTP method is considered safe
func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// setCSRFCookie sets the CSRF token cookie
func setCSRFCookie(c *gin.Context, config security.CSRFConfig, token string) {
	sameSite := http.SameSiteStrictMode
	switch config.SameSite {
	case "Lax":
		sameSite = http.SameSiteLaxMode
	case "None":
		sameSite = http.SameSiteNoneMode
	}

	c.SetSameSite(sameSite)
	c.SetCookie(
		config.CookieName,
		token,
		int(config.TokenExpiration.Seconds()),
		"/",
		"",
		config.Secure,
		true, // HttpOnly
	)
}

// GetCSRFToken retrieves the CSRF token from the context
func GetCSRFToken(c *gin.Context) string {
	token, exists := c.Get("csrf_token")
	if !exists {
		return ""
	}
	return token.(string)
}
