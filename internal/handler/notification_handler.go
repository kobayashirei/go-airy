package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/service"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationService service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotifications handles retrieving notification list for the authenticated user
// GET /api/v1/notifications
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Parse pagination parameters
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Get notifications
	result, err := h.notificationService.GetNotifications(c.Request.Context(), userID.(int64), page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to retrieve notifications")
		return
	}

	response.Success(c, result)
}

// MarkAsRead handles marking a single notification as read
// PUT /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Parse notification ID from URL parameter
	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseInt(notificationIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid notification ID", nil)
		return
	}

	// Mark as read
	if err := h.notificationService.MarkAsRead(c.Request.Context(), userID.(int64), notificationID); err != nil {
		if err == service.ErrNotificationNotFound {
			response.NotFound(c, "Notification not found")
		} else if err == service.ErrUnauthorizedNotification {
			response.Forbidden(c, "Unauthorized to access this notification")
		} else {
			response.InternalError(c, "Failed to mark notification as read")
		}
		return
	}

	response.Success(c, gin.H{"message": "Notification marked as read"})
}

// MarkAllAsRead handles marking all notifications as read for the authenticated user
// PUT /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Mark all as read
	if err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID.(int64)); err != nil {
		response.InternalError(c, "Failed to mark all notifications as read")
		return
	}

	response.Success(c, gin.H{"message": "All notifications marked as read"})
}

// GetUnreadCount handles retrieving the count of unread notifications
// GET /api/v1/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Get unread count
	count, err := h.notificationService.GetUnreadCount(c.Request.Context(), userID.(int64))
	if err != nil {
		response.InternalError(c, "Failed to retrieve unread count")
		return
	}

	response.Success(c, gin.H{"unread_count": count})
}
