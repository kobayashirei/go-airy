package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kobayashirei/airy/internal/auth"
	"github.com/kobayashirei/airy/internal/middleware"
	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/service"
)

// This is an example showing how to use the authentication and authorization system
// This file is for demonstration purposes only and should not be used in production

func main() {
	router := gin.Default()

	// Initialize JWT service
	jwtService := auth.NewJWTService("your-secret-key", 24*time.Hour)

	// Initialize permission service (you would inject real repositories here)
	var permissionService service.PermissionService
	// permissionService = service.NewPermissionService(permissionRepo, roleRepo, userRoleRepo)

	// Public routes (no authentication required)
	public := router.Group("/api/v1")
	{
		public.POST("/auth/login", func(c *gin.Context) {
			// Login logic here
			// After successful authentication:
			userID := int64(123)
			roles := []string{"user"}

			token, err := jwtService.GenerateToken(userID, roles)
			if err != nil {
				response.InternalError(c, "failed to generate token")
				return
			}

			response.Success(c, gin.H{
				"token": token,
			})
		})

		public.POST("/auth/register", func(c *gin.Context) {
			// Registration logic here
			response.Success(c, gin.H{
				"message": "registration successful",
			})
		})
	}

	// Protected routes (authentication required)
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(jwtService))
	{
		// Simple authenticated route
		protected.GET("/profile", func(c *gin.Context) {
			userID, _ := middleware.GetUserID(c)
			roles, _ := middleware.GetRoles(c)

			response.Success(c, gin.H{
				"user_id": userID,
				"roles":   roles,
			})
		})

		// Route with permission check
		protected.POST("/posts",
			middleware.RequirePermission(permissionService, "post:create"),
			func(c *gin.Context) {
				response.Success(c, gin.H{
					"message": "post created",
				})
			},
		)

		// Route with circle-specific permission check
		protected.DELETE("/circles/:circleId/posts/:id",
			middleware.RequireCirclePermission(permissionService, "post:delete", "circleId"),
			func(c *gin.Context) {
				response.Success(c, gin.H{
					"message": "post deleted",
				})
			},
		)

		// Admin routes with multiple permission checks
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireAnyPermission(permissionService, "admin:view", "moderator:view"))
		{
			admin.GET("/dashboard", func(c *gin.Context) {
				response.Success(c, gin.H{
					"message": "admin dashboard",
				})
			})

			admin.POST("/users/ban",
				middleware.RequireAllPermissions(permissionService, "admin:users", "admin:ban"),
				func(c *gin.Context) {
					response.Success(c, gin.H{
						"message": "user banned",
					})
				},
			)
		}

		// Token refresh endpoint
		protected.POST("/auth/refresh", func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				response.Unauthorized(c, "missing token")
				return
			}

			// Extract token
			token := authHeader[7:] // Remove "Bearer " prefix

			newToken, err := jwtService.RefreshToken(token)
			if err != nil {
				response.Unauthorized(c, "failed to refresh token")
				return
			}

			response.Success(c, gin.H{
				"token": newToken,
			})
		})
	}

	// Optional authentication (for routes that work with or without auth)
	optional := router.Group("/api/v1")
	optional.Use(middleware.OptionalAuthMiddleware(jwtService))
	{
		optional.GET("/posts", func(c *gin.Context) {
			userID, authenticated := middleware.GetUserID(c)

			if authenticated {
				// Return personalized content
				response.Success(c, gin.H{
					"message": "personalized posts",
					"user_id": userID,
				})
			} else {
				// Return public content
				response.Success(c, gin.H{
					"message": "public posts",
				})
			}
		})
	}

	router.Run(":8080")
}
