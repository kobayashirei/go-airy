package service

import (
	"context"
	"testing"
	"time"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCalculateRedditHotness(t *testing.T) {
	// Create a hotness service with Reddit algorithm
	service := &hotnessService{
		algorithm: AlgorithmReddit,
	}

	// Test case 1: New post with positive score
	now := time.Now()
	post := &models.Post{
		ID:          1,
		Title:       "Test Post",
		CreatedAt:   now,
		PublishedAt: &now,
	}
	counts := &models.EntityCount{
		EntityType:    "post",
		EntityID:      1,
		UpvoteCount:   10,
		DownvoteCount: 2,
	}

	score := service.calculateRedditHotness(post, counts)
	assert.Greater(t, score, 0.0, "Hotness score should be positive for upvoted post")

	// Test case 2: Post with negative score
	counts2 := &models.EntityCount{
		EntityType:    "post",
		EntityID:      2,
		UpvoteCount:   2,
		DownvoteCount: 10,
	}

	score2 := service.calculateRedditHotness(post, counts2)
	assert.Less(t, score2, score, "Post with negative score should have lower hotness")

	// Test case 3: Old post should have lower hotness than new post with same score
	oldTime := now.Add(-24 * time.Hour)
	oldPost := &models.Post{
		ID:          3,
		Title:       "Old Post",
		CreatedAt:   oldTime,
		PublishedAt: &oldTime,
	}

	oldScore := service.calculateRedditHotness(oldPost, counts)
	assert.Less(t, oldScore, score, "Older post should have lower hotness than newer post")
}

func TestCalculateHackerNewsHotness(t *testing.T) {
	// Create a hotness service with Hacker News algorithm
	service := &hotnessService{
		algorithm: AlgorithmHackerNews,
	}

	// Test case 1: New post with positive score
	now := time.Now()
	post := &models.Post{
		ID:          1,
		Title:       "Test Post",
		CreatedAt:   now,
		PublishedAt: &now,
	}
	counts := &models.EntityCount{
		EntityType:    "post",
		EntityID:      1,
		UpvoteCount:   10,
		DownvoteCount: 2,
	}

	score := service.calculateHackerNewsHotness(post, counts)
	assert.Greater(t, score, 0.0, "Hotness score should be positive for upvoted post")

	// Test case 2: Post with more upvotes should have higher hotness
	counts2 := &models.EntityCount{
		EntityType:    "post",
		EntityID:      2,
		UpvoteCount:   20,
		DownvoteCount: 2,
	}

	score2 := service.calculateHackerNewsHotness(post, counts2)
	assert.Greater(t, score2, score, "Post with more upvotes should have higher hotness")

	// Test case 3: Old post should have lower hotness than new post with same score
	oldTime := now.Add(-24 * time.Hour)
	oldPost := &models.Post{
		ID:          3,
		Title:       "Old Post",
		CreatedAt:   oldTime,
		PublishedAt: &oldTime,
	}

	oldScore := service.calculateHackerNewsHotness(oldPost, counts)
	assert.Less(t, oldScore, score, "Older post should have lower hotness than newer post")
}

func TestCalculateHotness_AlgorithmSelection(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	post := &models.Post{
		ID:          1,
		Title:       "Test Post",
		CreatedAt:   now,
		PublishedAt: &now,
	}
	counts := &models.EntityCount{
		EntityType:    "post",
		EntityID:      1,
		UpvoteCount:   10,
		DownvoteCount: 2,
	}

	// Test Reddit algorithm
	redditService := &hotnessService{
		algorithm: AlgorithmReddit,
	}
	redditScore, err := redditService.CalculateHotness(ctx, post, counts)
	assert.NoError(t, err)
	assert.Greater(t, redditScore, 0.0)

	// Test Hacker News algorithm
	hnService := &hotnessService{
		algorithm: AlgorithmHackerNews,
	}
	hnScore, err := hnService.CalculateHotness(ctx, post, counts)
	assert.NoError(t, err)
	assert.Greater(t, hnScore, 0.0)

	// Scores should be different (algorithms are different)
	assert.NotEqual(t, redditScore, hnScore, "Different algorithms should produce different scores")
}

func TestCalculateHotness_InvalidAlgorithm(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	post := &models.Post{
		ID:          1,
		Title:       "Test Post",
		CreatedAt:   now,
		PublishedAt: &now,
	}
	counts := &models.EntityCount{
		EntityType:  "post",
		EntityID:    1,
		UpvoteCount: 10,
	}

	// Test with invalid algorithm
	service := &hotnessService{
		algorithm: "invalid",
	}
	_, err := service.CalculateHotness(ctx, post, counts)
	assert.Error(t, err, "Should return error for invalid algorithm")
}

func TestCalculateHotness_NilPost(t *testing.T) {
	ctx := context.Background()
	service := &hotnessService{
		algorithm: AlgorithmReddit,
	}
	counts := &models.EntityCount{
		EntityType:  "post",
		EntityID:    1,
		UpvoteCount: 10,
	}

	_, err := service.CalculateHotness(ctx, nil, counts)
	assert.Error(t, err, "Should return error for nil post")
}

func TestCalculateHotness_ZeroScore(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	post := &models.Post{
		ID:          1,
		Title:       "Test Post",
		CreatedAt:   now,
		PublishedAt: &now,
	}
	counts := &models.EntityCount{
		EntityType:    "post",
		EntityID:      1,
		UpvoteCount:   5,
		DownvoteCount: 5,
	}

	// Test Reddit algorithm with zero score
	redditService := &hotnessService{
		algorithm: AlgorithmReddit,
	}
	score, err := redditService.CalculateHotness(ctx, post, counts)
	assert.NoError(t, err)
	// Reddit algorithm should still produce a score based on time
	// The score can be 0 when upvotes equal downvotes, but the calculation should succeed
	_ = score

	// Test Hacker News algorithm with zero score
	hnService := &hotnessService{
		algorithm: AlgorithmHackerNews,
	}
	hnScore, err := hnService.CalculateHotness(ctx, post, counts)
	assert.NoError(t, err)
	// HN algorithm should produce a non-negative score
	assert.GreaterOrEqual(t, hnScore, 0.0)
}
