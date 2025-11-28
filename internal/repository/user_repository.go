package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// UserListOptions defines options for listing users
type UserListOptions struct {
	Status  string
	Keyword string
	SortBy  string // "created_at", "username"
	Order   string // "asc", "desc"
	Limit   int
	Offset  int
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id int64) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByPhone(ctx context.Context, phone string) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int64) error
	UpdateLoginInfo(ctx context.Context, id int64, ip string) error
	List(ctx context.Context, opts UserListOptions) ([]*models.User, error)
	Count(ctx context.Context, opts UserListOptions) (int64, error)
	CountByLastLogin(ctx context.Context, date string) (int64, error)
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByID finds a user by ID
func (r *userRepository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByPhone finds a user by phone number
func (r *userRepository) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsername finds a user by username
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete soft deletes a user by ID
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// UpdateLoginInfo updates user's last login time and IP
func (r *userRepository) UpdateLoginInfo(ctx context.Context, id int64, ip string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_ip": ip,
			"last_login_at": gorm.Expr("NOW()"),
		}).Error
}

// List retrieves users based on options
func (r *userRepository) List(ctx context.Context, opts UserListOptions) ([]*models.User, error) {
	var users []*models.User
	query := r.buildListQuery(ctx, opts)

	// Apply sorting
	sortBy := "created_at"
	if opts.SortBy != "" {
		sortBy = opts.SortBy
	}
	order := "DESC"
	if opts.Order != "" {
		order = opts.Order
	}
	query = query.Order(sortBy + " " + order)

	// Apply pagination
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	err := query.Find(&users).Error
	return users, err
}

// Count counts users based on options
func (r *userRepository) Count(ctx context.Context, opts UserListOptions) (int64, error) {
	var count int64
	query := r.buildListQuery(ctx, opts)
	err := query.Count(&count).Error
	return count, err
}

// buildListQuery builds the base query for listing users
func (r *userRepository) buildListQuery(ctx context.Context, opts UserListOptions) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&models.User{})

	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}
	if opts.Keyword != "" {
		keyword := "%" + opts.Keyword + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR phone LIKE ?", keyword, keyword, keyword)
	}

	return query
}

// CountByLastLogin counts users who logged in on a specific date
func (r *userRepository) CountByLastLogin(ctx context.Context, date string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("DATE(last_login_at) = ?", date).
		Count(&count).Error
	return count, err
}
