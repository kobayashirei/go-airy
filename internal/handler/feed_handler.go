package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	appLogger "github.com/kobayashirei/airy/internal/logger"
	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/service"
)

// FeedHandler handles feed-related HTTP requests
type FeedHandler struct {
	feedService service.FeedService
}

// NewFeedHandler creates a new feed handler
func NewFeedHandler(feedService service.FeedService) *FeedHandler {
	return &FeedHandler{
		feedService: feedService,
	}
}

// GetUserFeedRequest represents the request for getting user feed
type GetUserFeedRequest struct {
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int    `form:"offset" binding:"omitempty,min=0"`
	SortBy string `form:"sort_by" binding:"omitempty,oneof=created_at hotness_score"`
}

// GetUserFeed handles GET /api/v1/feed
// @Summary Get user's personalized feed
// @Description Retrieves the authenticated user's personalized content feed
// @Tags Feed
// @Accept json
// @Produce json
// @Param limit query int false "Number of posts to return (default: 20, max: 100)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Param sort_by query string false "Sort by field: created_at or hotness_score (default: created_at)"
// @Success 200 {object} response.Response{data=[]models.Post}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/feed [get]
func (h *FeedHandler) GetUserFeed(c *gin.Context) {
	// Get authenticated user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		appLogger.Error("User ID not found in context")
		response.Error(c, http.StatusUnauthorized, "unauthorized", "User not authenticated", nil)
		return
	}

	// Parse and validate request
	var req GetUserFeedRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		appLogger.Warn("Invalid feed request", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}

	// Get user feed
	posts, err := h.feedService.GetUserFeed(c.Request.Context(), userID.(int64), req.Limit, req.Offset, req.SortBy)
	if err != nil {
		appLogger.Error("Failed to get user feed",
			zap.Int64("user_id", userID.(int64)),
			zap.Error(err),
		)
		response.Error(c, http.StatusInternalServerError, "feed_error", "Failed to retrieve feed", nil)
		return
	}

	response.Success(c, posts)
}

// GetCircleFeedRequest represents the request for getting circle feed
type GetCircleFeedRequest struct {
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int    `form:"offset" binding:"omitempty,min=0"`
	SortBy string `form:"sort_by" binding:"omitempty,oneof=created_at hotness_score"`
}

// GetCircleFeed handles GET /api/v1/circles/:id/feed
// @Summary Get circle feed
// @Description Retrieves posts from a specific circle
// @Tags Feed
// @Accept json
// @Produce json
// @Param id path int true "Circle ID"
// @Param limit query int false "Number of posts to return (default: 20, max: 100)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Param sort_by query string false "Sort by field: created_at or hotness_score (default: created_at)"
// @Success 200 {object} response.Response{data=[]models.Post}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/circles/{id}/feed [get]
func (h *FeedHandler) GetCircleFeed(c *gin.Context) {
	// Parse circle ID from path
	circleIDStr := c.Param("id")
	circleID, err := strconv.ParseInt(circleIDStr, 10, 64)
	if err != nil {
		appLogger.Warn("Invalid circle ID", zap.String("id", circleIDStr))
		response.Error(c, http.StatusBadRequest, "invalid_id", "Invalid circle ID", nil)
		return
	}

	// Parse and validate request
	var req GetCircleFeedRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		appLogger.Warn("Invalid circle feed request", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}

	// Get circle feed
	posts, err := h.feedService.GetCircleFeed(c.Request.Context(), circleID, req.Limit, req.Offset, req.SortBy)
	if err != nil {
		appLogger.Error("Failed to get circle feed",
			zap.Int64("circle_id", circleID),
			zap.Error(err),
		)
		response.Error(c, http.StatusInternalServerError, "feed_error", "Failed to retrieve circle feed", nil)
		return
	}

	response.Success(c, posts)
}

// RegisterRoutes registers feed routes
func (h *FeedHandler) RegisterRoutes(r *gin.RouterGroup) {
	feed := r.Group("/feed")
	{
		feed.GET("", h.GetUserFeed)
	}
	
	circles := r.Group("/circles")
	{
		circles.GET("/:id/feed", h.GetCircleFeed)
	}
}
