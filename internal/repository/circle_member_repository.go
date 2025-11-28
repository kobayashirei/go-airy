package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// CircleMemberRepository defines the interface for circle member data operations
type CircleMemberRepository interface {
	Create(ctx context.Context, member *models.CircleMember) error
	FindByID(ctx context.Context, id int64) (*models.CircleMember, error)
	FindByCircleAndUser(ctx context.Context, circleID, userID int64) (*models.CircleMember, error)
	FindByCircleID(ctx context.Context, circleID int64, role string) ([]*models.CircleMember, error)
	FindByUserID(ctx context.Context, userID int64) ([]*models.CircleMember, error)
	Update(ctx context.Context, member *models.CircleMember) error
	UpdateRole(ctx context.Context, id int64, role string) error
	Delete(ctx context.Context, id int64) error
	DeleteByCircleAndUser(ctx context.Context, circleID, userID int64) error
	CountByCircleID(ctx context.Context, circleID int64, role string) (int64, error)
	IsMember(ctx context.Context, circleID, userID int64) (bool, error)
}

// circleMemberRepository implements CircleMemberRepository interface
type circleMemberRepository struct {
	db *gorm.DB
}

// NewCircleMemberRepository creates a new circle member repository
func NewCircleMemberRepository(db *gorm.DB) CircleMemberRepository {
	return &circleMemberRepository{db: db}
}

// Create creates a new circle member record
func (r *circleMemberRepository) Create(ctx context.Context, member *models.CircleMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// FindByID finds a circle member by ID
func (r *circleMemberRepository) FindByID(ctx context.Context, id int64) (*models.CircleMember, error) {
	var member models.CircleMember
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

// FindByCircleAndUser finds a circle member by circle ID and user ID
func (r *circleMemberRepository) FindByCircleAndUser(ctx context.Context, circleID, userID int64) (*models.CircleMember, error) {
	var member models.CircleMember
	err := r.db.WithContext(ctx).
		Where("circle_id = ? AND user_id = ?", circleID, userID).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

// FindByCircleID finds all members of a circle, optionally filtered by role
func (r *circleMemberRepository) FindByCircleID(ctx context.Context, circleID int64, role string) ([]*models.CircleMember, error) {
	var members []*models.CircleMember
	query := r.db.WithContext(ctx).Where("circle_id = ?", circleID)
	
	if role != "" {
		query = query.Where("role = ?", role)
	}
	
	err := query.Order("joined_at DESC").Find(&members).Error
	return members, err
}

// FindByUserID finds all circles a user is a member of
func (r *circleMemberRepository) FindByUserID(ctx context.Context, userID int64) ([]*models.CircleMember, error) {
	var members []*models.CircleMember
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("joined_at DESC").
		Find(&members).Error
	return members, err
}

// Update updates a circle member record
func (r *circleMemberRepository) Update(ctx context.Context, member *models.CircleMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

// UpdateRole updates the role of a circle member
func (r *circleMemberRepository) UpdateRole(ctx context.Context, id int64, role string) error {
	return r.db.WithContext(ctx).Model(&models.CircleMember{}).
		Where("id = ?", id).
		Update("role", role).Error
}

// Delete deletes a circle member by ID
func (r *circleMemberRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.CircleMember{}, id).Error
}

// DeleteByCircleAndUser deletes a circle member by circle ID and user ID
func (r *circleMemberRepository) DeleteByCircleAndUser(ctx context.Context, circleID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("circle_id = ? AND user_id = ?", circleID, userID).
		Delete(&models.CircleMember{}).Error
}

// CountByCircleID counts members in a circle, optionally filtered by role
func (r *circleMemberRepository) CountByCircleID(ctx context.Context, circleID int64, role string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.CircleMember{}).
		Where("circle_id = ?", circleID)
	
	if role != "" {
		query = query.Where("role = ?", role)
	}
	
	err := query.Count(&count).Error
	return count, err
}

// IsMember checks if a user is a member of a circle
func (r *circleMemberRepository) IsMember(ctx context.Context, circleID, userID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.CircleMember{}).
		Where("circle_id = ? AND user_id = ? AND role != ?", circleID, userID, "pending").
		Count(&count).Error
	return count > 0, err
}
