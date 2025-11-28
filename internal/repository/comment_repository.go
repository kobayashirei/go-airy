package repository

import (
	"context"
	"errors"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// CommentListOptions defines options for listing comments
type CommentListOptions struct {
	PostID   *int64
	AuthorID *int64
	Status   string
	Limit    int
	Offset   int
}

// CommentRepository defines the interface for comment data operations
type CommentRepository interface {
	Create(ctx context.Context, comment *models.Comment) error
	FindByID(ctx context.Context, id int64) (*models.Comment, error)
	FindByPostID(ctx context.Context, postID int64) ([]*models.Comment, error)
	FindByParentID(ctx context.Context, parentID int64) ([]*models.Comment, error)
	FindRootComments(ctx context.Context, postID int64, limit, offset int) ([]*models.Comment, error)
	Update(ctx context.Context, comment *models.Comment) error
	Delete(ctx context.Context, id int64) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	CountByPostID(ctx context.Context, postID int64) (int64, error)
	Count(ctx context.Context, opts CommentListOptions) (int64, error)
}

// commentRepository implements CommentRepository interface
type commentRepository struct {
	db *gorm.DB
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

// Create creates a new comment
func (r *commentRepository) Create(ctx context.Context, comment *models.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

// FindByID finds a comment by ID
func (r *commentRepository) FindByID(ctx context.Context, id int64) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&comment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &comment, nil
}

// FindByPostID finds all comments for a post
func (r *commentRepository) FindByPostID(ctx context.Context, postID int64) ([]*models.Comment, error) {
	var comments []*models.Comment
	err := r.db.WithContext(ctx).
		Where("post_id = ? AND status = ?", postID, "published").
		Order("path ASC").
		Find(&comments).Error
	return comments, err
}

// FindByParentID finds all direct replies to a comment
func (r *commentRepository) FindByParentID(ctx context.Context, parentID int64) ([]*models.Comment, error) {
	var comments []*models.Comment
	err := r.db.WithContext(ctx).
		Where("parent_id = ? AND status = ?", parentID, "published").
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

// FindRootComments finds root-level comments for a post
func (r *commentRepository) FindRootComments(ctx context.Context, postID int64, limit, offset int) ([]*models.Comment, error) {
	var comments []*models.Comment
	query := r.db.WithContext(ctx).
		Where("post_id = ? AND parent_id IS NULL AND status = ?", postID, "published").
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&comments).Error
	return comments, err
}

// Update updates a comment
func (r *commentRepository) Update(ctx context.Context, comment *models.Comment) error {
	return r.db.WithContext(ctx).Save(comment).Error
}

// Delete soft deletes a comment by ID
func (r *commentRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&models.Comment{}).
		Where("id = ?", id).
		Update("status", "deleted").Error
}

// UpdateStatus updates the status of a comment
func (r *commentRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	return r.db.WithContext(ctx).Model(&models.Comment{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// CountByPostID counts comments for a post
func (r *commentRepository) CountByPostID(ctx context.Context, postID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Comment{}).
		Where("post_id = ? AND status = ?", postID, "published").
		Count(&count).Error
	return count, err
}

// Count counts comments based on options
func (r *commentRepository) Count(ctx context.Context, opts CommentListOptions) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Comment{})

	if opts.PostID != nil {
		query = query.Where("post_id = ?", *opts.PostID)
	}
	if opts.AuthorID != nil {
		query = query.Where("author_id = ?", *opts.AuthorID)
	}
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}

	err := query.Count(&count).Error
	return count, err
}
