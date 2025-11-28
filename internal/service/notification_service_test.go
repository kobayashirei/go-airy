package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kobayashirei/airy/internal/models"
)

func TestCreateNotification_Success(t *testing.T) {
	// Setup mocks
	mockNotificationRepo := new(MockNotificationRepository)
	mockUserRepo := new(MockUserRepository)
	mockPostRepo := new(MockPostRepository)
	mockCommentRepo := new(MockCommentRepository)

	// Create service
	service := NewNotificationService(mockNotificationRepo, mockUserRepo, mockPostRepo, mockCommentRepo)

	// Setup expectations
	mockNotificationRepo.On("Create", mock.Anything, mock.MatchedBy(func(n *models.Notification) bool {
		return n.ReceiverID == 1 && n.Type == "comment" && n.Content != ""
	})).Return(nil)

	// Test
	req := CreateNotificationRequest{
		ReceiverID:    1,
		TriggerUserID: int64Ptr(2),
		Type:          "comment",
		EntityType:    "post",
		EntityID:      int64Ptr(10),
		Content:       "User commented on your post",
	}

	notification, err := service.CreateNotification(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, notification)
	assert.Equal(t, int64(1), notification.ReceiverID)
	assert.Equal(t, "comment", notification.Type)
	assert.Equal(t, "User commented on your post", notification.Content)
	assert.False(t, notification.IsRead)

	mockNotificationRepo.AssertExpectations(t)
}

func TestCreateNotification_InvalidType(t *testing.T) {
	// Setup mocks
	mockNotificationRepo := new(MockNotificationRepository)
	mockUserRepo := new(MockUserRepository)
	mockPostRepo := new(MockPostRepository)
	mockCommentRepo := new(MockCommentRepository)

	// Create service
	service := NewNotificationService(mockNotificationRepo, mockUserRepo, mockPostRepo, mockCommentRepo)

	// Test with invalid type
	req := CreateNotificationRequest{
		ReceiverID: 1,
		Type:       "invalid_type",
		Content:    "Test notification",
	}

	notification, err := service.CreateNotification(context.Background(), req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, notification)
	assert.Contains(t, err.Error(), "invalid notification type")
}

func TestGetNotifications_Success(t *testing.T) {
	// Setup mocks
	mockNotificationRepo := new(MockNotificationRepository)
	mockUserRepo := new(MockUserRepository)
	mockPostRepo := new(MockPostRepository)
	mockCommentRepo := new(MockCommentRepository)

	// Create service
	service := NewNotificationService(mockNotificationRepo, mockUserRepo, mockPostRepo, mockCommentRepo)

	// Create test notifications
	notifications := []*models.Notification{
		{
			ID:         1,
			ReceiverID: 1,
			Type:       "comment",
			Content:    "User commented on your post",
			IsRead:     false,
		},
		{
			ID:         2,
			ReceiverID: 1,
			Type:       "vote",
			Content:    "User upvoted your post",
			IsRead:     true,
		},
	}

	// Setup expectations
	mockNotificationRepo.On("FindByReceiverID", mock.Anything, int64(1), 20, 0).Return(notifications, nil)
	mockNotificationRepo.On("CountUnreadByReceiverID", mock.Anything, int64(1)).Return(int64(1), nil)

	// Test
	result, err := service.GetNotifications(context.Background(), 1, 1, 20)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.Notifications))
	assert.Equal(t, int64(1), result.UnreadCount)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PageSize)

	mockNotificationRepo.AssertExpectations(t)
}

