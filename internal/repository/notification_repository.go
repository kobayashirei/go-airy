package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// NotificationRepository defines the interface for notification data operations
type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	FindByID(ctx context.Context, id int64) (*models.Notification, error)
	FindByReceiverID(ctx context.Context, receiverID int64, limit, offset int) ([]*models.Notification, error)
	FindUnreadByReceiverID(ctx context.Context, receiverID int64, limit, offset int) ([]*models.Notification, error)
	Update(ctx context.Context, notification *models.Notification) error
	MarkAsRead(ctx context.Context, id int64) error
	MarkAllAsRead(ctx context.Context, receiverID int64) error
	Delete(ctx context.Context, id int64) error
	CountUnreadByReceiverID(ctx context.Context, receiverID int64) (int64, error)
}

// notificationRepository implements NotificationRepository interface
type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// Create creates a new notification
func (r *notificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// FindByID finds a notification by ID
func (r *notificationRepository) FindByID(ctx context.Context, id int64) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&notification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &notification, nil
}

// FindByReceiverID finds all notifications for a receiver with pagination
func (r *notificationRepository) FindByReceiverID(ctx context.Context, receiverID int64, limit, offset int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	query := r.db.WithContext(ctx).
		Where("receiver_id = ?", receiverID).
		Order("is_read ASC, created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&notifications).Error
	return notifications, err
}

// FindUnreadByReceiverID finds unread notifications for a receiver
func (r *notificationRepository) FindUnreadByReceiverID(ctx context.Context, receiverID int64, limit, offset int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	query := r.db.WithContext(ctx).
		Where("receiver_id = ? AND is_read = ?", receiverID, false).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&notifications).Error
	return notifications, err
}

// Update updates a notification
func (r *notificationRepository) Update(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

// MarkAsRead marks a notification as read
func (r *notificationRepository) MarkAsRead(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Update("is_read", true).Error
}

// MarkAllAsRead marks all notifications for a receiver as read
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, receiverID int64) error {
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("receiver_id = ? AND is_read = ?", receiverID, false).
		Update("is_read", true).Error
}

// Delete deletes a notification by ID
func (r *notificationRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Notification{}, id).Error
}

// CountUnreadByReceiverID counts unread notifications for a receiver
func (r *notificationRepository) CountUnreadByReceiverID(ctx context.Context, receiverID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("receiver_id = ? AND is_read = ?", receiverID, false).
		Count(&count).Error
	return count, err
}
