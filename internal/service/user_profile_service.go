package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kobayashirei/airy/internal/cache"
	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/repository"
)

var (
	// ErrProfileNotFound is returned when user profile is not found
	ErrProfileNotFound = errors.New("user profile not found")
)

// UserProfileService defines the interface for user profile business logic
type UserProfileService interface {
	// GetProfile retrieves a user's complete profile (with caching)
	GetProfile(ctx context.Context, userID int64) (*UserProfileResponse, error)
	
	// UpdateProfile updates a user's profile information
	UpdateProfile(ctx context.Context, userID int64, req UpdateProfileRequest) error
	
	// GetUserPosts retrieves posts created by a user
	GetUserPosts(ctx context.Context, userID int64, opts UserPostsOptions) (*UserPostsResponse, error)
	
	// UpdateFollowerCount updates the follower count for a user
	UpdateFollowerCount(ctx context.Context, userID int64, delta int) error
	
	// UpdateFollowingCount updates the following count for a user
	UpdateFollowingCount(ctx context.Context, userID int64, delta int) error
	
	// UpdatePostCount updates the post count for a user
	UpdatePostCount(ctx context.Context, userID int64, delta int) error
	
	// UpdateCommentCount updates the comment count for a user
	UpdateCommentCount(ctx context.Context, userID int64, delta int) error
	
	// UpdateVoteReceivedCount updates the vote received count for a user
	UpdateVoteReceivedCount(ctx context.Context, userID int64, delta int) error
}

// UserProfileResponse represents a complete user profile
type UserProfileResponse struct {
	User    *models.User        `json:"user"`
	Profile *models.UserProfile `json:"profile"`
	Stats   *models.UserStats   `json:"stats"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Avatar   *string    `json:"avatar,omitempty"`
	Gender   *string    `json:"gender,omitempty"`
	Birthday *time.Time `json:"birthday,omitempty"`
	Bio      *string    `json:"bio,omitempty"`
}

// UserPostsOptions defines options for retrieving user posts
type UserPostsOptions struct {
	Status string // Filter by status (published, draft, etc.)
	SortBy string // Sort by field (created_at, hotness_score, view_count)
	Order  string // Sort order (asc, desc)
	Page   int    // Page number (1-indexed)
	Limit  int    // Number of posts per page
}

// UserPostsResponse represents the response for user posts query
type UserPostsResponse struct {
	Posts      []*models.Post `json:"posts"`
	TotalCount int64          `json:"total_count"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
}

// userProfileService implements UserProfileService interface
type userProfileService struct {
	userRepo        repository.UserRepository
	userProfileRepo repository.UserProfileRepository
	userStatsRepo   repository.UserStatsRepository
	postRepo        repository.PostRepository
	cacheService    cache.Service
}

// NewUserProfileService creates a new user profile service
func NewUserProfileService(
	userRepo repository.UserRepository,
	userProfileRepo repository.UserProfileRepository,
	userStatsRepo repository.UserStatsRepository,
	postRepo repository.PostRepository,
	cacheService cache.Service,
) UserProfileService {
	return &userProfileService{
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
		userStatsRepo:   userStatsRepo,
		postRepo:        postRepo,
		cacheService:    cacheService,
	}
}

// GetProfile retrieves a user's complete profile with caching
func (s *userProfileService) GetProfile(ctx context.Context, userID int64) (*UserProfileResponse, error) {
	// Try to get from cache first
	cacheKey := cache.UserKey(userID)
	var cachedProfile UserProfileResponse
	if err := s.cacheService.Get(ctx, cacheKey, &cachedProfile); err == nil {
		return &cachedProfile, nil
	}

	// Cache miss, query from database
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Remove password hash from response
	user.PasswordHash = ""

	// Get user profile
	profile, err := s.userProfileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user profile: %w", err)
	}
	if profile == nil {
		// Create default profile if not exists
		profile = &models.UserProfile{
			UserID:         userID,
			Points:         0,
			Level:          1,
			FollowerCount:  0,
			FollowingCount: 0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if err := s.userProfileRepo.Create(ctx, profile); err != nil {
			return nil, fmt.Errorf("failed to create user profile: %w", err)
		}
	}

	// Get user stats
	stats, err := s.userStatsRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user stats: %w", err)
	}
	if stats == nil {
		// Create default stats if not exists
		stats = &models.UserStats{
			UserID:            userID,
			PostCount:         0,
			CommentCount:      0,
			VoteReceivedCount: 0,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
		if err := s.userStatsRepo.Create(ctx, stats); err != nil {
			return nil, fmt.Errorf("failed to create user stats: %w", err)
		}
	}

	response := &UserProfileResponse{
		User:    user,
		Profile: profile,
		Stats:   stats,
	}

	// Store in cache with 5 minute expiration
	if err := s.cacheService.Set(ctx, cacheKey, response, 5*time.Minute); err != nil {
		// Log error but don't fail the request
		fmt.Printf("failed to cache user profile: %v\n", err)
	}

	return response, nil
}

