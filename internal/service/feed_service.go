package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/kobayashirei/airy/internal/cache"
	appLogger "github.com/kobayashirei/airy/internal/logger"
	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/repository"
)

const (
	// FanoutThreshold is the follower count threshold for choosing feed distribution strategy
	FanoutThreshold = 1000
	
	// FeedExpiration is the TTL for feed entries in Redis
	FeedExpiration = 7 * 24 * time.Hour // 7 days
	
	// MaxFeedSize is the maximum number of posts to keep in a user's feed
	MaxFeedSize = 1000
)

// FeedService defines the interface for feed operations
type FeedService interface {
	// PushToFollowerFeeds pushes a post to followers' feeds (fan-out write)
	PushToFollowerFeeds(ctx context.Context, postID int64, authorID int64) error
	
	// GetUserFeed retrieves a user's personalized feed
	GetUserFeed(ctx context.Context, userID int64, limit int, offset int, sortBy string) ([]*models.Post, error)
	
	// GetCircleFeed retrieves posts from a specific circle
	GetCircleFeed(ctx context.Context, circleID int64, limit int, offset int, sortBy string) ([]*models.Post, error)
	
	// RemoveFromFeeds removes a post from all feeds (when deleted)
	RemoveFromFeeds(ctx context.Context, postID int64) error
}

// feedService implements FeedService interface
type feedService struct {
	redisClient       *redis.Client
	postRepo          repository.PostRepository
	userProfileRepo   repository.UserProfileRepository
	cacheService      cache.Service
}

// NewFeedService creates a new feed service
func NewFeedService(
	redisClient *redis.Client,
	postRepo repository.PostRepository,
	userProfileRepo repository.UserProfileRepository,
	cacheService cache.Service,
) FeedService {
	return &feedService{
		redisClient:     redisClient,
		postRepo:        postRepo,
		userProfileRepo: userProfileRepo,
		cacheService:    cacheService,
	}
}

// PushToFollowerFeeds pushes a post to followers' feeds using fan-out write strategy
func (s *feedService) PushToFollowerFeeds(ctx context.Context, postID int64, authorID int64) error {
	// Get author's follower count to determine strategy
	profile, err := s.userProfileRepo.FindByUserID(ctx, authorID)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}
	
	followerCount := 0
	if profile != nil {
		followerCount = profile.FollowerCount
	}
	
	// If author has too many followers, use fan-in read strategy (do nothing here)
	if followerCount > FanoutThreshold {
		appLogger.Info("Skipping fan-out for high-follower user",
			zap.Int64("author_id", authorID),
			zap.Int("follower_count", followerCount),
			zap.Int("threshold", FanoutThreshold),
		)
		return nil
	}
	
	// Fan-out write: push post to all followers' feeds
	// In a real implementation, we would query followers from a follows table
	// For now, we'll use a placeholder that can be extended
	followerIDs, err := s.getFollowerIDs(ctx, authorID)
	if err != nil {
		return fmt.Errorf("failed to get follower IDs: %w", err)
	}
	
	// Use pipeline for efficient batch operations
	pipe := s.redisClient.Pipeline()
	timestamp := float64(time.Now().Unix())
	
	for _, followerID := range followerIDs {
		feedKey := cache.GetUserFeedKey(followerID)
		
		// Add post to follower's feed (sorted set with timestamp as score)
		pipe.ZAdd(ctx, feedKey, redis.Z{
			Score:  timestamp,
			Member: postID,
		})
		
		// Trim feed to max size (keep most recent posts)
		pipe.ZRemRangeByRank(ctx, feedKey, 0, -MaxFeedSize-1)
		
		// Set expiration
		pipe.Expire(ctx, feedKey, FeedExpiration)
	}
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to push to follower feeds: %w", err)
	}
	
	appLogger.Info("Pushed post to follower feeds",
		zap.Int64("post_id", postID),
		zap.Int64("author_id", authorID),
		zap.Int("follower_count", len(followerIDs)),
	)
	
	return nil
}

