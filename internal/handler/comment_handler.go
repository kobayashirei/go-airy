package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/service"
)

// CommentHandler handles comment-related HTTP requests
type CommentHandler struct {
	commentService service.CommentService
}

// NewCommentHandler creates a new comment handler
func NewCommentHandler(commentService service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// CreateComment handles comment creation
// POST /api/v1/posts/:id/comments
// Implements Requirement 5.1
func (h *CommentHandler) CreateComment(c *gin.Context) {
	// Parse post ID from URL parameter
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid post ID", nil)
		return
	}

	var req service.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Set post ID from URL parameter
	req.PostID = postID

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	req.AuthorID = userID.(int64)

	comment, err := h.commentService.CreateComment(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			response.NotFound(c, "Post not found")
		case errors.Is(err, service.ErrInvalidParentComment):
			response.BadRequest(c, "Invalid parent comment", nil)
		case err.Error() == "comments are not allowed on this post":
			response.Forbidden(c, "Comments are not allowed on this post")
		case err.Error() == "parent comment does not belong to the same post":
			response.BadRequest(c, "Parent comment does not belong to the same post", nil)
		default:
			response.InternalError(c, "Failed to create comment")
		}
		return
	}

	response.Success(c, comment)
}

// GetCommentTree handles retrieving all comments for a post in tree structure
// GET /api/v1/posts/:id/comments
// Implements Requirement 5.2
func (h *CommentHandler) GetCommentTree(c *gin.Context) {
	// Parse post ID from URL parameter
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid post ID", nil)
		return
	}

	commentTree, err := h.commentService.GetCommentTree(c.Request.Context(), postID)
	if err != nil {
		response.InternalError(c, "Failed to retrieve comments")
		return
	}

	response.Success(c, gin.H{
		"post_id":  postID,
		"comments": commentTree,
	})
}

// DeleteComment handles deleting a comment
// DELETE /api/v1/comments/:id
// Implements Requirement 5.1
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	// Parse comment ID from URL parameter
	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid comment ID", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	err = h.commentService.DeleteComment(c.Request.Context(), commentID, userID.(int64))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			response.NotFound(c, "Comment not found")
		case errors.Is(err, service.ErrUnauthorized):
			response.Forbidden(c, "You are not authorized to delete this comment")
		default:
			response.InternalError(c, "Failed to delete comment")
		}
		return
	}

	response.Success(c, gin.H{
		"message": "Comment deleted successfully",
	})
}
