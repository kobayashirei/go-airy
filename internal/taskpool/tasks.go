package taskpool

import (
	"context"
	"fmt"
)

// Common task implementations for the Airy system

// UpdateSearchIndexTask updates the search index for a post
type UpdateSearchIndexTask struct {
	PostID int64
	// SearchService would be injected here
}

func (t *UpdateSearchIndexTask) Execute(ctx context.Context) error {
	// Implementation will be added when search service is implemented
	return fmt.Errorf("search service not yet implemented")
}

// SendNotificationTask sends a notification to a user
type SendNotificationTask struct {
	UserID       int64
	Notification interface{} // Will be replaced with actual Notification type
	// NotificationService would be injected here
}

func (t *SendNotificationTask) Execute(ctx context.Context) error {
	// Implementation will be added when notification service is implemented
	return fmt.Errorf("notification service not yet implemented")
}

// UpdateFeedTask updates user feeds after a post is published
type UpdateFeedTask struct {
	PostID   int64
	AuthorID int64
	// FeedService would be injected here
}

func (t *UpdateFeedTask) Execute(ctx context.Context) error {
	// Implementation will be added when feed service is implemented
	return fmt.Errorf("feed service not yet implemented")
}

// UpdateHotnessScoreTask recalculates hotness score for a post
type UpdateHotnessScoreTask struct {
	PostID int64
	// HotnessService would be injected here
}

func (t *UpdateHotnessScoreTask) Execute(ctx context.Context) error {
	// Implementation will be added when hotness service is implemented
	return fmt.Errorf("hotness service not yet implemented")
}

// UpdateCountTask updates aggregated counts for entities
type UpdateCountTask struct {
	EntityType string
	EntityID   int64
	CountType  string // "upvote", "downvote", "comment", "favorite"
	Delta      int
	// CountService would be injected here
}

func (t *UpdateCountTask) Execute(ctx context.Context) error {
	// Implementation will be added when count service is implemented
	return fmt.Errorf("count service not yet implemented")
}

// SendEmailTask sends an email
type SendEmailTask struct {
	To      string
	Subject string
	Body    string
	// EmailService would be injected here
}

func (t *SendEmailTask) Execute(ctx context.Context) error {
	// Implementation will be added when email service is implemented
	return fmt.Errorf("email service not yet implemented")
}
