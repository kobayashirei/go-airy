package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kobayashirei/airy/internal/models"
)

func TestGetOrCreateConversation_CreateNew(t *testing.T) {
	// Setup mocks
	mockConversationRepo := new(MockConversationRepository)
	mockMessageRepo := new(MockMessageRepository)
	mockUserRepo := new(MockUserRepository)

	// Create service
	service := NewMessageService(mockConversationRepo, mockMessageRepo, mockUserRepo)

	// Setup expectations - no existing conversation
	mockConversationRepo.On("FindByUsers", mock.Anything, int64(1), int64(2)).Return(nil, nil)

	// Users exist
	mockUserRepo.On("FindByID", mock.Anything, int64(1)).Return(&models.User{ID: 1, Username: "user1"}, nil)
	mockUserRepo.On("FindByID", mock.Anything, int64(2)).Return(&models.User{ID: 2, Username: "user2"}, nil)

	// Create conversation
	mockConversationRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *models.Conversation) bool {
		return c.User1ID == 1 && c.User2ID == 2
	})).Return(nil)

	// Test
	conversation, err := service.GetOrCreateConversation(context.Background(), 1, 2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, conversation)
	assert.Equal(t, int64(1), conversation.User1ID)
	assert.Equal(t, int64(2), conversation.User2ID)

	mockConversationRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestGetOrCreateConversation_ExistingConversation(t *testing.T) {
	// Setup mocks
	mockConversationRepo := new(MockConversationRepository)
	mockMessageRepo := new(MockMessageRepository)
	mockUserRepo := new(MockUserRepository)

	// Create service
	service := NewMessageService(mockConversationRepo, mockMessageRepo, mockUserRepo)

	// Existing conversation
	existingConv := &models.Conversation{
		ID:            1,
		User1ID:       1,
		User2ID:       2,
		LastMessageAt: time.Now(),
		CreatedAt:     time.Now(),
	}

	// Setup expectations
	mockConversationRepo.On("FindByUsers", mock.Anything, int64(1), int64(2)).Return(existingConv, nil)

	// Test
	conversation, err := service.GetOrCreateConversation(context.Background(), 1, 2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, conversation)
	assert.Equal(t, int64(1), conversation.ID)
	assert.Equal(t, int64(1), conversation.User1ID)
	assert.Equal(t, int64(2), conversation.User2ID)

	mockConversationRepo.AssertExpectations(t)
}

func TestGetOrCreateConversation_CannotMessageSelf(t *testing.T) {
	// Setup mocks
	mockConversationRepo := new(MockConversationRepository)
	mockMessageRepo := new(MockMessageRepository)
	mockUserRepo := new(MockUserRepository)

	// Create service
	service := NewMessageService(mockConversationRepo, mockMessageRepo, mockUserRepo)

	// Test
	conversation, err := service.GetOrCreateConversation(context.Background(), 1, 1)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, conversation)
	assert.Equal(t, ErrCannotMessageSelf, err)
}

func TestSendMessage_Success(t *testing.T) {
	// Setup mocks
	mockConversationRepo := new(MockConversationRepository)
	mockMessageRepo := new(MockMessageRepository)
	mockUserRepo := new(MockUserRepository)

	// Create service
	service := NewMessageService(mockConversationRepo, mockMessageRepo, mockUserRepo)

	// Existing conversation
	existingConv := &models.Conversation{
		ID:      1,
		User1ID: 1,
		User2ID: 2,
	}

	// Setup expectations
	mockConversationRepo.On("FindByID", mock.Anything, int64(1)).Return(existingConv, nil)
	mockMessageRepo.On("Create", mock.Anything, mock.MatchedBy(func(m *models.Message) bool {
		return m.ConversationID == 1 && m.SenderID == 1 && m.Content == "Hello"
	})).Return(nil)
	mockConversationRepo.On("UpdateLastMessageAt", mock.Anything, int64(1)).Return(nil)

	// Test
	req := SendMessageRequest{
		SenderID:       1,
		ConversationID: 1,
		ContentType:    "text",
		Content:        "Hello",
	}

	message, err := service.SendMessage(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, int64(1), message.ConversationID)
	assert.Equal(t, int64(1), message.SenderID)
	assert.Equal(t, "Hello", message.Content)
	assert.Equal(t, "text", message.ContentType)
	assert.False(t, message.IsRead)

	mockConversationRepo.AssertExpectations(t)
	mockMessageRepo.AssertExpectations(t)
}

