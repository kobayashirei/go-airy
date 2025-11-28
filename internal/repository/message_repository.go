package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// MessageRepository defines the interface for message data operations
type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	FindByID(ctx context.Context, id int64) (*models.Message, error)
	FindByConversationID(ctx context.Context, conversationID int64, limit, offset int) ([]*models.Message, error)
	FindUnreadByConversationAndReceiver(ctx context.Context, conversationID, receiverID int64) ([]*models.Message, error)
	Update(ctx context.Context, message *models.Message) error
	MarkAsRead(ctx context.Context, id int64) error
	MarkConversationAsRead(ctx context.Context, conversationID, receiverID int64) error
	Delete(ctx context.Context, id int64) error
	CountUnreadByReceiver(ctx context.Context, receiverID int64) (int64, error)
}

// messageRepository implements MessageRepository interface
type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

// Create creates a new message
func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// FindByID finds a message by ID
func (r *messageRepository) FindByID(ctx context.Context, id int64) (*models.Message, error) {
	var message models.Message
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&message).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

// FindByConversationID finds all messages in a conversation with pagination
func (r *messageRepository) FindByConversationID(ctx context.Context, conversationID int64, limit, offset int) ([]*models.Message, error) {
	var messages []*models.Message
	query := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&messages).Error
	return messages, err
}

// FindUnreadByConversationAndReceiver finds unread messages in a conversation for a receiver
func (r *messageRepository) FindUnreadByConversationAndReceiver(ctx context.Context, conversationID, receiverID int64) ([]*models.Message, error) {
	var messages []*models.Message
	err := r.db.WithContext(ctx).
		Where("conversation_id = ? AND sender_id != ? AND is_read = ?", conversationID, receiverID, false).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

// Update updates a message
func (r *messageRepository) Update(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

// MarkAsRead marks a message as read
func (r *messageRepository) MarkAsRead(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&models.Message{}).
		Where("id = ?", id).
		Update("is_read", true).Error
}

// MarkConversationAsRead marks all messages in a conversation as read for a receiver
func (r *messageRepository) MarkConversationAsRead(ctx context.Context, conversationID, receiverID int64) error {
	return r.db.WithContext(ctx).Model(&models.Message{}).
		Where("conversation_id = ? AND sender_id != ? AND is_read = ?", conversationID, receiverID, false).
		Update("is_read", true).Error
}

// Delete deletes a message by ID
func (r *messageRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Message{}, id).Error
}

// CountUnreadByReceiver counts unread messages for a receiver across all conversations
func (r *messageRepository) CountUnreadByReceiver(ctx context.Context, receiverID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Message{}).
		Where("sender_id != ? AND is_read = ?", receiverID, false).
		Joins("JOIN conversations ON conversations.id = messages.conversation_id").
		Where("conversations.user1_id = ? OR conversations.user2_id = ?", receiverID, receiverID).
		Count(&count).Error
	return count, err
}
