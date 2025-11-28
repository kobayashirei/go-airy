package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/service"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
    userService service.UserService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	resp, err := h.userService.Register(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			response.BadRequest(c, "Invalid email format", nil)
		case errors.Is(err, service.ErrInvalidPhone):
			response.BadRequest(c, "Invalid phone format", nil)
		case errors.Is(err, service.ErrUserExists):
			response.Conflict(c, "User already exists")
		default:
			response.InternalError(c, "Failed to register user")
		}
		return
	}

	response.Success(c, resp)
}

// Activate handles user account activation
// POST /api/v1/auth/activate
func (h *AuthHandler) Activate(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		// Try to get from body
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Activation token is required", nil)
			return
		}
		token = req.Token
	}

	if err := h.userService.Activate(c.Request.Context(), token); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidToken):
			response.BadRequest(c, "Invalid or expired activation token", nil)
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, "User not found")
		default:
			response.InternalError(c, "Failed to activate account")
		}
		return
	}

	response.Success(c, gin.H{
		"message": "Account activated successfully",
	})
}

// ResendActivation handles resending activation email
// POST /api/v1/auth/resend-activation
func (h *AuthHandler) ResendActivation(c *gin.Context) {
    var req struct{
        Identifier string `json:"identifier" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "Identifier is required", nil)
        return
    }
    if err := h.userService.ResendActivation(c.Request.Context(), req.Identifier); err != nil {
        switch {
        case errors.Is(err, service.ErrUserNotFound):
            response.NotFound(c, "User not found")
        default:
            if strings.Contains(err.Error(), "already active") {
                response.BadRequest(c, "User already active", nil)
            } else if strings.Contains(err.Error(), "no email") {
                response.BadRequest(c, "User has no email to send activation", nil)
            } else {
                response.InternalError(c, "Failed to resend activation")
            }
        }
        return
    }
    response.Success(c, gin.H{"message":"Activation email resent"})
}

// Login handles user login with password
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Get client IP for login logging
	req.ClientIP = getClientIP(c)

	resp, err := h.userService.Login(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			response.Unauthorized(c, "Invalid credentials")
		case errors.Is(err, service.ErrUserNotFound):
			response.Unauthorized(c, "Invalid credentials")
		default:
			if strings.Contains(err.Error(), "not active") {
				response.Forbidden(c, "Account is not active. Please activate your account first.")
			} else {
				response.InternalError(c, "Failed to login")
			}
		}
		return
	}

	response.Success(c, resp)
}

// LoginWithCode handles user login with verification code
// POST /api/v1/auth/login/code
func (h *AuthHandler) LoginWithCode(c *gin.Context) {
	var req service.LoginWithCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Get client IP for login logging
	req.ClientIP = getClientIP(c)

	resp, err := h.userService.LoginWithCode(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidVerificationCode):
			response.BadRequest(c, "Invalid or expired verification code", nil)
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, "User not found")
		default:
			if strings.Contains(err.Error(), "not active") {
				response.Forbidden(c, "Account is not active")
			} else if strings.Contains(err.Error(), "invalid identifier") {
				response.BadRequest(c, "Invalid identifier format", nil)
			} else {
				response.InternalError(c, "Failed to login")
			}
		}
		return
	}

	response.Success(c, resp)
}

// RefreshToken handles token refresh
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get token from Authorization header or request body
	token := extractToken(c)
	if token == "" {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Refresh token is required", nil)
			return
		}
		token = req.RefreshToken
	}

	resp, err := h.userService.RefreshToken(c.Request.Context(), token)
	if err != nil {
		response.Unauthorized(c, "Invalid or expired refresh token")
		return
	}

	response.Success(c, resp)
}

// extractToken extracts the JWT token from the Authorization header
func extractToken(c *gin.Context) string {
	bearerToken := c.GetHeader("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

// getClientIP gets the client IP address from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (for proxies)
	ip := c.GetHeader("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	ip = c.GetHeader("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Fall back to RemoteAddr
	return c.ClientIP()
}
