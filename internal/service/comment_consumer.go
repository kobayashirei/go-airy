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

// CommentEventConsumer handles comment-related events from the message queue
type CommentEventConsumer struct {
	commentRepo      repository.CommentRepository
	postRepo         repository.PostRepository
	userRepo         repository.UserRepository
	entityCountRepo  repository.EntityCountRepository
	notificationRepo repository.NotificationRepository
}

// NewCommentEventConsumer creates a new comment event consumer
func NewCommentEventConsumer(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
	userRepo repository.UserRepository,
	entityCountRepo repository.EntityCountRepository,
	notificationRepo repository.NotificationRepository,
) *CommentEventConsumer {
	return &CommentEventConsumer{
		commentRepo:      commentRepo,
		postRepo:         postRepo,
		userRepo:         userRepo,
		entityCountRepo:  entityCountRepo,
		notificationRepo: notificationRepo,
	}
}

// HandleCommentCreated handles the comment.created event
// Implements Requirement 5.5
func (c *CommentEventConsumer) HandleCommentCreated(ctx context.Context, message []byte) error {
	var event mq.CommentCreatedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal comment created event: %w", err)
	}

	// 1. Update post comment count
	if err := c.updatePostCommentCount(ctx, event.PostID); err != nil {
		fmt.Printf("failed to update post comment count: %v\n", err)
		// Don't return error, continue with other tasks
	}

	// 2. Generate notification for post author
	if err := c.notifyPostAuthor(ctx, &event); err != nil {
		fmt.Printf("failed to notify post author: %v\n", err)
	}

	// 3. Generate notification for parent comment author (if reply)
	if event.ParentID != nil {
		if err := c.notifyParentCommentAuthor(ctx, &event); err != nil {
			fmt.Printf("failed to notify parent comment author: %v\n", err)
		}
	}

	// 4. Parse @mentions and generate notifications
	if err := c.notifyMentionedUsers(ctx, &event); err != nil {
		fmt.Printf("failed to notify mentioned users: %v\n", err)
	}

	return nil
}

// HandleCommentDeleted handles the comment.deleted event
func (c *CommentEventConsumer) HandleCommentDeleted(ctx context.Context, message []byte) error {
	var event mq.CommentDeletedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal comment deleted event: %w", err)
	}

	// Update post comment count (decrement)
	if err := c.decrementPostCommentCount(ctx, event.PostID); err != nil {
		fmt.Printf("failed to decrement post comment count: %v\n", err)
	}

	return nil
}

// updatePostCommentCount increments the comment count for a post
func (c *CommentEventConsumer) updatePostCommentCount(ctx context.Context, postID int64) error {
	// First, ensure entity count record exists
	entityCount, err := c.entityCountRepo.FindByEntity(ctx, "post", postID)
	if err != nil {
		return fmt.Errorf("failed to find entity count: %w", err)
	}

	if entityCount == nil {
		// Create new entity count record
		entityCount = &models.EntityCount{
			EntityType:    "post",
			EntityID:      postID,
			CommentCount:  1,
			UpvoteCount:   0,
			DownvoteCount: 0,
			FavoriteCount: 0,
			UpdatedAt:     time.Now(),
		}
		return c.entityCountRepo.Create(ctx, entityCount)
	}

	// Increment comment count
	return c.entityCountRepo.IncrementCommentCount(ctx, "post", postID, 1)
}

// decrementPostCommentCount decrements the comment count for a post
func (c *CommentEventConsumer) decrementPostCommentCount(ctx context.Context, postID int64) error {
	return c.entityCountRepo.IncrementCommentCount(ctx, "post", postID, -1)
}

