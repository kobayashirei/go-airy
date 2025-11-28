package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kobayashirei/airy/internal/mq"
	"github.com/kobayashirei/airy/internal/repository"
	"go.uber.org/zap"
)

// SearchConsumer handles search indexing events from message queue
type SearchConsumer struct {
	searchService SearchService
	postRepo      repository.PostRepository
	userRepo      repository.UserRepository
	profileRepo   repository.UserProfileRepository
	statsRepo     repository.UserStatsRepository
	log           *zap.Logger
}

// NewSearchConsumer creates a new search consumer
func NewSearchConsumer(
	searchService SearchService,
	postRepo repository.PostRepository,
	userRepo repository.UserRepository,
	profileRepo repository.UserProfileRepository,
	statsRepo repository.UserStatsRepository,
	log *zap.Logger,
) *SearchConsumer {
	return &SearchConsumer{
		searchService: searchService,
		postRepo:      postRepo,
		userRepo:      userRepo,
		profileRepo:   profileRepo,
		statsRepo:     statsRepo,
		log:           log,
	}
}

// HandlePostPublished handles post published events
func (c *SearchConsumer) HandlePostPublished(ctx context.Context, message []byte) error {
	var event mq.PostPublishedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal post published event: %w", err)
	}

	c.log.Info("Processing post published event", zap.Int64("post_id", event.PostID))

	// Get post from database
	post, err := c.postRepo.FindByID(ctx, event.PostID)
	if err != nil {
		return fmt.Errorf("failed to get post %d: %w", event.PostID, err)
	}

	// Index post in Elasticsearch
	if err := c.searchService.IndexPost(ctx, post); err != nil {
		return fmt.Errorf("failed to index post %d: %w", event.PostID, err)
	}

	c.log.Info("Successfully indexed post", zap.Int64("post_id", event.PostID))
	return nil
}

// HandlePostUpdated handles post updated events
func (c *SearchConsumer) HandlePostUpdated(ctx context.Context, message []byte) error {
	var event mq.PostUpdatedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal post updated event: %w", err)
	}

	c.log.Info("Processing post updated event", zap.Int64("post_id", event.PostID))

	// Get post from database
	post, err := c.postRepo.FindByID(ctx, event.PostID)
	if err != nil {
		return fmt.Errorf("failed to get post %d: %w", event.PostID, err)
	}

	// Update post in Elasticsearch
	if err := c.searchService.UpdatePost(ctx, event.PostID, post); err != nil {
		return fmt.Errorf("failed to update post %d: %w", event.PostID, err)
	}

	c.log.Info("Successfully updated post in search index", zap.Int64("post_id", event.PostID))
	return nil
}

// HandlePostDeleted handles post deleted events
func (c *SearchConsumer) HandlePostDeleted(ctx context.Context, message []byte) error {
	var event mq.PostDeletedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal post deleted event: %w", err)
	}

	c.log.Info("Processing post deleted event", zap.Int64("post_id", event.PostID))

	// Delete post from Elasticsearch
	if err := c.searchService.DeletePost(ctx, event.PostID); err != nil {
		return fmt.Errorf("failed to delete post %d: %w", event.PostID, err)
	}

	c.log.Info("Successfully deleted post from search index", zap.Int64("post_id", event.PostID))
	return nil
}

// HandleUserRegistered handles user registered events
func (c *SearchConsumer) HandleUserRegistered(ctx context.Context, message []byte) error {
	var event mq.UserRegisteredEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal user registered event: %w", err)
	}

	c.log.Info("Processing user registered event", zap.Int64("user_id", event.UserID))

	// Get user from database
	user, err := c.userRepo.FindByID(ctx, event.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user %d: %w", event.UserID, err)
	}

	// Get user profile
	profile, err := c.profileRepo.FindByUserID(ctx, event.UserID)
	if err != nil {
		c.log.Warn("Failed to get profile for user", zap.Int64("user_id", event.UserID), zap.Error(err))
		profile = nil
	}

	// Get user stats
	stats, err := c.statsRepo.FindByUserID(ctx, event.UserID)
	if err != nil {
		c.log.Warn("Failed to get stats for user", zap.Int64("user_id", event.UserID), zap.Error(err))
		stats = nil
	}

	// Index user in Elasticsearch
	if err := c.searchService.IndexUser(ctx, user, profile, stats); err != nil {
		return fmt.Errorf("failed to index user %d: %w", event.UserID, err)
	}

	c.log.Info("Successfully indexed user", zap.Int64("user_id", event.UserID))
	return nil
}

// Subscribe subscribes to all search-related events
func (c *SearchConsumer) Subscribe(messageQueue mq.MessageQueue) error {
	// Subscribe to post events
	if err := messageQueue.Subscribe(mq.TopicPostPublished, c.HandlePostPublished); err != nil {
		return fmt.Errorf("failed to subscribe to post published events: %w", err)
	}

	if err := messageQueue.Subscribe(mq.TopicPostUpdated, c.HandlePostUpdated); err != nil {
		return fmt.Errorf("failed to subscribe to post updated events: %w", err)
	}

	if err := messageQueue.Subscribe(mq.TopicPostDeleted, c.HandlePostDeleted); err != nil {
		return fmt.Errorf("failed to subscribe to post deleted events: %w", err)
	}

	// Subscribe to user events
	if err := messageQueue.Subscribe(mq.TopicUserRegistered, c.HandleUserRegistered); err != nil {
		return fmt.Errorf("failed to subscribe to user registered events: %w", err)
	}

	c.log.Info("Search consumer subscribed to all events")
	return nil
}