// GetUserFeed retrieves a user's personalized feed
func (s *feedService) GetUserFeed(ctx context.Context, userID int64, limit int, offset int, sortBy string) ([]*models.Post, error) {
	// Try to get feed from Redis first (fan-out write result)
	feedKey := cache.GetUserFeedKey(userID)
	
	var postIDs []int64
	var err error
	
	if sortBy == "hotness_score" {
		// For hotness sorting, we need to fetch from database
		postIDs, err = s.getFeedFromDatabase(ctx, userID, limit, offset, sortBy)
	} else {
		// For time-based sorting, use Redis sorted set
		postIDs, err = s.getFeedFromRedis(ctx, feedKey, limit, offset)
		
		// If Redis feed is empty or error, fall back to database (fan-in read)
		if err != nil || len(postIDs) == 0 {
			appLogger.Info("Falling back to database for user feed",
				zap.Int64("user_id", userID),
				zap.Error(err),
			)
			postIDs, err = s.getFeedFromDatabase(ctx, userID, limit, offset, sortBy)
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get feed post IDs: %w", err)
	}
	
	// Batch query post details with cache
	posts, err := s.batchGetPosts(ctx, postIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get post details: %w", err)
	}
	
	return posts, nil
}

// GetCircleFeed retrieves posts from a specific circle
func (s *feedService) GetCircleFeed(ctx context.Context, circleID int64, limit int, offset int, sortBy string) ([]*models.Post, error) {
	// Query posts from the circle
	opts := repository.PostListOptions{
		CircleID: &circleID,
		Status:   "published",
		SortBy:   sortBy,
		Order:    "DESC",
		Limit:    limit,
		Offset:   offset,
	}
	
	posts, err := s.postRepo.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get circle posts: %w", err)
	}
	
	return posts, nil
}

// RemoveFromFeeds removes a post from all feeds
func (s *feedService) RemoveFromFeeds(ctx context.Context, postID int64) error {
	// In a production system, we would need to track which feeds contain this post
	// For now, this is a placeholder that can be extended
	appLogger.Info("Removing post from feeds", zap.Int64("post_id", postID))
	
	// This would require maintaining a reverse index or scanning feeds
	// Implementation depends on scale and requirements
	
	return nil
}

// getFeedFromRedis retrieves post IDs from Redis sorted set
func (s *feedService) getFeedFromRedis(ctx context.Context, feedKey string, limit int, offset int) ([]int64, error) {
	// Get post IDs from sorted set (most recent first)
	start := int64(offset)
	stop := int64(offset + limit - 1)
	
	results, err := s.redisClient.ZRevRange(ctx, feedKey, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get feed from Redis: %w", err)
	}
	
	postIDs := make([]int64, 0, len(results))
	for _, result := range results {
		postID, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			appLogger.Warn("Invalid post ID in feed", zap.String("value", result))
			continue
		}
		postIDs = append(postIDs, postID)
	}
	
	return postIDs, nil
}

// getFeedFromDatabase retrieves feed using fan-in read strategy
func (s *feedService) getFeedFromDatabase(ctx context.Context, userID int64, limit int, offset int, sortBy string) ([]int64, error) {
	// In a real implementation, we would:
	// 1. Get list of users that this user follows
	// 2. Query posts from those users
	// 3. Merge and sort the results
	
	// For now, we'll return recent published posts as a placeholder
	opts := repository.PostListOptions{
		Status: "published",
		SortBy: sortBy,
		Order:  "DESC",
		Limit:  limit,
		Offset: offset,
	}
	
	posts, err := s.postRepo.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	
	postIDs := make([]int64, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
	}
	
	return postIDs, nil
}

// batchGetPosts retrieves multiple posts with caching
func (s *feedService) batchGetPosts(ctx context.Context, postIDs []int64) ([]*models.Post, error) {
	posts := make([]*models.Post, 0, len(postIDs))
	
	for _, postID := range postIDs {
		// Try cache first
		var post models.Post
		cacheKey := cache.GetPostKey(postID)
		err := s.cacheService.Get(ctx, cacheKey, &post)
		
		if err == nil {
			// Cache hit
			posts = append(posts, &post)
			continue
		}
		
		// Cache miss, query from database
		dbPost, err := s.postRepo.FindByID(ctx, postID)
		if err != nil {
			appLogger.Error("Failed to get post", zap.Int64("post_id", postID), zap.Error(err))
			continue
		}
		
		if dbPost == nil {
			appLogger.Warn("Post not found", zap.Int64("post_id", postID))
			continue
		}
		
		// Store in cache
		_ = s.cacheService.Set(ctx, cacheKey, dbPost, 1*time.Hour)
		
		posts = append(posts, dbPost)
	}
	
	return posts, nil
}

// getFollowerIDs retrieves the list of follower IDs for a user
// This is a placeholder - in a real implementation, this would query a follows table
func (s *feedService) getFollowerIDs(ctx context.Context, userID int64) ([]int64, error) {
	// TODO: Implement actual follower query from follows table
	// For now, return empty list
	return []int64{}, nil
}
