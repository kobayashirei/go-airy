package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kobayashirei/airy/internal/auth"
	"github.com/kobayashirei/airy/internal/response"
)

// AuthMiddleware creates a JWT authentication middleware
func AuthMiddleware(jwtService auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate the token
		claims, err := jwtService.ParseToken(tokenString)
		if err != nil {
			if err == auth.ErrExpiredToken {
				response.Unauthorized(c, "token has expired")
			} else {
				response.Unauthorized(c, "invalid token")
			}
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("userID", claims.UserID)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't abort if no token is provided
func OptionalAuthMiddleware(jwtService auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := jwtService.ParseToken(tokenString)
		if err == nil {
			c.Set("userID", claims.UserID)
			c.Set("roles", claims.Roles)
		}

		c.Next()
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	id, ok := userID.(int64)
	return id, ok
}

// GetRoles retrieves the user roles from the context
func GetRoles(c *gin.Context) ([]string, bool) {
	roles, exists := c.Get("roles")
	if !exists {
		return nil, false
	}
	roleList, ok := roles.([]string)
	return roleList, ok
}
