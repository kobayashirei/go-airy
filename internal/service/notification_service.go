package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/repository"
)

var (
	// ErrNotificationNotFound is returned when notification is not found
	ErrNotificationNotFound = errors.New("notification not found")
	// ErrUnauthorizedNotification is returned when user tries to access notification they don't own
	ErrUnauthorizedNotification = errors.New("unauthorized to access this notification")
)

// NotificationService defines the interface for notification business logic
type NotificationService interface {
	CreateNotification(ctx context.Context, req CreateNotificationRequest) (*models.Notification, error)
	GetNotifications(ctx context.Context, userID int64, page, pageSize int) (*NotificationListResponse, error)
	MarkAsRead(ctx context.Context, userID, notificationID int64) error
	MarkAllAsRead(ctx context.Context, userID int64) error
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)
}

// CreateNotificationRequest represents a notification creation request
type CreateNotificationRequest struct {
	ReceiverID    int64  `json:"receiver_id"`
	TriggerUserID *int64 `json:"trigger_user_id,omitempty"`
	Type          string `json:"type"`          // comment, vote, mention, system
	EntityType    string `json:"entity_type"`   // post, comment
	EntityID      *int64 `json:"entity_id,omitempty"`
	Content       string `json:"content"`
}

// NotificationListResponse represents a paginated list of notifications
type NotificationListResponse struct {
	Notifications []*models.Notification `json:"notifications"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
	UnreadCount   int64                  `json:"unread_count"`
}

// notificationService implements NotificationService interface
type notificationService struct {
	notificationRepo repository.NotificationRepository
	userRepo         repository.UserRepository
	postRepo         repository.PostRepository
	commentRepo      repository.CommentRepository
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	notificationRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		postRepo:         postRepo,
		commentRepo:      commentRepo,
	}
}

// CreateNotification creates a new notification with pre-rendered content
func (s *notificationService) CreateNotification(ctx context.Context, req CreateNotificationRequest) (*models.Notification, error) {
	// Validate notification type
	validTypes := map[string]bool{
		"comment": true,
		"vote":    true,
		"mention": true,
		"system":  true,
	}
	if !validTypes[req.Type] {
		return nil, fmt.Errorf("invalid notification type: %s", req.Type)
	}

	// Validate entity type if provided
	if req.EntityType != "" {
		validEntityTypes := map[string]bool{
			"post":    true,
			"comment": true,
		}
		if !validEntityTypes[req.EntityType] {
			return nil, fmt.Errorf("invalid entity type: %s", req.EntityType)
		}
	}

	// Pre-render notification content if not provided
	content := req.Content
	if content == "" {
		var err error
		content, err = s.renderNotificationContent(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to render notification content: %w", err)
		}
	}

	// Create notification
	notification := &models.Notification{
		ReceiverID:    req.ReceiverID,
		TriggerUserID: req.TriggerUserID,
		Type:          req.Type,
		EntityType:    req.EntityType,
		EntityID:      req.EntityID,
		Content:       content,
		IsRead:        false,
		CreatedAt:     time.Now(),
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notification, nil
}

// GetNotifications retrieves notifications for a user with unread notifications first
func (s *notificationService) GetNotifications(ctx context.Context, userID int64, page, pageSize int) (*NotificationListResponse, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get notifications (repository already orders by is_read ASC, created_at DESC)
	notifications, err := s.notificationRepo.FindByReceiverID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve notifications: %w", err)
	}

	// Get unread count
	unreadCount, err := s.notificationRepo.CountUnreadByReceiverID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	// For total count, we would need another query or cache this value
	// For now, we'll use a simple approach
	total := int64(len(notifications))
	if len(notifications) == pageSize {
		// There might be more, but we don't know exactly how many
		// In production, you'd want to do a separate count query
		total = int64(offset + pageSize + 1)
	}

	return &NotificationListResponse{
		Notifications: notifications,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
		UnreadCount:   unreadCount,
	}, nil
}

// MarkAsRead marks a notification as read
func (s *notificationService) MarkAsRead(ctx context.Context, userID, notificationID int64) error {
	// Verify notification belongs to user
	notification, err := s.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to find notification: %w", err)
	}
	if notification == nil {
		return ErrNotificationNotFound
	}
	if notification.ReceiverID != userID {
		return ErrUnauthorizedNotification
	}

	// Mark as read
	if err := s.notificationRepo.MarkAsRead(ctx, notificationID); err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (s *notificationService) MarkAllAsRead(ctx context.Context, userID int64) error {
	if err := s.notificationRepo.MarkAllAsRead(ctx, userID); err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *notificationService) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	count, err := s.notificationRepo.CountUnreadByReceiverID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}
	return count, nil
}

// renderNotificationContent generates pre-rendered notification content
func (s *notificationService) renderNotificationContent(ctx context.Context, req CreateNotificationRequest) (string, error) {
	var triggerUsername string
	if req.TriggerUserID != nil {
		triggerUser, err := s.userRepo.FindByID(ctx, *req.TriggerUserID)
		if err != nil {
			return "", err
		}
		if triggerUser != nil {
			triggerUsername = triggerUser.Username
		}
	}

	switch req.Type {
	case "comment":
		if req.EntityType == "post" && req.EntityID != nil {
			post, err := s.postRepo.FindByID(ctx, *req.EntityID)
			if err != nil {
				return "", err
			}
			if post != nil {
				return fmt.Sprintf("%s commented on your post: %s", triggerUsername, post.Title), nil
			}
		} else if req.EntityType == "comment" && req.EntityID != nil {
			return fmt.Sprintf("%s replied to your comment", triggerUsername), nil
		}
		return fmt.Sprintf("%s commented on your content", triggerUsername), nil

	case "vote":
		if req.EntityType == "post" && req.EntityID != nil {
			post, err := s.postRepo.FindByID(ctx, *req.EntityID)
			if err != nil {
				return "", err
			}
			if post != nil {
				return fmt.Sprintf("%s upvoted your post: %s", triggerUsername, post.Title), nil
			}
		} else if req.EntityType == "comment" {
			return fmt.Sprintf("%s upvoted your comment", triggerUsername), nil
		}
		return fmt.Sprintf("%s upvoted your content", triggerUsername), nil

	case "mention":
		if req.EntityType == "post" && req.EntityID != nil {
			post, err := s.postRepo.FindByID(ctx, *req.EntityID)
			if err != nil {
				return "", err
			}
			if post != nil {
				return fmt.Sprintf("%s mentioned you in a post: %s", triggerUsername, post.Title), nil
			}
		} else if req.EntityType == "comment" {
			return fmt.Sprintf("%s mentioned you in a comment", triggerUsername), nil
		}
		return fmt.Sprintf("%s mentioned you", triggerUsername), nil

	case "system":
		// System notifications should have content provided
		return "System notification", nil

	default:
		return "New notification", nil
	}
}
