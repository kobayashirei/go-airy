package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// CircleRepository defines the interface for circle data operations
type CircleRepository interface {
	Create(ctx context.Context, circle *models.Circle) error
	FindByID(ctx context.Context, id int64) (*models.Circle, error)
	FindByName(ctx context.Context, name string) (*models.Circle, error)
	FindAll(ctx context.Context, limit, offset int) ([]*models.Circle, error)
	FindByCreatorID(ctx context.Context, creatorID int64) ([]*models.Circle, error)
	Update(ctx context.Context, circle *models.Circle) error
	Delete(ctx context.Context, id int64) error
	IncrementMemberCount(ctx context.Context, id int64, delta int) error
	IncrementPostCount(ctx context.Context, id int64, delta int) error
	Count(ctx context.Context) (int64, error)
}

// circleRepository implements CircleRepository interface
type circleRepository struct {
	db *gorm.DB
}

// NewCircleRepository creates a new circle repository
func NewCircleRepository(db *gorm.DB) CircleRepository {
	return &circleRepository{db: db}
}

// Create creates a new circle
func (r *circleRepository) Create(ctx context.Context, circle *models.Circle) error {
	return r.db.WithContext(ctx).Create(circle).Error
}

// FindByID finds a circle by ID
func (r *circleRepository) FindByID(ctx context.Context, id int64) (*models.Circle, error) {
	var circle models.Circle
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&circle).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &circle, nil
}

// FindByName finds a circle by name
func (r *circleRepository) FindByName(ctx context.Context, name string) (*models.Circle, error) {
	var circle models.Circle
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&circle).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &circle, nil
}

// FindAll retrieves all circles with pagination
func (r *circleRepository) FindAll(ctx context.Context, limit, offset int) ([]*models.Circle, error) {
	var circles []*models.Circle
	query := r.db.WithContext(ctx).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&circles).Error
	return circles, err
}

// FindByCreatorID finds all circles created by a user
func (r *circleRepository) FindByCreatorID(ctx context.Context, creatorID int64) ([]*models.Circle, error) {
	var circles []*models.Circle
	err := r.db.WithContext(ctx).
		Where("creator_id = ?", creatorID).
		Order("created_at DESC").
		Find(&circles).Error
	return circles, err
}

// Update updates a circle
func (r *circleRepository) Update(ctx context.Context, circle *models.Circle) error {
	return r.db.WithContext(ctx).Save(circle).Error
}

// Delete deletes a circle by ID
func (r *circleRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Circle{}, id).Error
}

// IncrementMemberCount increments or decrements the member count
func (r *circleRepository) IncrementMemberCount(ctx context.Context, id int64, delta int) error {
	return r.db.WithContext(ctx).Model(&models.Circle{}).
		Where("id = ?", id).
		UpdateColumn("member_count", gorm.Expr("member_count + ?", delta)).Error
}

// IncrementPostCount increments or decrements the post count
func (r *circleRepository) IncrementPostCount(ctx context.Context, id int64, delta int) error {
	return r.db.WithContext(ctx).Model(&models.Circle{}).
		Where("id = ?", id).
		UpdateColumn("post_count", gorm.Expr("post_count + ?", delta)).Error
}

// Count counts total circles
func (r *circleRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Circle{}).Count(&count).Error
	return count, err
}