func TestMarkAsRead_Success(t *testing.T) {
	// Setup mocks
	mockNotificationRepo := new(MockNotificationRepository)
	mockUserRepo := new(MockUserRepository)
	mockPostRepo := new(MockPostRepository)
	mockCommentRepo := new(MockCommentRepository)

	// Create service
	service := NewNotificationService(mockNotificationRepo, mockUserRepo, mockPostRepo, mockCommentRepo)

	// Create test notification
	notification := &models.Notification{
		ID:         1,
		ReceiverID: 1,
		Type:       "comment",
		Content:    "Test notification",
		IsRead:     false,
	}

	// Setup expectations
	mockNotificationRepo.On("FindByID", mock.Anything, int64(1)).Return(notification, nil)
	mockNotificationRepo.On("MarkAsRead", mock.Anything, int64(1)).Return(nil)

	// Test
	err := service.MarkAsRead(context.Background(), 1, 1)

	// Assertions
	assert.NoError(t, err)

	mockNotificationRepo.AssertExpectations(t)
}

func TestMarkAsRead_Unauthorized(t *testing.T) {
	// Setup mocks
	mockNotificationRepo := new(MockNotificationRepository)
	mockUserRepo := new(MockUserRepository)
	mockPostRepo := new(MockPostRepository)
	mockCommentRepo := new(MockCommentRepository)

	// Create service
	service := NewNotificationService(mockNotificationRepo, mockUserRepo, mockPostRepo, mockCommentRepo)

	// Create test notification belonging to user 1
	notification := &models.Notification{
		ID:         1,
		ReceiverID: 1,
		Type:       "comment",
		Content:    "Test notification",
		IsRead:     false,
	}

	// Setup expectations
	mockNotificationRepo.On("FindByID", mock.Anything, int64(1)).Return(notification, nil)

	// Test - user 2 trying to mark user 1's notification as read
	err := service.MarkAsRead(context.Background(), 2, 1)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorizedNotification, err)

	mockNotificationRepo.AssertExpectations(t)
}

func TestMarkAsRead_NotFound(t *testing.T) {
	// Setup mocks
	mockNotificationRepo := new(MockNotificationRepository)
	mockUserRepo := new(MockUserRepository)
	mockPostRepo := new(MockPostRepository)
	mockCommentRepo := new(MockCommentRepository)

	// Create service
	service := NewNotificationService(mockNotificationRepo, mockUserRepo, mockPostRepo, mockCommentRepo)

	// Setup expectations - notification not found
	mockNotificationRepo.On("FindByID", mock.Anything, int64(999)).Return(nil, nil)

	// Test
	err := service.MarkAsRead(context.Background(), 1, 999)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrNotificationNotFound, err)

	mockNotificationRepo.AssertExpectations(t)
}

func TestMarkAllAsRead_Success(t *testing.T) {
	// Setup mocks
	mockNotificationRepo := new(MockNotificationRepository)
	mockUserRepo := new(MockUserRepository)
	mockPostRepo := new(MockPostRepository)
	mockCommentRepo := new(MockCommentRepository)

	// Create service
	service := NewNotificationService(mockNotificationRepo, mockUserRepo, mockPostRepo, mockCommentRepo)

	// Setup expectations
	mockNotificationRepo.On("MarkAllAsRead", mock.Anything, int64(1)).Return(nil)

	// Test
	err := service.MarkAllAsRead(context.Background(), 1)

	// Assertions
	assert.NoError(t, err)

	mockNotificationRepo.AssertExpectations(t)
}

func TestGetUnreadCount_Success(t *testing.T) {
	// Setup mocks
	mockNotificationRepo := new(MockNotificationRepository)
	mockUserRepo := new(MockUserRepository)
	mockPostRepo := new(MockPostRepository)
	mockCommentRepo := new(MockCommentRepository)

	// Create service
	service := NewNotificationService(mockNotificationRepo, mockUserRepo, mockPostRepo, mockCommentRepo)

	// Setup expectations
	mockNotificationRepo.On("CountUnreadByReceiverID", mock.Anything, int64(1)).Return(int64(5), nil)

	// Test
	count, err := service.GetUnreadCount(context.Background(), 1)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)

	mockNotificationRepo.AssertExpectations(t)
}

// Helper function to create int64 pointer
func int64Ptr(i int64) *int64 {
	return &i
}
