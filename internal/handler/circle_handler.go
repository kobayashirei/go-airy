package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/service"
)

// CircleHandler handles circle-related HTTP requests
type CircleHandler struct {
	circleService service.CircleService
}

// NewCircleHandler creates a new circle handler
func NewCircleHandler(circleService service.CircleService) *CircleHandler {
	return &CircleHandler{
		circleService: circleService,
	}
}

// CreateCircle handles circle creation
// POST /api/v1/circles
// Requirement 7.1: WHEN a User creates a Circle THEN the System SHALL store the Circle information with creator ID and initial status
func (h *CircleHandler) CreateCircle(c *gin.Context) {
	var req service.CreateCircleRequest
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
	req.CreatorID = userID.(int64)

	circle, err := h.circleService.CreateCircle(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCircleNameExists):
			response.Conflict(c, "Circle name already exists")
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, "User not found")
		default:
			response.InternalError(c, "Failed to create circle")
		}
		return
	}

	response.Success(c, circle)
}

// GetCircle handles retrieving a single circle
// GET /api/v1/circles/:id
func (h *CircleHandler) GetCircle(c *gin.Context) {
	// Parse circle ID from URL parameter
	circleIDStr := c.Param("id")
	circleID, err := strconv.ParseInt(circleIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid circle ID", nil)
		return
	}

	circle, err := h.circleService.GetCircle(c.Request.Context(), circleID)
	if err != nil {
		if errors.Is(err, service.ErrCircleNotFound) {
			response.NotFound(c, "Circle not found")
		} else {
			response.InternalError(c, "Failed to retrieve circle")
		}
		return
	}

	response.Success(c, circle)
}

// JoinCircle handles a user joining a circle
// POST /api/v1/circles/:id/join
// Requirement 7.2: WHEN a User requests to join a Circle THEN the System SHALL process the request based on the Circle join rules
// Requirement 7.3: WHERE a Circle requires approval for joining THEN the System SHALL create a pending membership record for Moderator review
func (h *CircleHandler) JoinCircle(c *gin.Context) {
	// Parse circle ID from URL parameter
	circleIDStr := c.Param("id")
	circleID, err := strconv.ParseInt(circleIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid circle ID", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	err = h.circleService.JoinCircle(c.Request.Context(), circleID, userID.(int64))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCircleNotFound):
			response.NotFound(c, "Circle not found")
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, "User not found")
		case errors.Is(err, service.ErrAlreadyMember):
			response.Conflict(c, "User is already a member")
		case errors.Is(err, service.ErrPendingApproval):
			response.Conflict(c, "Membership request is pending approval")
		default:
			response.InternalError(c, "Failed to join circle")
		}
		return
	}

	response.Success(c, gin.H{
		"message": "Successfully joined circle",
	})
}

// ApproveMember handles approving a pending membership request
// POST /api/v1/circles/:id/members/:userId/approve
// Requirement 7.4: WHEN a Moderator approves a membership request THEN the System SHALL update the User-Circle relationship to active member
func (h *CircleHandler) ApproveMember(c *gin.Context) {
	// Parse circle ID from URL parameter
	circleIDStr := c.Param("id")
	circleID, err := strconv.ParseInt(circleIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid circle ID", nil)
		return
	}

	// Parse user ID from URL parameter
	userIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", nil)
		return
	}

	// Get approver ID from context
	approverID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	err = h.circleService.ApproveMember(c.Request.Context(), circleID, targetUserID, approverID.(int64))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotAuthorized):
			response.Forbidden(c, "You are not authorized to approve members")
		case errors.Is(err, service.ErrCannotApproveSelf):
			response.BadRequest(c, "Cannot approve your own membership", nil)
		case errors.Is(err, service.ErrNotMember):
			response.NotFound(c, "User is not a member")
		default:
			response.InternalError(c, "Failed to approve member")
		}
		return
	}

	response.Success(c, gin.H{
		"message": "Member approved successfully",
	})
}

// AssignModerator handles assigning a user as a moderator
// POST /api/v1/circles/:id/moderators
// Requirement 7.5: WHEN a User is assigned as Moderator THEN the System SHALL grant Circle-specific moderation Permissions
func (h *CircleHandler) AssignModerator(c *gin.Context) {
	// Parse circle ID from URL parameter
	circleIDStr := c.Param("id")
	circleID, err := strconv.ParseInt(circleIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid circle ID", nil)
		return
	}

	// Parse request body for target user ID
	var req struct {
		UserID int64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Get assigner ID from context
	assignerID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	err = h.circleService.AssignModerator(c.Request.Context(), circleID, req.UserID, assignerID.(int64))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotAuthorized):
			response.Forbidden(c, "You are not authorized to assign moderators")
		case errors.Is(err, service.ErrNotMember):
			response.NotFound(c, "User is not a member")
		default:
			response.InternalError(c, "Failed to assign moderator")
		}
		return
	}

	response.Success(c, gin.H{
		"message": "Moderator assigned successfully",
	})
}

// GetCircleMembers handles retrieving circle members
// GET /api/v1/circles/:id/members
func (h *CircleHandler) GetCircleMembers(c *gin.Context) {
	// Parse circle ID from URL parameter
	circleIDStr := c.Param("id")
	circleID, err := strconv.ParseInt(circleIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid circle ID", nil)
		return
	}

	// Optional role filter
	role := c.Query("role")

	members, err := h.circleService.GetCircleMembers(c.Request.Context(), circleID, role)
	if err != nil {
		response.InternalError(c, "Failed to retrieve members")
		return
	}

	response.Success(c, members)
}
