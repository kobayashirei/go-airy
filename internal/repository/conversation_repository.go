package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// ConversationRepository defines the interface for conversation data operations
type ConversationRepository interface {
	Create(ctx context.Context, conversation *models.Conversation) error
	FindByID(ctx context.Context, id int64) (*models.Conversation, error)
	FindByUsers(ctx context.Context, user1ID, user2ID int64) (*models.Conversation, error)
	FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]*models.Conversation, error)
	Update(ctx context.Context, conversation *models.Conversation) error
	UpdateLastMessageAt(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
}

// conversationRepository implements ConversationRepository interface
type conversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

// Create creates a new conversation
func (r *conversationRepository) Create(ctx context.Context, conversation *models.Conversation) error {
	return r.db.WithContext(ctx).Create(conversation).Error
}

// FindByID finds a conversation by ID
func (r *conversationRepository) FindByID(ctx context.Context, id int64) (*models.Conversation, error) {
	var conversation models.Conversation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&conversation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &conversation, nil
}

// FindByUsers finds a conversation between two users
func (r *conversationRepository) FindByUsers(ctx context.Context, user1ID, user2ID int64) (*models.Conversation, error) {
	var conversation models.Conversation
	
	// Ensure consistent ordering: smaller ID as user1, larger as user2
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}
	
	err := r.db.WithContext(ctx).
		Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", 
			user1ID, user2ID, user2ID, user1ID).
		First(&conversation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &conversation, nil
}

// FindByUserID finds all conversations for a user, sorted by last message time
func (r *conversationRepository) FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]*models.Conversation, error) {
	var conversations []*models.Conversation
	query := r.db.WithContext(ctx).
		Where("user1_id = ? OR user2_id = ?", userID, userID).
		Order("last_message_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&conversations).Error
	return conversations, err
}

// Update updates a conversation
func (r *conversationRepository) Update(ctx context.Context, conversation *models.Conversation) error {
	return r.db.WithContext(ctx).Save(conversation).Error
}

// UpdateLastMessageAt updates the last message timestamp
func (r *conversationRepository) UpdateLastMessageAt(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&models.Conversation{}).
		Where("id = ?", id).
		Update("last_message_at", gorm.Expr("NOW()")).Error
}

// Delete deletes a conversation by ID
func (r *conversationRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Conversation{}, id).Error
}
