package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/repository"
)

// TestFeedService_GetCircleFeed tests retrieving circle feed
func TestFeedService_GetCircleFeed(t *testing.T) {
	// This test verifies that GetCircleFeed correctly queries posts from a specific circle
	// and returns them in the expected format
	
	t.Run("successfully retrieves circle feed", func(t *testing.T) {
		// Setup
		mockPostRepo := new(MockPostRepository)
		mockUserProfileRepo := new(MockUserProfileRepository)
		mockCacheService := new(MockCacheService)
		
		// We don't need a real Redis client for this test since GetCircleFeed
		// queries directly from the database
		feedService := &feedService{
			redisClient:     nil,
			postRepo:        mockPostRepo,
			userProfileRepo: mockUserProfileRepo,
			cacheService:    mockCacheService,
		}
		
		ctx := context.Background()
		circleID := int64(999)
		limit := 20
		offset := 0
		sortBy := "created_at"
		
		// Mock post repository to return circle posts
		expectedPosts := []*models.Post{
			{
				ID:        10,
				Title:     "Circle Post 1",
				AuthorID:  123,
				CircleID:  &circleID,
				Status:    "published",
				CreatedAt: time.Now(),
			},
			{
				ID:        11,
				Title:     "Circle Post 2",
				AuthorID:  456,
				CircleID:  &circleID,
				Status:    "published",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
		}
		
		mockPostRepo.On("List", ctx, mock.MatchedBy(func(opts repository.PostListOptions) bool {
			return opts.CircleID != nil && *opts.CircleID == circleID &&
				opts.Status == "published" &&
				opts.SortBy == sortBy &&
				opts.Limit == limit &&
				opts.Offset == offset
		})).Return(expectedPosts, nil)
		
		// Execute
		posts, err := feedService.GetCircleFeed(ctx, circleID, limit, offset, sortBy)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, posts)
		assert.Len(t, posts, 2)
		assert.Equal(t, circleID, *posts[0].CircleID)
		assert.Equal(t, circleID, *posts[1].CircleID)
		mockPostRepo.AssertExpectations(t)
	})
}


