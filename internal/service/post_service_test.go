package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kobayashirei/airy/internal/models"
)

// MockCacheService is a mock implementation of cache.Service
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheService) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	args := m.Called(ctx, key, value, expiration)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockCacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}

// MockContentModerationService is a mock implementation of ContentModerationService
type MockContentModerationService struct {
	mock.Mock
}

func (m *MockContentModerationService) CheckContent(ctx context.Context, content string) (*ModerationResult, error) {
	args := m.Called(ctx, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ModerationResult), args.Error(1)
}

func TestCreatePost_Success(t *testing.T) {
	// Setup mocks
	mockPostRepo := new(MockPostRepository)
	mockCache := new(MockCacheService)
	mockModeration := new(MockContentModerationService)
	mockMQ := new(MockMessageQueue)

	// Create service
	service := NewPostService(mockPostRepo, mockCache, mockModeration, mockMQ, nil)

	// Setup expectations
	mockModeration.On("CheckContent", mock.Anything, mock.Anything).Return(&ModerationResult{
		Status: "pass",
		Reason: "content passed moderation",
	}, nil)

	mockPostRepo.On("Create", mock.Anything, mock.MatchedBy(func(post *models.Post) bool {
		return post.Title == "Test Post" && post.Status == "published"
	})).Return(nil)

	mockMQ.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Test
	req := CreatePostRequest{
		Title:        "Test Post",
		Content:      "This is a test post content",
		AuthorID:     1,
		AllowComment: true,
	}

	post, err := service.CreatePost(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "Test Post", post.Title)
	assert.Equal(t, "published", post.Status)
	assert.NotEmpty(t, post.ContentHTML)
	assert.NotNil(t, post.PublishedAt)

	mockModeration.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

func TestCreatePost_ModerationReview(t *testing.T) {
	// Setup mocks
	mockPostRepo := new(MockPostRepository)
	mockCache := new(MockCacheService)
	mockModeration := new(MockContentModerationService)

	// Create service
	service := NewPostService(mockPostRepo, mockCache, mockModeration, nil, nil)

	// Setup expectations - content flagged for review
	mockModeration.On("CheckContent", mock.Anything, mock.Anything).Return(&ModerationResult{
		Status:   "review",
		Reason:   "content contains suspicious keywords",
		Keywords: []string{"advertisement"},
	}, nil)

	mockPostRepo.On("Create", mock.Anything, mock.MatchedBy(func(post *models.Post) bool {
		return post.Status == "pending"
	})).Return(nil)

	// Test
	req := CreatePostRequest{
		Title:    "Test Post",
		Content:  "This is an advertisement",
		AuthorID: 1,
	}

	post, err := service.CreatePost(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "pending", post.Status)
	assert.Nil(t, post.PublishedAt) // Should not be published

	mockModeration.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

func TestCreatePost_ModerationReject(t *testing.T) {
	// Setup mocks
	mockPostRepo := new(MockPostRepository)
	mockCache := new(MockCacheService)
	mockModeration := new(MockContentModerationService)

	// Create service
	service := NewPostService(mockPostRepo, mockCache, mockModeration, nil, nil)

	// Setup expectations - content rejected
	mockModeration.On("CheckContent", mock.Anything, mock.Anything).Return(&ModerationResult{
		Status:   "reject",
		Reason:   "content contains banned keywords",
		Keywords: []string{"spam"},
	}, nil)

	mockPostRepo.On("Create", mock.Anything, mock.MatchedBy(func(post *models.Post) bool {
		return post.Status == "hidden"
	})).Return(nil)

	// Test
	req := CreatePostRequest{
		Title:    "Test Post",
		Content:  "This is spam content",
		AuthorID: 1,
	}

	post, err := service.CreatePost(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "hidden", post.Status)
	assert.Nil(t, post.PublishedAt)

	mockModeration.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

func TestMapModerationStatusToPostStatus(t *testing.T) {
	tests := []struct {
		name             string
		moderationStatus string
		expectedStatus   string
	}{
		{"Pass maps to published", "pass", "published"},
		{"Review maps to pending", "review", "pending"},
		{"Reject maps to hidden", "reject", "hidden"},
		{"Unknown maps to pending", "unknown", "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapModerationStatusToPostStatus(tt.moderationStatus)
			assert.Equal(t, tt.expectedStatus, result)
		})
	}
}

func TestMarkdownToHTML(t *testing.T) {
	// Setup
	mockPostRepo := new(MockPostRepository)
	mockCache := new(MockCacheService)
	mockModeration := new(MockContentModerationService)

	service := NewPostService(mockPostRepo, mockCache, mockModeration, nil, nil).(*postService)

	// Test markdown conversion
	markdown := "# Heading\n\nThis is **bold** text."
	html := service.markdownToHTML(markdown)

	assert.Contains(t, html, "<h1>")
	assert.Contains(t, html, "<strong>")
}
