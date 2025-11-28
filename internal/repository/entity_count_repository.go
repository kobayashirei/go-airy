package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// EntityCountRepository defines the interface for entity count data operations
type EntityCountRepository interface {
	Create(ctx context.Context, count *models.EntityCount) error
	FindByEntity(ctx context.Context, entityType string, entityID int64) (*models.EntityCount, error)
	Update(ctx context.Context, count *models.EntityCount) error
	Upsert(ctx context.Context, count *models.EntityCount) error
	IncrementUpvoteCount(ctx context.Context, entityType string, entityID int64, delta int) error
	IncrementDownvoteCount(ctx context.Context, entityType string, entityID int64, delta int) error
	IncrementCommentCount(ctx context.Context, entityType string, entityID int64, delta int) error
	IncrementFavoriteCount(ctx context.Context, entityType string, entityID int64, delta int) error
}

// entityCountRepository implements EntityCountRepository interface
type entityCountRepository struct {
	db *gorm.DB
}

// NewEntityCountRepository creates a new entity count repository
func NewEntityCountRepository(db *gorm.DB) EntityCountRepository {
	return &entityCountRepository{db: db}
}

// Create creates a new entity count record
func (r *entityCountRepository) Create(ctx context.Context, count *models.EntityCount) error {
	return r.db.WithContext(ctx).Create(count).Error
}

// FindByEntity finds entity count by entity type and ID
func (r *entityCountRepository) FindByEntity(ctx context.Context, entityType string, entityID int64) (*models.EntityCount, error) {
	var count models.EntityCount
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		First(&count).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &count, nil
}

// Update updates an entity count record
func (r *entityCountRepository) Update(ctx context.Context, count *models.EntityCount) error {
	return r.db.WithContext(ctx).Save(count).Error
}

// Upsert creates or updates an entity count record
func (r *entityCountRepository) Upsert(ctx context.Context, count *models.EntityCount) error {
	// Try to find existing count
	existing, err := r.FindByEntity(ctx, count.EntityType, count.EntityID)
	if err != nil {
		return err
	}
	
	if existing != nil {
		// Update existing count
		return r.Update(ctx, count)
	}
	
	// Create new count
	return r.Create(ctx, count)
}

// IncrementUpvoteCount increments or decrements the upvote count
func (r *entityCountRepository) IncrementUpvoteCount(ctx context.Context, entityType string, entityID int64, delta int) error {
	return r.db.WithContext(ctx).Model(&models.EntityCount{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		UpdateColumn("upvote_count", gorm.Expr("upvote_count + ?", delta)).Error
}

// IncrementDownvoteCount increments or decrements the downvote count
func (r *entityCountRepository) IncrementDownvoteCount(ctx context.Context, entityType string, entityID int64, delta int) error {
	return r.db.WithContext(ctx).Model(&models.EntityCount{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		UpdateColumn("downvote_count", gorm.Expr("downvote_count + ?", delta)).Error
}

// IncrementCommentCount increments or decrements the comment count
func (r *entityCountRepository) IncrementCommentCount(ctx context.Context, entityType string, entityID int64, delta int) error {
	return r.db.WithContext(ctx).Model(&models.EntityCount{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", delta)).Error
}

// IncrementFavoriteCount increments or decrements the favorite count
func (r *entityCountRepository) IncrementFavoriteCount(ctx context.Context, entityType string, entityID int64, delta int) error {
	return r.db.WithContext(ctx).Model(&models.EntityCount{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", delta)).Error
}
