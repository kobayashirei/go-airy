package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// UserProfileRepository defines the interface for user profile data operations
type UserProfileRepository interface {
	Create(ctx context.Context, profile *models.UserProfile) error
	FindByUserID(ctx context.Context, userID int64) (*models.UserProfile, error)
	Update(ctx context.Context, profile *models.UserProfile) error
	IncrementFollowerCount(ctx context.Context, userID int64, delta int) error
	IncrementFollowingCount(ctx context.Context, userID int64, delta int) error
	UpdatePoints(ctx context.Context, userID int64, points int) error
}

// userProfileRepository implements UserProfileRepository interface
type userProfileRepository struct {
	db *gorm.DB
}

// NewUserProfileRepository creates a new user profile repository
func NewUserProfileRepository(db *gorm.DB) UserProfileRepository {
	return &userProfileRepository{db: db}
}

// Create creates a new user profile
func (r *userProfileRepository) Create(ctx context.Context, profile *models.UserProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

// FindByUserID finds a user profile by user ID
func (r *userProfileRepository) FindByUserID(ctx context.Context, userID int64) (*models.UserProfile, error) {
	var profile models.UserProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

// Update updates a user profile
func (r *userProfileRepository) Update(ctx context.Context, profile *models.UserProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

// IncrementFollowerCount increments or decrements the follower count
func (r *userProfileRepository) IncrementFollowerCount(ctx context.Context, userID int64, delta int) error {
	return r.db.WithContext(ctx).Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("follower_count", gorm.Expr("follower_count + ?", delta)).Error
}

// IncrementFollowingCount increments or decrements the following count
func (r *userProfileRepository) IncrementFollowingCount(ctx context.Context, userID int64, delta int) error {
	return r.db.WithContext(ctx).Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("following_count", gorm.Expr("following_count + ?", delta)).Error
}

// UpdatePoints updates user points
func (r *userProfileRepository) UpdatePoints(ctx context.Context, userID int64, points int) error {
	return r.db.WithContext(ctx).Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		Update("points", points).Error
}
