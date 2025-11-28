package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// PostListOptions defines options for listing posts
type PostListOptions struct {
	AuthorID   *int64
	CircleID   *int64
	Status     string
	SortBy     string // "created_at", "hotness_score", "view_count"
	Order      string // "asc", "desc"
	Limit      int
	Offset     int
}

// PostRepository defines the interface for post data operations
type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	FindByID(ctx context.Context, id int64) (*models.Post, error)
	Update(ctx context.Context, post *models.Post) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, opts PostListOptions) ([]*models.Post, error)
	Count(ctx context.Context, opts PostListOptions) (int64, error)
	IncrementViewCount(ctx context.Context, id int64) error
	UpdateHotnessScore(ctx context.Context, id int64, score float64) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	CountByDate(ctx context.Context, date string) (int64, error)
}

// postRepository implements PostRepository interface
type postRepository struct {
	db *gorm.DB
}

// NewPostRepository creates a new post repository
func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

// Create creates a new post
func (r *postRepository) Create(ctx context.Context, post *models.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

// FindByID finds a post by ID
func (r *postRepository) FindByID(ctx context.Context, id int64) (*models.Post, error) {
	var post models.Post
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&post).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

// Update updates a post
func (r *postRepository) Update(ctx context.Context, post *models.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

// Delete soft deletes a post by ID
func (r *postRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&models.Post{}).
		Where("id = ?", id).
		Update("status", "deleted").Error
}

// List retrieves posts based on options
func (r *postRepository) List(ctx context.Context, opts PostListOptions) ([]*models.Post, error) {
	var posts []*models.Post
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
	
	err := query.Find(&posts).Error
	return posts, err
}

// Count counts posts based on options
func (r *postRepository) Count(ctx context.Context, opts PostListOptions) (int64, error) {
	var count int64
	query := r.buildListQuery(ctx, opts)
	err := query.Count(&count).Error
	return count, err
}

// buildListQuery builds the base query for listing posts
func (r *postRepository) buildListQuery(ctx context.Context, opts PostListOptions) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&models.Post{})
	
	if opts.AuthorID != nil {
		query = query.Where("author_id = ?", *opts.AuthorID)
	}
	if opts.CircleID != nil {
		query = query.Where("circle_id = ?", *opts.CircleID)
	}
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}
	
	return query
}

// IncrementViewCount increments the view count for a post
func (r *postRepository) IncrementViewCount(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&models.Post{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// UpdateHotnessScore updates the hotness score for a post
func (r *postRepository) UpdateHotnessScore(ctx context.Context, id int64, score float64) error {
	return r.db.WithContext(ctx).Model(&models.Post{}).
		Where("id = ?", id).
		Update("hotness_score", score).Error
}

// UpdateStatus updates the status of a post
func (r *postRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	return r.db.WithContext(ctx).Model(&models.Post{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// CountByDate counts posts created on a specific date
func (r *postRepository) CountByDate(ctx context.Context, date string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Post{}).
		Where("DATE(created_at) = ?", date).
		Count(&count).Error
	return count, err
}
