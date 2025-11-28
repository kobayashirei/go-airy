package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// PermissionRepository defines the interface for permission data operations
type PermissionRepository interface {
	Create(ctx context.Context, permission *models.Permission) error
	FindByID(ctx context.Context, id int64) (*models.Permission, error)
	FindByName(ctx context.Context, name string) (*models.Permission, error)
	FindAll(ctx context.Context) ([]*models.Permission, error)
	FindByRoleID(ctx context.Context, roleID int64) ([]*models.Permission, error)
	Update(ctx context.Context, permission *models.Permission) error
	Delete(ctx context.Context, id int64) error
}

// permissionRepository implements PermissionRepository interface
type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

// Create creates a new permission
func (r *permissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

// FindByID finds a permission by ID
func (r *permissionRepository) FindByID(ctx context.Context, id int64) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

// FindByName finds a permission by name
func (r *permissionRepository) FindByName(ctx context.Context, name string) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

// FindAll retrieves all permissions
func (r *permissionRepository) FindAll(ctx context.Context) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.WithContext(ctx).Find(&permissions).Error
	return permissions, err
}

// FindByRoleID finds all permissions for a given role
func (r *permissionRepository) FindByRoleID(ctx context.Context, roleID int64) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

// Update updates a permission
func (r *permissionRepository) Update(ctx context.Context, permission *models.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

// Delete deletes a permission by ID
func (r *permissionRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Permission{}, id).Error
}
