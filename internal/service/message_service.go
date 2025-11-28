package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/repository"
)

var (
	// ErrConversationNotFound is returned when conversation is not found
	ErrConversationNotFound = errors.New("conversation not found")
	// ErrMessageNotFound is returned when message is not found
	ErrMessageNotFound = errors.New("message not found")
	// ErrUnauthorizedConversation is returned when user tries to access conversation they're not part of
	ErrUnauthorizedConversation = errors.New("unauthorized to access this conversation")
	// ErrInvalidMessageContent is returned when message content is empty or invalid
	ErrInvalidMessageContent = errors.New("message content cannot be empty")
	// ErrCannotMessageSelf is returned when user tries to message themselves
	ErrCannotMessageSelf = errors.New("cannot send message to yourself")
)

// MessageService defines the interface for messaging business logic
type MessageService interface {
	// GetOrCreateConversation gets an existing conversation or creates a new one between two users
	GetOrCreateConversation(ctx context.Context, user1ID, user2ID int64) (*models.Conversation, error)
	// SendMessage sends a message in a conversation
	SendMessage(ctx context.Context, req SendMessageRequest) (*models.Message, error)
	// GetConversations retrieves all conversations for a user, sorted by last message time
	GetConversations(ctx context.Context, userID int64, page, pageSize int) (*ConversationListResponse, error)
	// GetMessages retrieves messages in a conversation with pagination
	GetMessages(ctx context.Context, userID, conversationID int64, page, pageSize int) (*MessageListResponse, error)
	// MarkConversationAsRead marks all messages in a conversation as read for the user
	MarkConversationAsRead(ctx context.Context, userID, conversationID int64) error
}

// SendMessageRequest represents a message sending request
type SendMessageRequest struct {
	SenderID       int64  `json:"sender_id"`
	ReceiverID     int64  `json:"receiver_id"`
	ConversationID int64  `json:"conversation_id"`
	ContentType    string `json:"content_type"` // text, image
	Content        string `json:"content"`
}

// ConversationListResponse represents a paginated list of conversations
type ConversationListResponse struct {
	Conversations []*ConversationWithDetails `json:"conversations"`
	Total         int64                      `json:"total"`
	Page          int                        `json:"page"`
	PageSize      int                        `json:"page_size"`
}

// ConversationWithDetails includes conversation and additional details
type ConversationWithDetails struct {
	*models.Conversation
	OtherUserID      int64          `json:"other_user_id"`
	OtherUser        *models.User   `json:"other_user,omitempty"`
	LastMessage      *models.Message `json:"last_message,omitempty"`
	UnreadCount      int64          `json:"unread_count"`
}

