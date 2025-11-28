package repository

import (
	"context"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// UserRoleRepository defines the interface for user role data operations
type UserRoleRepository interface {
	Create(ctx context.Context, userRole *models.UserRole) error
	FindByUserID(ctx context.Context, userID int64) ([]*models.UserRole, error)
	FindByUserIDAndCircleID(ctx context.Context, userID int64, circleID *int64) ([]*models.UserRole, error)
	FindRolesByUserID(ctx context.Context, userID int64) ([]*models.Role, error)
	Delete(ctx context.Context, userID, roleID int64, circleID *int64) error
	DeleteByUserID(ctx context.Context, userID int64) error
}

// userRoleRepository implements UserRoleRepository interface
type userRoleRepository struct {
	db *gorm.DB
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(db *gorm.DB) UserRoleRepository {
	return &userRoleRepository{db: db}
}

// Create creates a new user role association
func (r *userRoleRepository) Create(ctx context.Context, userRole *models.UserRole) error {
	return r.db.WithContext(ctx).Create(userRole).Error
}

// FindByUserID finds all role associations for a user
func (r *userRoleRepository) FindByUserID(ctx context.Context, userID int64) ([]*models.UserRole, error) {
	var userRoles []*models.UserRole
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&userRoles).Error
	return userRoles, err
}

// FindByUserIDAndCircleID finds role associations for a user in a specific circle
func (r *userRoleRepository) FindByUserIDAndCircleID(ctx context.Context, userID int64, circleID *int64) ([]*models.UserRole, error) {
	var userRoles []*models.UserRole
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	
	if circleID == nil {
		query = query.Where("circle_id IS NULL")
	} else {
		query = query.Where("circle_id = ?", *circleID)
	}
	
	err := query.Find(&userRoles).Error
	return userRoles, err
}

// FindRolesByUserID finds all roles for a user (including circle-specific roles)
func (r *userRoleRepository) FindRolesByUserID(ctx context.Context, userID int64) ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

// Delete removes a user role association
func (r *userRoleRepository) Delete(ctx context.Context, userID, roleID int64, circleID *int64) error {
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID)
	
	if circleID == nil {
		query = query.Where("circle_id IS NULL")
	} else {
		query = query.Where("circle_id = ?", *circleID)
	}
	
	return query.Delete(&models.UserRole{}).Error
}

// DeleteByUserID removes all role associations for a user
func (r *userRoleRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserRole{}).Error
}
