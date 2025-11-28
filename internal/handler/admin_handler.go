package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/service"
)

// AdminHandler handles admin-related HTTP requests
type AdminHandler struct {
	adminService service.AdminService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// GetDashboard retrieves dashboard metrics
// GET /api/v1/admin/dashboard
func (h *AdminHandler) GetDashboard(c *gin.Context) {
	dashboard, err := h.adminService.GetDashboard(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get dashboard", err.Error())
		return
	}

	response.Success(c, dashboard)
}

// ListUsers retrieves a list of users with filtering and pagination
// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var req service.ListUsersRequest

	// Parse query parameters
	req.Status = c.Query("status")
	req.Keyword = c.Query("keyword")
	req.SortBy = c.DefaultQuery("sort_by", "created_at")
	req.Order = c.DefaultQuery("order", "desc")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	req.Page = page
	req.PageSize = pageSize

	result, err := h.adminService.ListUsers(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list users", err.Error())
		return
	}

	response.Success(c, result)
}

// BanUser bans a user
// POST /api/v1/admin/users/:id/ban
func (h *AdminHandler) BanUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid user ID", err.Error())
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request", err.Error())
		return
	}

	// Get operator ID from context (set by auth middleware)
	operatorID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	clientIP := c.ClientIP()
	if err := h.adminService.BanUser(c.Request.Context(), operatorID.(int64), userID, req.Reason, clientIP); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to ban user", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "user banned successfully"})
}

// UnbanUser unbans a user
// POST /api/v1/admin/users/:id/unban
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid user ID", err.Error())
		return
	}

	// Get operator ID from context (set by auth middleware)
	operatorID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	clientIP := c.ClientIP()
	if err := h.adminService.UnbanUser(c.Request.Context(), operatorID.(int64), userID, clientIP); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to unban user", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "user unbanned successfully"})
}

// ListPosts retrieves a list of posts with filtering and pagination
// GET /api/v1/admin/posts
func (h *AdminHandler) ListPosts(c *gin.Context) {
	var req service.AdminListPostsRequest

	// Parse query parameters
	req.Status = c.Query("status")
	req.Keyword = c.Query("keyword")
	req.SortBy = c.DefaultQuery("sort_by", "created_at")
	req.Order = c.DefaultQuery("order", "desc")

	if authorIDStr := c.Query("author_id"); authorIDStr != "" {
		authorID, _ := strconv.ParseInt(authorIDStr, 10, 64)
		req.AuthorID = &authorID
	}

	if circleIDStr := c.Query("circle_id"); circleIDStr != "" {
		circleID, _ := strconv.ParseInt(circleIDStr, 10, 64)
		req.CircleID = &circleID
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	req.Page = page
	req.PageSize = pageSize

	result, err := h.adminService.ListPosts(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list posts", err.Error())
		return
	}

	response.Success(c, result)
}

// BatchReviewPosts reviews multiple posts in batch
// POST /api/v1/admin/posts/batch-review
func (h *AdminHandler) BatchReviewPosts(c *gin.Context) {
	var req service.BatchReviewRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request", err.Error())
		return
	}

	// Get operator ID from context (set by auth middleware)
	operatorID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	req.OperatorID = operatorID.(int64)
	req.IP = c.ClientIP()

	if err := h.adminService.BatchReviewPosts(c.Request.Context(), req); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to batch review posts", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "posts reviewed successfully"})
}

// ListLogs retrieves a list of admin logs with filtering and pagination
// GET /api/v1/admin/logs
func (h *AdminHandler) ListLogs(c *gin.Context) {
	var req service.ListLogsRequest

	// Parse query parameters
	if operatorIDStr := c.Query("operator_id"); operatorIDStr != "" {
		operatorID, _ := strconv.ParseInt(operatorIDStr, 10, 64)
		req.OperatorID = &operatorID
	}

	req.Action = c.Query("action")
	req.EntityType = c.Query("entity_type")

	if startDate := c.Query("start_date"); startDate != "" {
		req.StartDate = &startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		req.EndDate = &endDate
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	req.Page = page
	req.PageSize = pageSize

	result, err := h.adminService.ListLogs(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list logs", err.Error())
		return
	}

	response.Success(c, result)
}
