package repository

import (
	"context"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// AdminLogListOptions defines options for listing admin logs
type AdminLogListOptions struct {
	OperatorID *int64
	Action     string
	EntityType string
	StartDate  *string
	EndDate    *string
	SortBy     string // "created_at"
	Order      string // "asc", "desc"
	Limit      int
	Offset     int
}

// AdminLogRepository defines the interface for admin log data operations
type AdminLogRepository interface {
	Create(ctx context.Context, log *models.AdminLog) error
	List(ctx context.Context, opts AdminLogListOptions) ([]*models.AdminLog, error)
	Count(ctx context.Context, opts AdminLogListOptions) (int64, error)
}

// adminLogRepository implements AdminLogRepository interface
type adminLogRepository struct {
	db *gorm.DB
}

// NewAdminLogRepository creates a new admin log repository
func NewAdminLogRepository(db *gorm.DB) AdminLogRepository {
	return &adminLogRepository{db: db}
}

// Create creates a new admin log entry
func (r *adminLogRepository) Create(ctx context.Context, log *models.AdminLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// List retrieves admin logs based on options
func (r *adminLogRepository) List(ctx context.Context, opts AdminLogListOptions) ([]*models.AdminLog, error) {
	var logs []*models.AdminLog
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

	err := query.Find(&logs).Error
	return logs, err
}

// Count counts admin logs based on options
func (r *adminLogRepository) Count(ctx context.Context, opts AdminLogListOptions) (int64, error) {
	var count int64
	query := r.buildListQuery(ctx, opts)
	err := query.Count(&count).Error
	return count, err
}

// buildListQuery builds the base query for listing admin logs
func (r *adminLogRepository) buildListQuery(ctx context.Context, opts AdminLogListOptions) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&models.AdminLog{})

	if opts.OperatorID != nil {
		query = query.Where("operator_id = ?", *opts.OperatorID)
	}
	if opts.Action != "" {
		query = query.Where("action = ?", opts.Action)
	}
	if opts.EntityType != "" {
		query = query.Where("entity_type = ?", opts.EntityType)
	}
	if opts.StartDate != nil {
		query = query.Where("created_at >= ?", *opts.StartDate)
	}
	if opts.EndDate != nil {
		query = query.Where("created_at <= ?", *opts.EndDate)
	}

	return query
}