// UpdateProfile updates a user's profile information
func (s *userProfileService) UpdateProfile(ctx context.Context, userID int64, req UpdateProfileRequest) error {
	// Find user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Update user fields if provided
	updated := false
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
		updated = true
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
		updated = true
	}
	if req.Birthday != nil {
		user.Birthday = *req.Birthday
		updated = true
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
		updated = true
	}

	if updated {
		user.UpdatedAt = time.Now()
		if err := s.userRepo.Update(ctx, user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		// Invalidate cache
		cacheKey := cache.UserKey(userID)
		if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
			// Log error but don't fail the request
			fmt.Printf("failed to delete user cache: %v\n", err)
		}
	}

	return nil
}

// GetUserPosts retrieves posts created by a user
func (s *userProfileService) GetUserPosts(ctx context.Context, userID int64, opts UserPostsOptions) (*UserPostsResponse, error) {
	// Verify user exists
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Set defaults
	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.SortBy == "" {
		opts.SortBy = "created_at"
	}
	if opts.Order == "" {
		opts.Order = "desc"
	}
	if opts.Status == "" {
		opts.Status = "published"
	}

	// Build repository options
	repoOpts := repository.PostListOptions{
		AuthorID: &userID,
		Status:   opts.Status,
		SortBy:   opts.SortBy,
		Order:    opts.Order,
		Limit:    opts.Limit,
		Offset:   (opts.Page - 1) * opts.Limit,
	}

	// Get posts
	posts, err := s.postRepo.List(ctx, repoOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	// Get total count
	totalCount, err := s.postRepo.Count(ctx, repoOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to count posts: %w", err)
	}

	return &UserPostsResponse{
		Posts:      posts,
		TotalCount: totalCount,
		Page:       opts.Page,
		Limit:      opts.Limit,
	}, nil
}

// UpdateFollowerCount updates the follower count for a user
func (s *userProfileService) UpdateFollowerCount(ctx context.Context, userID int64, delta int) error {
	if err := s.userProfileRepo.IncrementFollowerCount(ctx, userID, delta); err != nil {
		return fmt.Errorf("failed to update follower count: %w", err)
	}

	// Invalidate cache
	cacheKey := cache.UserKey(userID)
	if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("failed to delete user cache: %v\n", err)
	}

	return nil
}

// UpdateFollowingCount updates the following count for a user
func (s *userProfileService) UpdateFollowingCount(ctx context.Context, userID int64, delta int) error {
	if err := s.userProfileRepo.IncrementFollowingCount(ctx, userID, delta); err != nil {
		return fmt.Errorf("failed to update following count: %w", err)
	}

	// Invalidate cache
	cacheKey := cache.UserKey(userID)
	if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("failed to delete user cache: %v\n", err)
	}

	return nil
}

// UpdatePostCount updates the post count for a user
func (s *userProfileService) UpdatePostCount(ctx context.Context, userID int64, delta int) error {
	if err := s.userStatsRepo.IncrementPostCount(ctx, userID, delta); err != nil {
		return fmt.Errorf("failed to update post count: %w", err)
	}

	// Invalidate cache
	cacheKey := cache.UserKey(userID)
	if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("failed to delete user cache: %v\n", err)
	}

	return nil
}

// UpdateCommentCount updates the comment count for a user
func (s *userProfileService) UpdateCommentCount(ctx context.Context, userID int64, delta int) error {
	if err := s.userStatsRepo.IncrementCommentCount(ctx, userID, delta); err != nil {
		return fmt.Errorf("failed to update comment count: %w", err)
	}

	// Invalidate cache
	cacheKey := cache.UserKey(userID)
	if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("failed to delete user cache: %v\n", err)
	}

	return nil
}

// UpdateVoteReceivedCount updates the vote received count for a user
func (s *userProfileService) UpdateVoteReceivedCount(ctx context.Context, userID int64, delta int) error {
	if err := s.userStatsRepo.IncrementVoteReceivedCount(ctx, userID, delta); err != nil {
		return fmt.Errorf("failed to update vote received count: %w", err)
	}

	// Invalidate cache
	cacheKey := cache.UserKey(userID)
	if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("failed to delete user cache: %v\n", err)
	}

	return nil
}
