package repository

import (
	"context"

	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

// BatchRepository provides batch query operations to avoid N+1 queries
type BatchRepository interface {
	// FindUsersByIDs retrieves multiple users by their IDs
	FindUsersByIDs(ctx context.Context, ids []int64) (map[int64]*models.User, error)
	// FindPostsByIDs retrieves multiple posts by their IDs
	FindPostsByIDs(ctx context.Context, ids []int64) (map[int64]*models.Post, error)
	// FindCirclesByIDs retrieves multiple circles by their IDs
	FindCirclesByIDs(ctx context.Context, ids []int64) (map[int64]*models.Circle, error)
	// FindCommentsByIDs retrieves multiple comments by their IDs
	FindCommentsByIDs(ctx context.Context, ids []int64) (map[int64]*models.Comment, error)
	// FindEntityCountsByIDs retrieves entity counts for multiple entities
	FindEntityCountsByIDs(ctx context.Context, entityType string, ids []int64) (map[int64]*models.EntityCount, error)
	// FindUserProfilesByIDs retrieves user profiles for multiple users
	FindUserProfilesByIDs(ctx context.Context, userIDs []int64) (map[int64]*models.UserProfile, error)
	// FindUserStatsByIDs retrieves user stats for multiple users
	FindUserStatsByIDs(ctx context.Context, userIDs []int64) (map[int64]*models.UserStats, error)
}

// batchRepository implements BatchRepository interface
type batchRepository struct {
	db *gorm.DB
}

// NewBatchRepository creates a new batch repository
func NewBatchRepository(db *gorm.DB) BatchRepository {
	return &batchRepository{db: db}
}

// FindUsersByIDs retrieves multiple users by their IDs
func (r *batchRepository) FindUsersByIDs(ctx context.Context, ids []int64) (map[int64]*models.User, error) {
	if len(ids) == 0 {
		return make(map[int64]*models.User), nil
	}

	var users []*models.User
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*models.User, len(users))
	for _, user := range users {
		result[user.ID] = user
	}
	return result, nil
}

// FindPostsByIDs retrieves multiple posts by their IDs
func (r *batchRepository) FindPostsByIDs(ctx context.Context, ids []int64) (map[int64]*models.Post, error) {
	if len(ids) == 0 {
		return make(map[int64]*models.Post), nil
	}

	var posts []*models.Post
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&posts).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*models.Post, len(posts))
	for _, post := range posts {
		result[post.ID] = post
	}
	return result, nil
}

// FindCirclesByIDs retrieves multiple circles by their IDs
func (r *batchRepository) FindCirclesByIDs(ctx context.Context, ids []int64) (map[int64]*models.Circle, error) {
	if len(ids) == 0 {
		return make(map[int64]*models.Circle), nil
	}

	var circles []*models.Circle
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&circles).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*models.Circle, len(circles))
	for _, circle := range circles {
		result[circle.ID] = circle
	}
	return result, nil
}

// FindCommentsByIDs retrieves multiple comments by their IDs
func (r *batchRepository) FindCommentsByIDs(ctx context.Context, ids []int64) (map[int64]*models.Comment, error) {
	if len(ids) == 0 {
		return make(map[int64]*models.Comment), nil
	}

	var comments []*models.Comment
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&comments).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*models.Comment, len(comments))
	for _, comment := range comments {
		result[comment.ID] = comment
	}
	return result, nil
}

// FindEntityCountsByIDs retrieves entity counts for multiple entities
func (r *batchRepository) FindEntityCountsByIDs(ctx context.Context, entityType string, ids []int64) (map[int64]*models.EntityCount, error) {
	if len(ids) == 0 {
		return make(map[int64]*models.EntityCount), nil
	}

	var counts []*models.EntityCount
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id IN ?", entityType, ids).
		Find(&counts).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*models.EntityCount, len(counts))
	for _, count := range counts {
		result[count.EntityID] = count
	}
	return result, nil
}

// FindUserProfilesByIDs retrieves user profiles for multiple users
func (r *batchRepository) FindUserProfilesByIDs(ctx context.Context, userIDs []int64) (map[int64]*models.UserProfile, error) {
	if len(userIDs) == 0 {
		return make(map[int64]*models.UserProfile), nil
	}

	var profiles []*models.UserProfile
	err := r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&profiles).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*models.UserProfile, len(profiles))
	for _, profile := range profiles {
		result[profile.UserID] = profile
	}
	return result, nil
}

// FindUserStatsByIDs retrieves user stats for multiple users
func (r *batchRepository) FindUserStatsByIDs(ctx context.Context, userIDs []int64) (map[int64]*models.UserStats, error) {
	if len(userIDs) == 0 {
		return make(map[int64]*models.UserStats), nil
	}

	var stats []*models.UserStats
	err := r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&stats).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*models.UserStats, len(stats))
	for _, stat := range stats {
		result[stat.UserID] = stat
	}
	return result, nil
}

// BatchLoader provides a generic batch loading utility
type BatchLoader[K comparable, V any] struct {
	loader func(ctx context.Context, keys []K) (map[K]V, error)
}

// NewBatchLoader creates a new batch loader
func NewBatchLoader[K comparable, V any](loader func(ctx context.Context, keys []K) (map[K]V, error)) *BatchLoader[K, V] {
	return &BatchLoader[K, V]{loader: loader}
}

// Load loads multiple items by their keys
func (l *BatchLoader[K, V]) Load(ctx context.Context, keys []K) (map[K]V, error) {
	return l.loader(ctx, keys)
}

// LoadOne loads a single item by key
func (l *BatchLoader[K, V]) LoadOne(ctx context.Context, key K) (V, bool, error) {
	result, err := l.loader(ctx, []K{key})
	if err != nil {
		var zero V
		return zero, false, err
	}
	val, ok := result[key]
	return val, ok, nil
}
