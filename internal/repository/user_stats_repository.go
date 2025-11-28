package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// UserStatsRepository defines the interface for user stats data operations
type UserStatsRepository interface {
	Create(ctx context.Context, stats *models.UserStats) error
	FindByUserID(ctx context.Context, userID int64) (*models.UserStats, error)
	Update(ctx context.Context, stats *models.UserStats) error
	IncrementPostCount(ctx context.Context, userID int64, delta int) error
	IncrementCommentCount(ctx context.Context, userID int64, delta int) error
	IncrementVoteReceivedCount(ctx context.Context, userID int64, delta int) error
}

// userStatsRepository implements UserStatsRepository interface
type userStatsRepository struct {
	db *gorm.DB
}

// NewUserStatsRepository creates a new user stats repository
func NewUserStatsRepository(db *gorm.DB) UserStatsRepository {
	return &userStatsRepository{db: db}
}

// Create creates a new user stats record
func (r *userStatsRepository) Create(ctx context.Context, stats *models.UserStats) error {
	return r.db.WithContext(ctx).Create(stats).Error
}

// FindByUserID finds user stats by user ID
func (r *userStatsRepository) FindByUserID(ctx context.Context, userID int64) (*models.UserStats, error) {
	var stats models.UserStats
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&stats).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &stats, nil
}

// Update updates user stats
func (r *userStatsRepository) Update(ctx context.Context, stats *models.UserStats) error {
	return r.db.WithContext(ctx).Save(stats).Error
}

// IncrementPostCount increments or decrements the post count
func (r *userStatsRepository) IncrementPostCount(ctx context.Context, userID int64, delta int) error {
	return r.db.WithContext(ctx).Model(&models.UserStats{}).
		Where("user_id = ?", userID).
		UpdateColumn("post_count", gorm.Expr("post_count + ?", delta)).Error
}

// IncrementCommentCount increments or decrements the comment count
func (r *userStatsRepository) IncrementCommentCount(ctx context.Context, userID int64, delta int) error {
	return r.db.WithContext(ctx).Model(&models.UserStats{}).
		Where("user_id = ?", userID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", delta)).Error
}

// IncrementVoteReceivedCount increments or decrements the vote received count
func (r *userStatsRepository) IncrementVoteReceivedCount(ctx context.Context, userID int64, delta int) error {
	return r.db.WithContext(ctx).Model(&models.UserStats{}).
		Where("user_id = ?", userID).
		UpdateColumn("vote_received_count", gorm.Expr("vote_received_count + ?", delta)).Error
}
