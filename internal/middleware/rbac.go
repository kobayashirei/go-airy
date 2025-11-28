package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/service"
)

// RBACMiddleware creates a role-based access control middleware
func RBACMiddleware(permissionService service.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// RequirePermission creates a middleware that checks if the user has the required permission
func RequirePermission(permissionService service.PermissionService, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		_, exists := GetUserID(c)
		if !exists {
			response.Unauthorized(c, "user not authenticated")
			c.Abort()
			return
		}

		// Get roles from context
		roles, exists := GetRoles(c)
		if !exists || len(roles) == 0 {
			response.Forbidden(c, "user has no roles assigned")
			c.Abort()
			return
		}

		// Check if user has the required permission
		hasPermission, err := permissionService.CheckPermission(c.Request.Context(), roles, permission)
		if err != nil {
			response.InternalError(c, "failed to check permission")
			c.Abort()
			return
		}

		if !hasPermission {
			response.Forbidden(c, "insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireCirclePermission creates a middleware that checks if the user has the required permission in a specific circle
func RequireCirclePermission(permissionService service.PermissionService, permission string, circleIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := GetUserID(c)
		if !exists {
			response.Unauthorized(c, "user not authenticated")
			c.Abort()
			return
		}

		// Get circle ID from URL parameter
		var circleID *int64
		if circleIDParam != "" {
			circleIDStr := c.Param(circleIDParam)
			if circleIDStr != "" {
				id, err := strconv.ParseInt(circleIDStr, 10, 64)
				if err != nil {
					response.BadRequest(c, "invalid circle ID", nil)
					c.Abort()
					return
				}
				circleID = &id
			}
		}

		// Check if user has the required permission in the circle
		hasPermission, err := permissionService.CheckPermissionWithCircle(c.Request.Context(), userID, circleID, permission)
		if err != nil {
			response.InternalError(c, "failed to check permission")
			c.Abort()
			return
		}

		if !hasPermission {
			response.Forbidden(c, "insufficient permissions for this circle")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission creates a middleware that checks if the user has any of the required permissions
func RequireAnyPermission(permissionService service.PermissionService, permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		_, exists := GetUserID(c)
		if !exists {
			response.Unauthorized(c, "user not authenticated")
			c.Abort()
			return
		}

		// Get roles from context
		roles, exists := GetRoles(c)
		if !exists || len(roles) == 0 {
			response.Forbidden(c, "user has no roles assigned")
			c.Abort()
			return
		}

		// Check if user has any of the required permissions
		for _, permission := range permissions {
			hasPermission, err := permissionService.CheckPermission(c.Request.Context(), roles, permission)
			if err != nil {
				response.InternalError(c, "failed to check permission")
				c.Abort()
				return
			}

			if hasPermission {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "insufficient permissions")
		c.Abort()
	}
}

// RequireAllPermissions creates a middleware that checks if the user has all of the required permissions
func RequireAllPermissions(permissionService service.PermissionService, permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		_, exists := GetUserID(c)
		if !exists {
			response.Unauthorized(c, "user not authenticated")
			c.Abort()
			return
		}

		// Get roles from context
		roles, exists := GetRoles(c)
		if !exists || len(roles) == 0 {
			response.Forbidden(c, "user has no roles assigned")
			c.Abort()
			return
		}

		// Check if user has all of the required permissions
		for _, permission := range permissions {
			hasPermission, err := permissionService.CheckPermission(c.Request.Context(), roles, permission)
			if err != nil {
				response.InternalError(c, "failed to check permission")
				c.Abort()
				return
			}

			if !hasPermission {
				response.Forbidden(c, "insufficient permissions")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
