package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/service"
)

// VoteHandler handles vote-related HTTP requests
type VoteHandler struct {
	voteService service.VoteService
}

// NewVoteHandler creates a new vote handler
func NewVoteHandler(voteService service.VoteService) *VoteHandler {
	return &VoteHandler{
		voteService: voteService,
	}
}

// VoteRequest represents a vote request
type VoteRequest struct {
	EntityType string `json:"entity_type" binding:"required,oneof=post comment"`
	EntityID   int64  `json:"entity_id" binding:"required,min=1"`
	VoteType   string `json:"vote_type" binding:"required,oneof=up down"`
}

// CancelVoteRequest represents a cancel vote request
type CancelVoteRequest struct {
	EntityType string `json:"entity_type" binding:"required,oneof=post comment"`
	EntityID   int64  `json:"entity_id" binding:"required,min=1"`
}

// VoteResponse represents a vote response
type VoteResponse struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	EntityType string `json:"entity_type"`
	EntityID   int64  `json:"entity_id"`
	VoteType   string `json:"vote_type"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// Vote handles POST /api/v1/votes
// @Summary Vote on a post or comment
// @Description Create or update a vote on a post or comment (idempotent)
// @Tags votes
// @Accept json
// @Produce json
// @Param request body VoteRequest true "Vote request"
// @Success 200 {object} response.Response{data=VoteResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/votes [post]
func (h *VoteHandler) Vote(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Parse request
	var req VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Call service
	vote, err := h.voteService.Vote(c.Request.Context(), userID.(int64), req.EntityType, req.EntityID, req.VoteType)
	if err != nil {
		if err.Error() == "post not found: "+strconv.FormatInt(req.EntityID, 10) ||
			err.Error() == "comment not found: "+strconv.FormatInt(req.EntityID, 10) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "Failed to vote")
		return
	}

	// Build response
	resp := VoteResponse{
		ID:         vote.ID,
		UserID:     vote.UserID,
		EntityType: vote.EntityType,
		EntityID:   vote.EntityID,
		VoteType:   vote.VoteType,
		CreatedAt:  vote.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  vote.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(c, resp)
}

// CancelVote handles DELETE /api/v1/votes
// @Summary Cancel a vote
// @Description Remove a user's vote on a post or comment
// @Tags votes
// @Accept json
// @Produce json
// @Param request body CancelVoteRequest true "Cancel vote request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/votes [delete]
func (h *VoteHandler) CancelVote(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Parse request
	var req CancelVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Call service
	err := h.voteService.CancelVote(c.Request.Context(), userID.(int64), req.EntityType, req.EntityID)
	if err != nil {
		response.InternalError(c, "Failed to cancel vote")
		return
	}

	response.Success(c, gin.H{"message": "vote cancelled successfully"})
}

// GetVote handles GET /api/v1/votes/:entity_type/:entity_id
// @Summary Get user's vote
// @Description Get the current user's vote on a specific entity
// @Tags votes
// @Produce json
// @Param entity_type path string true "Entity type (post or comment)"
// @Param entity_id path int true "Entity ID"
// @Success 200 {object} response.Response{data=VoteResponse}
// @Success 204 "No vote found"
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/votes/{entity_type}/{entity_id} [get]
func (h *VoteHandler) GetVote(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Parse path parameters
	entityType := c.Param("entity_type")
	if entityType != "post" && entityType != "comment" {
		response.BadRequest(c, "Invalid entity type", nil)
		return
	}

	entityIDStr := c.Param("entity_id")
	entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid entity ID", nil)
		return
	}

	// Call service
	vote, err := h.voteService.GetVote(c.Request.Context(), userID.(int64), entityType, entityID)
	if err != nil {
		response.InternalError(c, "Failed to get vote")
		return
	}

	// If no vote found, return 204 No Content
	if vote == nil {
		c.Status(http.StatusNoContent)
		return
	}

	// Build response
	resp := VoteResponse{
		ID:         vote.ID,
		UserID:     vote.UserID,
		EntityType: vote.EntityType,
		EntityID:   vote.EntityID,
		VoteType:   vote.VoteType,
		CreatedAt:  vote.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  vote.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(c, resp)
}