// notifyPostAuthor creates a notification for the post author
func (c *CommentEventConsumer) notifyPostAuthor(ctx context.Context, event *mq.CommentCreatedEvent) error {
	// Get post to find author
	post, err := c.postRepo.FindByID(ctx, event.PostID)
	if err != nil {
		return fmt.Errorf("failed to find post: %w", err)
	}
	if post == nil {
		return fmt.Errorf("post not found")
	}

	// Don't notify if commenter is the post author
	if post.AuthorID == event.AuthorID {
		return nil
	}

	// Get commenter info for notification content
	commenter, err := c.userRepo.FindByID(ctx, event.AuthorID)
	if err != nil {
		return fmt.Errorf("failed to find commenter: %w", err)
	}
	if commenter == nil {
		return fmt.Errorf("commenter not found")
	}

	// Create notification
	notification := &models.Notification{
		ReceiverID:    post.AuthorID,
		TriggerUserID: &event.AuthorID,
		Type:          "comment",
		EntityType:    "post",
		EntityID:      &event.PostID,
		Content:       fmt.Sprintf("%s commented on your post: %s", commenter.Username, truncateContent(event.Content, 50)),
		IsRead:        false,
		CreatedAt:     time.Now(),
	}

	return c.notificationRepo.Create(ctx, notification)
}

// notifyParentCommentAuthor creates a notification for the parent comment author
func (c *CommentEventConsumer) notifyParentCommentAuthor(ctx context.Context, event *mq.CommentCreatedEvent) error {
	if event.ParentID == nil {
		return nil
	}

	// Get parent comment to find author
	parentComment, err := c.commentRepo.FindByID(ctx, *event.ParentID)
	if err != nil {
		return fmt.Errorf("failed to find parent comment: %w", err)
	}
	if parentComment == nil {
		return fmt.Errorf("parent comment not found")
	}

	// Don't notify if replier is the parent comment author
	if parentComment.AuthorID == event.AuthorID {
		return nil
	}

	// Get replier info for notification content
	replier, err := c.userRepo.FindByID(ctx, event.AuthorID)
	if err != nil {
		return fmt.Errorf("failed to find replier: %w", err)
	}
	if replier == nil {
		return fmt.Errorf("replier not found")
	}

	// Create notification
	notification := &models.Notification{
		ReceiverID:    parentComment.AuthorID,
		TriggerUserID: &event.AuthorID,
		Type:          "comment",
		EntityType:    "comment",
		EntityID:      event.ParentID,
		Content:       fmt.Sprintf("%s replied to your comment: %s", replier.Username, truncateContent(event.Content, 50)),
		IsRead:        false,
		CreatedAt:     time.Now(),
	}

	return c.notificationRepo.Create(ctx, notification)
}

// notifyMentionedUsers creates notifications for users mentioned in the comment
func (c *CommentEventConsumer) notifyMentionedUsers(ctx context.Context, event *mq.CommentCreatedEvent) error {
	// Extract mentions from content
	mentions := ExtractMentions(event.Content)
	if len(mentions) == 0 {
		return nil
	}

	// Get commenter info for notification content
	commenter, err := c.userRepo.FindByID(ctx, event.AuthorID)
	if err != nil {
		return fmt.Errorf("failed to find commenter: %w", err)
	}
	if commenter == nil {
		return fmt.Errorf("commenter not found")
	}

	// Create notifications for each mentioned user
	for _, username := range mentions {
		// Find user by username
		mentionedUser, err := c.userRepo.FindByUsername(ctx, username)
		if err != nil {
			fmt.Printf("failed to find mentioned user %s: %v\n", username, err)
			continue
		}
		if mentionedUser == nil {
			// User doesn't exist, skip
			continue
		}

		// Don't notify if user mentions themselves
		if mentionedUser.ID == event.AuthorID {
			continue
		}

		// Create notification
		notification := &models.Notification{
			ReceiverID:    mentionedUser.ID,
			TriggerUserID: &event.AuthorID,
			Type:          "mention",
			EntityType:    "comment",
			EntityID:      &event.CommentID,
			Content:       fmt.Sprintf("%s mentioned you in a comment: %s", commenter.Username, truncateContent(event.Content, 50)),
			IsRead:        false,
			CreatedAt:     time.Now(),
		}

		if err := c.notificationRepo.Create(ctx, notification); err != nil {
			fmt.Printf("failed to create mention notification for user %s: %v\n", username, err)
		}
	}

	return nil
}

// truncateContent truncates content to a maximum length
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}
