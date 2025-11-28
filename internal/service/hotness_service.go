package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/repository"
)

// HotnessAlgorithm defines the algorithm type for hotness calculation
type HotnessAlgorithm string

const (
	// AlgorithmReddit uses Reddit's hot ranking algorithm
	AlgorithmReddit HotnessAlgorithm = "reddit"
	// AlgorithmHackerNews uses Hacker News' ranking algorithm
	AlgorithmHackerNews HotnessAlgorithm = "hackernews"
)

// HotnessService defines the interface for hotness calculation
type HotnessService interface {
	CalculateHotness(ctx context.Context, post *models.Post, counts *models.EntityCount) (float64, error)
	RecalculatePostHotness(ctx context.Context, postID int64) (float64, error)
}

// hotnessService implements HotnessService interface
type hotnessService struct {
	postRepo        repository.PostRepository
	entityCountRepo repository.EntityCountRepository
	algorithm       HotnessAlgorithm
}

// NewHotnessService creates a new hotness service
// Implements Requirements 12.1, 12.2, 12.3
func NewHotnessService(
	postRepo repository.PostRepository,
	entityCountRepo repository.EntityCountRepository,
	algorithm HotnessAlgorithm,
) HotnessService {
	// Default to Reddit algorithm if not specified
	if algorithm == "" {
		algorithm = AlgorithmReddit
	}

	return &hotnessService{
		postRepo:        postRepo,
		entityCountRepo: entityCountRepo,
		algorithm:       algorithm,
	}
}

// CalculateHotness calculates the hotness score for a post
// Implements Requirement 12.3
func (s *hotnessService) CalculateHotness(ctx context.Context, post *models.Post, counts *models.EntityCount) (float64, error) {
	if post == nil {
		return 0, fmt.Errorf("post cannot be nil")
	}

	// Use the configured algorithm
	switch s.algorithm {
	case AlgorithmReddit:
		return s.calculateRedditHotness(post, counts), nil
	case AlgorithmHackerNews:
		return s.calculateHackerNewsHotness(post, counts), nil
	default:
		return 0, fmt.Errorf("unknown hotness algorithm: %s", s.algorithm)
	}
}

// RecalculatePostHotness recalculates and updates the hotness score for a post
// Implements Requirement 12.2
func (s *hotnessService) RecalculatePostHotness(ctx context.Context, postID int64) (float64, error) {
	// Get post
	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return 0, fmt.Errorf("failed to find post: %w", err)
	}
	if post == nil {
		return 0, fmt.Errorf("post not found: %d", postID)
	}

	// Get entity counts
	counts, err := s.entityCountRepo.FindByEntity(ctx, "post", postID)
	if err != nil {
		return 0, fmt.Errorf("failed to find entity counts: %w", err)
	}

	// If no counts exist, create default
	if counts == nil {
		counts = &models.EntityCount{
			EntityType:    "post",
			EntityID:      postID,
			UpvoteCount:   0,
			DownvoteCount: 0,
			CommentCount:  0,
			FavoriteCount: 0,
		}
	}

	// Calculate new hotness score
	newScore, err := s.CalculateHotness(ctx, post, counts)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate hotness: %w", err)
	}

	// Update post hotness score using the dedicated method
	if err := s.postRepo.UpdateHotnessScore(ctx, postID, newScore); err != nil {
		return 0, fmt.Errorf("failed to update post hotness: %w", err)
	}

	return newScore, nil
}

// calculateRedditHotness implements Reddit's hot ranking algorithm
// Formula: log10(max(|score|, 1)) + sign(score) * seconds / 45000
// where score = upvotes - downvotes
// and seconds = time since epoch
func (s *hotnessService) calculateRedditHotness(post *models.Post, counts *models.EntityCount) float64 {
	// Calculate score (upvotes - downvotes)
	score := counts.UpvoteCount - counts.DownvoteCount

	// Get the order of magnitude
	order := math.Log10(math.Max(math.Abs(float64(score)), 1))

	// Get sign of score
	var sign float64
	if score > 0 {
		sign = 1
	} else if score < 0 {
		sign = -1
	} else {
		sign = 0
	}

	// Get seconds since epoch
	// Use published time if available, otherwise created time
	var postTime time.Time
	if post.PublishedAt != nil {
		postTime = *post.PublishedAt
	} else {
		postTime = post.CreatedAt
	}
	seconds := float64(postTime.Unix())

	// Reddit's formula: log10(max(|score|, 1)) + sign(score) * seconds / 45000
	// The 45000 is approximately 12.5 hours in seconds
	hotness := order + sign*seconds/45000

	return hotness
}

// calculateHackerNewsHotness implements Hacker News' ranking algorithm
// Formula: (score - 1) / (age + 2)^gravity
// where score = upvotes - downvotes + 1
// age = hours since post creation
// gravity = 1.8 (controls how quickly posts fall)
func (s *hotnessService) calculateHackerNewsHotness(post *models.Post, counts *models.EntityCount) float64 {
	// Calculate score (upvotes - downvotes + 1)
	// The +1 prevents division issues and gives new posts a baseline
	score := float64(counts.UpvoteCount - counts.DownvoteCount + 1)

	// Ensure score is at least 0
	if score < 0 {
		score = 0
	}

	// Get post age in hours
	var postTime time.Time
	if post.PublishedAt != nil {
		postTime = *post.PublishedAt
	} else {
		postTime = post.CreatedAt
	}
	ageHours := time.Since(postTime).Hours()

	// Gravity factor (1.8 is HN's default)
	const gravity = 1.8

	// HN formula: (score - 1) / (age + 2)^gravity
	// The +2 prevents division by zero and gives very new posts a boost
	hotness := (score - 1) / math.Pow(ageHours+2, gravity)

	return hotness
}
