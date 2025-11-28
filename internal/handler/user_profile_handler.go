package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/service"
)

// UserProfileHandler handles user profile-related HTTP requests
type UserProfileHandler struct {
	userProfileService service.UserProfileService
}

// NewUserProfileHandler creates a new user profile handler
func NewUserProfileHandler(userProfileService service.UserProfileService) *UserProfileHandler {
	return &UserProfileHandler{
		userProfileService: userProfileService,
	}
}

// GetProfile handles retrieving a user's profile
// GET /api/v1/users/:id/profile
func (h *UserProfileHandler) GetProfile(c *gin.Context) {
	// Parse user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", nil)
		return
	}

	// Get profile from service
	profile, err := h.userProfileService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, "User not found")
		default:
			response.InternalError(c, "Failed to retrieve user profile")
		}
		return
	}

	response.Success(c, profile)
}

// UpdateProfile handles updating the authenticated user's profile
// PUT /api/v1/users/profile
func (h *UserProfileHandler) UpdateProfile(c *gin.Context) {
	// Get authenticated user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		response.InternalError(c, "Invalid user ID in context")
		return
	}

	// Parse request body
	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Update profile
	if err := h.userProfileService.UpdateProfile(c.Request.Context(), userID, req); err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, "User not found")
		default:
			response.InternalError(c, "Failed to update profile")
		}
		return
	}

	response.Success(c, gin.H{
		"message": "Profile updated successfully",
	})
}

// GetUserPosts handles retrieving posts created by a user
// GET /api/v1/users/:id/posts
func (h *UserProfileHandler) GetUserPosts(c *gin.Context) {
	// Parse user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", nil)
		return
	}

	// Parse query parameters
	opts := service.UserPostsOptions{
		Status: c.DefaultQuery("status", "published"),
		SortBy: c.DefaultQuery("sort_by", "created_at"),
		Order:  c.DefaultQuery("order", "desc"),
	}

	// Parse page and limit
	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			response.BadRequest(c, "Invalid page parameter", nil)
			return
		}
		opts.Page = page
	} else {
		opts.Page = 1
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			response.BadRequest(c, "Invalid limit parameter (must be between 1 and 100)", nil)
			return
		}
		opts.Limit = limit
	} else {
		opts.Limit = 20
	}

	// Get user posts
	postsResp, err := h.userProfileService.GetUserPosts(c.Request.Context(), userID, opts)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, "User not found")
		default:
			response.InternalError(c, "Failed to retrieve user posts")
		}
		return
	}

	response.Success(c, postsResp)
}
