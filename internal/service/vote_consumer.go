package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/mq"
	"github.com/kobayashirei/airy/internal/repository"
)

// VoteConsumer handles vote events from the message queue
type VoteConsumer struct {
	entityCountRepo    repository.EntityCountRepository
	notificationRepo   repository.NotificationRepository
	postRepo           repository.PostRepository
	commentRepo        repository.CommentRepository
}

// NewVoteConsumer creates a new vote consumer
func NewVoteConsumer(
	entityCountRepo repository.EntityCountRepository,
	notificationRepo repository.NotificationRepository,
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
) *VoteConsumer {
	return &VoteConsumer{
		entityCountRepo:  entityCountRepo,
		notificationRepo: notificationRepo,
		postRepo:         postRepo,
		commentRepo:      commentRepo,
	}
}

// HandleVoteCreated handles vote created events
func (c *VoteConsumer) HandleVoteCreated(ctx context.Context, message []byte) error {
	var event mq.VoteCreatedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal vote created event: %w", err)
	}

	// Update entity count
	if err := c.updateEntityCountForVote(ctx, event.EntityType, event.EntityID, event.VoteType, 1); err != nil {
		return fmt.Errorf("failed to update entity count: %w", err)
	}

	// Generate notification (exclude self-votes)
	if err := c.generateVoteNotification(ctx, event.UserID, event.EntityType, event.EntityID, event.VoteType); err != nil {
		// Log error but don't fail the event processing
		fmt.Printf("failed to generate vote notification: %v\n", err)
	}

	return nil
}

// HandleVoteUpdated handles vote updated events
func (c *VoteConsumer) HandleVoteUpdated(ctx context.Context, message []byte) error {
	var event mq.VoteUpdatedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal vote updated event: %w", err)
	}

	// Decrement old vote type count
	if err := c.updateEntityCountForVote(ctx, event.EntityType, event.EntityID, event.OldVoteType, -1); err != nil {
		return fmt.Errorf("failed to decrement old vote count: %w", err)
	}

	// Increment new vote type count
	if err := c.updateEntityCountForVote(ctx, event.EntityType, event.EntityID, event.NewVoteType, 1); err != nil {
		return fmt.Errorf("failed to increment new vote count: %w", err)
	}

	// Note: We don't generate a new notification for vote updates
	// The original notification still stands

	return nil
}

// HandleVoteDeleted handles vote deleted events
func (c *VoteConsumer) HandleVoteDeleted(ctx context.Context, message []byte) error {
	var event mq.VoteDeletedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal vote deleted event: %w", err)
	}

	// Decrement vote count
	if err := c.updateEntityCountForVote(ctx, event.EntityType, event.EntityID, event.VoteType, -1); err != nil {
		return fmt.Errorf("failed to update entity count: %w", err)
	}

	// Note: We don't delete the notification when a vote is cancelled
	// The notification remains as historical record

	return nil
}

// updateEntityCountForVote updates the entity count for a vote
func (c *VoteConsumer) updateEntityCountForVote(ctx context.Context, entityType string, entityID int64, voteType string, delta int) error {
	// Ensure entity count record exists
	existingCount, err := c.entityCountRepo.FindByEntity(ctx, entityType, entityID)
	if err != nil {
		return fmt.Errorf("failed to find entity count: %w", err)
	}

	if existingCount == nil {
		// Create initial entity count record
		initialCount := &models.EntityCount{
			EntityType:    entityType,
			EntityID:      entityID,
			UpvoteCount:   0,
			DownvoteCount: 0,
			CommentCount:  0,
			FavoriteCount: 0,
			UpdatedAt:     time.Now(),
		}
		if err := c.entityCountRepo.Create(ctx, initialCount); err != nil {
			return fmt.Errorf("failed to create entity count: %w", err)
		}
	}

	// Update the appropriate count
	if voteType == "up" {
		if err := c.entityCountRepo.IncrementUpvoteCount(ctx, entityType, entityID, delta); err != nil {
			return fmt.Errorf("failed to increment upvote count: %w", err)
		}
	} else if voteType == "down" {
		if err := c.entityCountRepo.IncrementDownvoteCount(ctx, entityType, entityID, delta); err != nil {
			return fmt.Errorf("failed to increment downvote count: %w", err)
		}
	}

	return nil
}

// generateVoteNotification generates a notification for a vote
func (c *VoteConsumer) generateVoteNotification(ctx context.Context, voterID int64, entityType string, entityID int64, voteType string) error {
	// Get the content author ID
	var authorID int64
	var contentTitle string

	if entityType == "post" {
		post, err := c.postRepo.FindByID(ctx, entityID)
		if err != nil {
			return fmt.Errorf("failed to find post: %w", err)
		}
		if post == nil {
			return fmt.Errorf("post not found: %d", entityID)
		}
		authorID = post.AuthorID
		contentTitle = post.Title
	} else if entityType == "comment" {
		comment, err := c.commentRepo.FindByID(ctx, entityID)
		if err != nil {
			return fmt.Errorf("failed to find comment: %w", err)
		}
		if comment == nil {
			return fmt.Errorf("comment not found: %d", entityID)
		}
		authorID = comment.AuthorID
		// For comments, use a truncated version of the content
		if len(comment.Content) > 50 {
			contentTitle = comment.Content[:50] + "..."
		} else {
			contentTitle = comment.Content
		}
	}

	// Exclude self-votes (don't notify if user votes on their own content)
	if voterID == authorID {
		return nil
	}

	// Generate notification content
	var notificationContent string
	voteTypeText := "upvoted"
	if voteType == "down" {
		voteTypeText = "downvoted"
	}

	if entityType == "post" {
		notificationContent = fmt.Sprintf("Someone %s your post: %s", voteTypeText, contentTitle)
	} else {
		notificationContent = fmt.Sprintf("Someone %s your comment: %s", voteTypeText, contentTitle)
	}

	// Create notification
	notification := &models.Notification{
		ReceiverID:    authorID,
		TriggerUserID: &voterID,
		Type:          "vote",
		EntityType:    entityType,
		EntityID:      &entityID,
		Content:       notificationContent,
		IsRead:        false,
		CreatedAt:     time.Now(),
	}

	if err := c.notificationRepo.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}
