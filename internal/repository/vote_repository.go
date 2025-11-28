package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// VoteRepository defines the interface for vote data operations
type VoteRepository interface {
	Create(ctx context.Context, vote *models.Vote) error
	FindByID(ctx context.Context, id int64) (*models.Vote, error)
	FindByUserAndEntity(ctx context.Context, userID int64, entityType string, entityID int64) (*models.Vote, error)
	Update(ctx context.Context, vote *models.Vote) error
	Upsert(ctx context.Context, vote *models.Vote) error
	Delete(ctx context.Context, id int64) error
	DeleteByUserAndEntity(ctx context.Context, userID int64, entityType string, entityID int64) error
	CountByEntity(ctx context.Context, entityType string, entityID int64, voteType string) (int64, error)
}

// voteRepository implements VoteRepository interface
type voteRepository struct {
	db *gorm.DB
}

// NewVoteRepository creates a new vote repository
func NewVoteRepository(db *gorm.DB) VoteRepository {
	return &voteRepository{db: db}
}

// Create creates a new vote
func (r *voteRepository) Create(ctx context.Context, vote *models.Vote) error {
	return r.db.WithContext(ctx).Create(vote).Error
}

// FindByID finds a vote by ID
func (r *voteRepository) FindByID(ctx context.Context, id int64) (*models.Vote, error) {
	var vote models.Vote
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&vote).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &vote, nil
}

// FindByUserAndEntity finds a vote by user and entity
func (r *voteRepository) FindByUserAndEntity(ctx context.Context, userID int64, entityType string, entityID int64) (*models.Vote, error) {
	var vote models.Vote
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND entity_type = ? AND entity_id = ?", userID, entityType, entityID).
		First(&vote).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &vote, nil
}

// Update updates a vote
func (r *voteRepository) Update(ctx context.Context, vote *models.Vote) error {
	return r.db.WithContext(ctx).Save(vote).Error
}

// Upsert creates or updates a vote (for idempotency)
func (r *voteRepository) Upsert(ctx context.Context, vote *models.Vote) error {
	// Try to find existing vote
	existing, err := r.FindByUserAndEntity(ctx, vote.UserID, vote.EntityType, vote.EntityID)
	if err != nil {
		return err
	}
	
	if existing != nil {
		// Update existing vote
		vote.ID = existing.ID
		return r.Update(ctx, vote)
	}
	
	// Create new vote
	return r.Create(ctx, vote)
}

// Delete deletes a vote by ID
func (r *voteRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Vote{}, id).Error
}

// DeleteByUserAndEntity deletes a vote by user and entity
func (r *voteRepository) DeleteByUserAndEntity(ctx context.Context, userID int64, entityType string, entityID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND entity_type = ? AND entity_id = ?", userID, entityType, entityID).
		Delete(&models.Vote{}).Error
}

// CountByEntity counts votes for an entity by vote type
func (r *voteRepository) CountByEntity(ctx context.Context, entityType string, entityID int64, voteType string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Vote{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID)
	
	if voteType != "" {
		query = query.Where("vote_type = ?", voteType)
	}
	
	err := query.Count(&count).Error
	return count, err
}