// MessageListResponse represents a paginated list of messages
type MessageListResponse struct {
	Messages []*models.Message `json:"messages"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// messageService implements MessageService interface
type messageService struct {
	conversationRepo repository.ConversationRepository
	messageRepo      repository.MessageRepository
	userRepo         repository.UserRepository
}

// NewMessageService creates a new message service
func NewMessageService(
	conversationRepo repository.ConversationRepository,
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
) MessageService {
	return &messageService{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		userRepo:         userRepo,
	}
}

// GetOrCreateConversation gets an existing conversation or creates a new one between two users
// This ensures conversation uniqueness between two users
func (s *messageService) GetOrCreateConversation(ctx context.Context, user1ID, user2ID int64) (*models.Conversation, error) {
	// Validate users are different
	if user1ID == user2ID {
		return nil, ErrCannotMessageSelf
	}

	// Ensure consistent ordering: smaller ID as user1, larger as user2
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}

	// Try to find existing conversation
	conversation, err := s.conversationRepo.FindByUsers(ctx, user1ID, user2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find conversation: %w", err)
	}

	// If conversation exists, return it
	if conversation != nil {
		return conversation, nil
	}

	// Verify both users exist
	user1, err := s.userRepo.FindByID(ctx, user1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user1: %w", err)
	}
	if user1 == nil {
		return nil, fmt.Errorf("user1 not found")
	}

	user2, err := s.userRepo.FindByID(ctx, user2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user2: %w", err)
	}
	if user2 == nil {
		return nil, fmt.Errorf("user2 not found")
	}

	// Create new conversation
	conversation = &models.Conversation{
		User1ID:       user1ID,
		User2ID:       user2ID,
		LastMessageAt: time.Now(),
		CreatedAt:     time.Now(),
	}

	if err := s.conversationRepo.Create(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conversation, nil
}

// SendMessage sends a message in a conversation
func (s *messageService) SendMessage(ctx context.Context, req SendMessageRequest) (*models.Message, error) {
	// Validate content
	if req.Content == "" {
		return nil, ErrInvalidMessageContent
	}

	// Validate content type
	if req.ContentType == "" {
		req.ContentType = "text"
	}
	validContentTypes := map[string]bool{
		"text":  true,
		"image": true,
	}
	if !validContentTypes[req.ContentType] {
		return nil, fmt.Errorf("invalid content type: %s", req.ContentType)
	}

	// Verify conversation exists and sender is part of it
	conversation, err := s.conversationRepo.FindByID(ctx, req.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find conversation: %w", err)
	}
	if conversation == nil {
		return nil, ErrConversationNotFound
	}

	// Verify sender is part of the conversation
	if conversation.User1ID != req.SenderID && conversation.User2ID != req.SenderID {
		return nil, ErrUnauthorizedConversation
	}

	// Create message
	message := &models.Message{
		ConversationID: req.ConversationID,
		SenderID:       req.SenderID,
		ContentType:    req.ContentType,
		Content:        req.Content,
		IsRead:         false,
		CreatedAt:      time.Now(),
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update conversation's last message timestamp
	if err := s.conversationRepo.UpdateLastMessageAt(ctx, req.ConversationID); err != nil {
		// Log error but don't fail the request
		// In production, you might want to handle this differently
		fmt.Printf("warning: failed to update conversation last message time: %v\n", err)
	}

	return message, nil
}

// GetConversations retrieves all conversations for a user, sorted by last message time
func (s *messageService) GetConversations(ctx context.Context, userID int64, page, pageSize int) (*ConversationListResponse, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get conversations (repository already orders by last_message_at DESC)
	conversations, err := s.conversationRepo.FindByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve conversations: %w", err)
	}

	// Enrich conversations with additional details
	conversationsWithDetails := make([]*ConversationWithDetails, 0, len(conversations))
	for _, conv := range conversations {
		// Determine the other user ID
		otherUserID := conv.User1ID
		if conv.User1ID == userID {
			otherUserID = conv.User2ID
		}

		// Get other user details
		otherUser, err := s.userRepo.FindByID(ctx, otherUserID)
		if err != nil {
			// Log error but continue
			fmt.Printf("warning: failed to get user %d: %v\n", otherUserID, err)
		}

		// Get last message
		messages, err := s.messageRepo.FindByConversationID(ctx, conv.ID, 1, 0)
		var lastMessage *models.Message
		if err == nil && len(messages) > 0 {
			lastMessage = messages[0]
		}

		// Get unread count for this user
		unreadMessages, err := s.messageRepo.FindUnreadByConversationAndReceiver(ctx, conv.ID, userID)
		unreadCount := int64(0)
		if err == nil {
			unreadCount = int64(len(unreadMessages))
		}

		conversationsWithDetails = append(conversationsWithDetails, &ConversationWithDetails{
			Conversation:    conv,
			OtherUserID:     otherUserID,
			OtherUser:       otherUser,
			LastMessage:     lastMessage,
			UnreadCount:     unreadCount,
		})
	}

	// For total count, we would need another query or cache this value
	total := int64(len(conversations))
	if len(conversations) == pageSize {
		total = int64(offset + pageSize + 1)
	}

	return &ConversationListResponse{
		Conversations: conversationsWithDetails,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

// GetMessages retrieves messages in a conversation with pagination
func (s *messageService) GetMessages(ctx context.Context, userID, conversationID int64, page, pageSize int) (*MessageListResponse, error) {
	// Verify conversation exists and user is part of it
	conversation, err := s.conversationRepo.FindByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find conversation: %w", err)
	}
	if conversation == nil {
		return nil, ErrConversationNotFound
	}

	// Verify user is part of the conversation
	if conversation.User1ID != userID && conversation.User2ID != userID {
		return nil, ErrUnauthorizedConversation
	}

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize

	// Get messages (repository orders by created_at DESC)
	messages, err := s.messageRepo.FindByConversationID(ctx, conversationID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve messages: %w", err)
	}

	// For total count
	total := int64(len(messages))
	if len(messages) == pageSize {
		total = int64(offset + pageSize + 1)
	}

	return &MessageListResponse{
		Messages: messages,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// MarkConversationAsRead marks all messages in a conversation as read for the user
func (s *messageService) MarkConversationAsRead(ctx context.Context, userID, conversationID int64) error {
	// Verify conversation exists and user is part of it
	conversation, err := s.conversationRepo.FindByID(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to find conversation: %w", err)
	}
	if conversation == nil {
		return ErrConversationNotFound
	}

	// Verify user is part of the conversation
	if conversation.User1ID != userID && conversation.User2ID != userID {
		return ErrUnauthorizedConversation
	}

	// Mark all messages in the conversation as read for this user
	// (messages where sender is not the current user)
	if err := s.messageRepo.MarkConversationAsRead(ctx, conversationID, userID); err != nil {
		return fmt.Errorf("failed to mark conversation as read: %w", err)
	}

	return nil
}
