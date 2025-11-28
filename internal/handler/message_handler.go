package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/service"
)

// MessageHandler handles message-related HTTP requests
type MessageHandler struct {
	messageService service.MessageService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// GetConversations handles retrieving conversation list for the authenticated user
// GET /api/v1/conversations
func (h *MessageHandler) GetConversations(c *gin.Context) {
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

	// Get conversations
	result, err := h.messageService.GetConversations(c.Request.Context(), userID.(int64), page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to retrieve conversations")
		return
	}

	response.Success(c, result)
}

// GetMessages handles retrieving messages in a conversation
// GET /api/v1/conversations/:id/messages
func (h *MessageHandler) GetMessages(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Parse conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid conversation ID", nil)
		return
	}

	// Parse pagination parameters
	page := 1
	pageSize := 50

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

	// Get messages
	result, err := h.messageService.GetMessages(c.Request.Context(), userID.(int64), conversationID, page, pageSize)
	if err != nil {
		if err == service.ErrConversationNotFound {
			response.NotFound(c, "Conversation not found")
		} else if err == service.ErrUnauthorizedConversation {
			response.Forbidden(c, "Unauthorized to access this conversation")
		} else {
			response.InternalError(c, "Failed to retrieve messages")
		}
		return
	}

	// Mark conversation as read when user opens it
	if err := h.messageService.MarkConversationAsRead(c.Request.Context(), userID.(int64), conversationID); err != nil {
		// Log error but don't fail the request
		// In production, you might want to handle this differently
		c.Error(err)
	}

	response.Success(c, result)
}

// SendMessage handles sending a message in a conversation
// POST /api/v1/conversations/:id/messages
func (h *MessageHandler) SendMessage(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Parse conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid conversation ID", nil)
		return
	}

	// Parse request body
	var req struct {
		ContentType string `json:"content_type"`
		Content     string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	// Send message
	message, err := h.messageService.SendMessage(c.Request.Context(), service.SendMessageRequest{
		SenderID:       userID.(int64),
		ConversationID: conversationID,
		ContentType:    req.ContentType,
		Content:        req.Content,
	})

	if err != nil {
		if err == service.ErrConversationNotFound {
			response.NotFound(c, "Conversation not found")
		} else if err == service.ErrUnauthorizedConversation {
			response.Forbidden(c, "Unauthorized to send message in this conversation")
		} else if err == service.ErrInvalidMessageContent {
			response.BadRequest(c, "Message content cannot be empty", nil)
		} else {
			response.InternalError(c, "Failed to send message")
		}
		return
	}

	response.Success(c, message)
}

// CreateConversation handles creating or getting a conversation with another user
// POST /api/v1/conversations
func (h *MessageHandler) CreateConversation(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		OtherUserID int64 `json:"other_user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	// Get or create conversation
	conversation, err := h.messageService.GetOrCreateConversation(c.Request.Context(), userID.(int64), req.OtherUserID)
	if err != nil {
		if err == service.ErrCannotMessageSelf {
			response.BadRequest(c, "Cannot send message to yourself", nil)
		} else {
			response.InternalError(c, "Failed to create conversation")
		}
		return
	}

	response.Success(c, conversation)
}
