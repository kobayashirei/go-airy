package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	FindByID(ctx context.Context, id int64) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
	FindAll(ctx context.Context) ([]*models.Role, error)
	Update(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, id int64) error
}

// roleRepository implements RoleRepository interface
type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

// Create creates a new role
func (r *roleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// FindByID finds a role by ID
func (r *roleRepository) FindByID(ctx context.Context, id int64) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// FindByName finds a role by name
func (r *roleRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// FindAll retrieves all roles
func (r *roleRepository) FindAll(ctx context.Context) ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.WithContext(ctx).Find(&roles).Error
	return roles, err
}

// Update updates a role
func (r *roleRepository) Update(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete deletes a role by ID
func (r *roleRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Role{}, id).Error
}
