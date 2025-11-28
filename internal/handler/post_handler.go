package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/service"
)

// PostHandler handles post-related HTTP requests
type PostHandler struct {
	postService service.PostService
}

// NewPostHandler creates a new post handler
func NewPostHandler(postService service.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// CreatePost handles post creation
// POST /api/v1/posts
func (h *PostHandler) CreatePost(c *gin.Context) {
	var req service.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	req.AuthorID = userID.(int64)

	post, err := h.postService.CreatePost(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrModerationFailed) {
			response.InternalError(c, "Content moderation failed")
		} else {
			response.InternalError(c, "Failed to create post")
		}
		return
	}

	response.Success(c, post)
}

// GetPost handles retrieving a single post
// GET /api/v1/posts/:id
func (h *PostHandler) GetPost(c *gin.Context) {
	// Parse post ID from URL parameter
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid post ID", nil)
		return
	}

	// Get user ID from context if authenticated (optional)
	var userID *int64
	if uid, exists := c.Get("userID"); exists {
		id := uid.(int64)
		userID = &id
	}

	post, err := h.postService.GetPost(c.Request.Context(), postID, userID)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			response.NotFound(c, "Post not found")
		} else {
			response.InternalError(c, "Failed to retrieve post")
		}
		return
	}

	response.Success(c, post)
}

// UpdatePost handles updating a post
// PUT /api/v1/posts/:id
func (h *PostHandler) UpdatePost(c *gin.Context) {
	// Parse post ID from URL parameter
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid post ID", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req service.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	post, err := h.postService.UpdatePost(c.Request.Context(), postID, userID.(int64), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			response.NotFound(c, "Post not found")
		case errors.Is(err, service.ErrUnauthorized):
			response.Forbidden(c, "You are not authorized to update this post")
		case errors.Is(err, service.ErrModerationFailed):
			response.InternalError(c, "Content moderation failed")
		default:
			response.InternalError(c, "Failed to update post")
		}
		return
	}

	response.Success(c, post)
}

// DeletePost handles deleting a post
// DELETE /api/v1/posts/:id
func (h *PostHandler) DeletePost(c *gin.Context) {
	// Parse post ID from URL parameter
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid post ID", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	err = h.postService.DeletePost(c.Request.Context(), postID, userID.(int64))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			response.NotFound(c, "Post not found")
		case errors.Is(err, service.ErrUnauthorized):
			response.Forbidden(c, "You are not authorized to delete this post")
		default:
			response.InternalError(c, "Failed to delete post")
		}
		return
	}

	response.Success(c, gin.H{
		"message": "Post deleted successfully",
	})
}

// ListPosts handles listing posts with pagination and filtering
// GET /api/v1/posts
func (h *PostHandler) ListPosts(c *gin.Context) {
	var req service.ListPostsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Set defaults if not provided
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	result, err := h.postService.ListPosts(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, "Failed to list posts")
		return
	}

	response.Success(c, result)
}