func TestSendMessage_EmptyContent(t *testing.T) {
	// Setup mocks
	mockConversationRepo := new(MockConversationRepository)
	mockMessageRepo := new(MockMessageRepository)
	mockUserRepo := new(MockUserRepository)

	// Create service
	service := NewMessageService(mockConversationRepo, mockMessageRepo, mockUserRepo)

	// Test
	req := SendMessageRequest{
		SenderID:       1,
		ConversationID: 1,
		Content:        "",
	}

	message, err := service.SendMessage(context.Background(), req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, message)
	assert.Equal(t, ErrInvalidMessageContent, err)
}

func TestSendMessage_UnauthorizedSender(t *testing.T) {
	// Setup mocks
	mockConversationRepo := new(MockConversationRepository)
	mockMessageRepo := new(MockMessageRepository)
	mockUserRepo := new(MockUserRepository)

	// Create service
	service := NewMessageService(mockConversationRepo, mockMessageRepo, mockUserRepo)

	// Existing conversation between user 1 and 2
	existingConv := &models.Conversation{
		ID:      1,
		User1ID: 1,
		User2ID: 2,
	}

	// Setup expectations
	mockConversationRepo.On("FindByID", mock.Anything, int64(1)).Return(existingConv, nil)

	// Test - user 3 tries to send message
	req := SendMessageRequest{
		SenderID:       3,
		ConversationID: 1,
		Content:        "Hello",
	}

	message, err := service.SendMessage(context.Background(), req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, message)
	assert.Equal(t, ErrUnauthorizedConversation, err)

	mockConversationRepo.AssertExpectations(t)
}

func TestMarkConversationAsRead_Success(t *testing.T) {
	// Setup mocks
	mockConversationRepo := new(MockConversationRepository)
	mockMessageRepo := new(MockMessageRepository)
	mockUserRepo := new(MockUserRepository)

	// Create service
	service := NewMessageService(mockConversationRepo, mockMessageRepo, mockUserRepo)

	// Existing conversation
	existingConv := &models.Conversation{
		ID:      1,
		User1ID: 1,
		User2ID: 2,
	}

	// Setup expectations
	mockConversationRepo.On("FindByID", mock.Anything, int64(1)).Return(existingConv, nil)
	mockMessageRepo.On("MarkConversationAsRead", mock.Anything, int64(1), int64(2)).Return(nil)

	// Test - user 2 marks conversation as read
	err := service.MarkConversationAsRead(context.Background(), 2, 1)

	// Assertions
	assert.NoError(t, err)

	mockConversationRepo.AssertExpectations(t)
	mockMessageRepo.AssertExpectations(t)
}

func TestMarkConversationAsRead_Unauthorized(t *testing.T) {
	// Setup mocks
	mockConversationRepo := new(MockConversationRepository)
	mockMessageRepo := new(MockMessageRepository)
	mockUserRepo := new(MockUserRepository)

	// Create service
	service := NewMessageService(mockConversationRepo, mockMessageRepo, mockUserRepo)

	// Existing conversation between user 1 and 2
	existingConv := &models.Conversation{
		ID:      1,
		User1ID: 1,
		User2ID: 2,
	}

	// Setup expectations
	mockConversationRepo.On("FindByID", mock.Anything, int64(1)).Return(existingConv, nil)

	// Test - user 3 tries to mark conversation as read
	err := service.MarkConversationAsRead(context.Background(), 3, 1)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorizedConversation, err)

	mockConversationRepo.AssertExpectations(t)
}
