package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/kobayashirei/airy/internal/models"
)

// EntityCacheService provides cache operations for specific entities
type EntityCacheService struct {
	cacheAside *CacheAsideService
	keyGen     *KeyGenerator
}

// NewEntityCacheService creates a new entity cache service
func NewEntityCacheService(cache Service, defaultExpiration time.Duration) *EntityCacheService {
	return &EntityCacheService{
		cacheAside: NewCacheAsideService(cache, defaultExpiration),
		keyGen:     NewKeyGenerator(),
	}
}

// GetUser retrieves a user from cache or loads from database
func (s *EntityCacheService) GetUser(
	ctx context.Context,
	userID int64,
	loader func(ctx context.Context) (*models.User, error),
) (*models.User, error) {
	key := s.keyGen.UserKey(userID)
	var user models.User

	err := s.cacheAside.GetOrLoad(ctx, key, &user, func(ctx context.Context) (interface{}, error) {
		return loader(ctx)
	}, 1*time.Hour)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// InvalidateUser removes user from cache
func (s *EntityCacheService) InvalidateUser(ctx context.Context, userID int64) error {
	keys := []string{
		s.keyGen.UserKey(userID),
		s.keyGen.UserProfileKey(userID),
		s.keyGen.UserStatsKey(userID),
	}
	return s.cacheAside.InvalidateMultiple(ctx, keys)
}

// GetPost retrieves a post from cache or loads from database
func (s *EntityCacheService) GetPost(
	ctx context.Context,
	postID int64,
	loader func(ctx context.Context) (*models.Post, error),
) (*models.Post, error) {
	key := s.keyGen.PostKey(postID)
	var post models.Post

	err := s.cacheAside.GetOrLoad(ctx, key, &post, func(ctx context.Context) (interface{}, error) {
		return loader(ctx)
	}, 30*time.Minute)

	if err != nil {
		return nil, err
	}

	return &post, nil
}

// InvalidatePost removes post from cache
func (s *EntityCacheService) InvalidatePost(ctx context.Context, postID int64) error {
	keys := []string{
		s.keyGen.PostKey(postID),
		s.keyGen.PostCountKey(postID),
	}
	return s.cacheAside.InvalidateMultiple(ctx, keys)
}

// GetCircle retrieves a circle from cache or loads from database
func (s *EntityCacheService) GetCircle(
	ctx context.Context,
	circleID int64,
	loader func(ctx context.Context) (*models.Circle, error),
) (*models.Circle, error) {
	key := s.keyGen.CircleKey(circleID)
	var circle models.Circle

	err := s.cacheAside.GetOrLoad(ctx, key, &circle, func(ctx context.Context) (interface{}, error) {
		return loader(ctx)
	}, 1*time.Hour)

	if err != nil {
		return nil, err
	}

	return &circle, nil
}

// InvalidateCircle removes circle from cache
func (s *EntityCacheService) InvalidateCircle(ctx context.Context, circleID int64) error {
	keys := []string{
		s.keyGen.CircleKey(circleID),
		s.keyGen.CircleFeedKey(circleID),
	}
	return s.cacheAside.InvalidateMultiple(ctx, keys)
}

// SetSession stores a session token in cache
func (s *EntityCacheService) SetSession(
	ctx context.Context,
	token string,
	userID int64,
	expiration time.Duration,
) error {
	key := s.keyGen.SessionKey(token)
	return s.cacheAside.GetCache().Set(ctx, key, userID, expiration)
}

// GetSession retrieves a user ID from session token
func (s *EntityCacheService) GetSession(ctx context.Context, token string) (int64, error) {
	key := s.keyGen.SessionKey(token)
	var userID int64
	err := s.cacheAside.GetCache().Get(ctx, key, &userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// DeleteSession removes a session token from cache
func (s *EntityCacheService) DeleteSession(ctx context.Context, token string) error {
	key := s.keyGen.SessionKey(token)
	return s.cacheAside.Invalidate(ctx, key)
}

// SetVerificationCode stores a verification code in cache
func (s *EntityCacheService) SetVerificationCode(
	ctx context.Context,
	identifier string,
	code string,
	expiration time.Duration,
) error {
	key := s.keyGen.VerificationCodeKey(identifier)
	return s.cacheAside.GetCache().Set(ctx, key, code, expiration)
}

// GetVerificationCode retrieves a verification code from cache
func (s *EntityCacheService) GetVerificationCode(ctx context.Context, identifier string) (string, error) {
	key := s.keyGen.VerificationCodeKey(identifier)
	var code string
	err := s.cacheAside.GetCache().Get(ctx, key, &code)
	if err != nil {
		return "", err
	}
	return code, nil
}

// DeleteVerificationCode removes a verification code from cache
func (s *EntityCacheService) DeleteVerificationCode(ctx context.Context, identifier string) error {
	key := s.keyGen.VerificationCodeKey(identifier)
	return s.cacheAside.Invalidate(ctx, key)
}

// SetActivationToken stores an activation token in cache
func (s *EntityCacheService) SetActivationToken(
	ctx context.Context,
	token string,
	userID int64,
	expiration time.Duration,
) error {
	key := s.keyGen.ActivationTokenKey(token)
	return s.cacheAside.GetCache().Set(ctx, key, userID, expiration)
}

// GetActivationToken retrieves a user ID from activation token
func (s *EntityCacheService) GetActivationToken(ctx context.Context, token string) (int64, error) {
	key := s.keyGen.ActivationTokenKey(token)
	var userID int64
	err := s.cacheAside.GetCache().Get(ctx, key, &userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// DeleteActivationToken removes an activation token from cache
func (s *EntityCacheService) DeleteActivationToken(ctx context.Context, token string) error {
	key := s.keyGen.ActivationTokenKey(token)
	return s.cacheAside.Invalidate(ctx, key)
}

// WarmupHotData pre-loads frequently accessed data into cache
func (s *EntityCacheService) WarmupHotData(
	ctx context.Context,
	hotUserIDs []int64,
	hotPostIDs []int64,
	hotCircleIDs []int64,
	userLoader func(ctx context.Context, id int64) (*models.User, error),
	postLoader func(ctx context.Context, id int64) (*models.Post, error),
	circleLoader func(ctx context.Context, id int64) (*models.Circle, error),
) error {
	var items []WarmupItem

	// Add hot users
	for _, userID := range hotUserIDs {
		id := userID // Capture loop variable
		items = append(items, WarmupItem{
			Key: s.keyGen.UserKey(id),
			Loader: func(ctx context.Context) (interface{}, error) {
				return userLoader(ctx, id)
			},
			Expiration: 1 * time.Hour,
		})
	}

	// Add hot posts
	for _, postID := range hotPostIDs {
		id := postID // Capture loop variable
		items = append(items, WarmupItem{
			Key: s.keyGen.PostKey(id),
			Loader: func(ctx context.Context) (interface{}, error) {
				return postLoader(ctx, id)
			},
			Expiration: 30 * time.Minute,
		})
	}

	// Add hot circles
	for _, circleID := range hotCircleIDs {
		id := circleID // Capture loop variable
		items = append(items, WarmupItem{
			Key: s.keyGen.CircleKey(id),
			Loader: func(ctx context.Context) (interface{}, error) {
				return circleLoader(ctx, id)
			},
			Expiration: 1 * time.Hour,
		})
	}

	return s.cacheAside.WarmupBatch(ctx, items)
}

// InvalidateUserRelated invalidates all cache entries related to a user
func (s *EntityCacheService) InvalidateUserRelated(ctx context.Context, userID int64) error {
	keys := []string{
		s.keyGen.UserKey(userID),
		s.keyGen.UserProfileKey(userID),
		s.keyGen.UserStatsKey(userID),
		s.keyGen.UserFeedKey(userID),
		s.keyGen.NotificationListKey(userID),
		s.keyGen.ConversationListKey(userID),
	}
	return s.cacheAside.InvalidateMultiple(ctx, keys)
}

// InvalidatePostRelated invalidates all cache entries related to a post
func (s *EntityCacheService) InvalidatePostRelated(ctx context.Context, postID int64, authorID int64, circleID *int64) error {
	keys := []string{
		s.keyGen.PostKey(postID),
		s.keyGen.PostCountKey(postID),
		s.keyGen.UserFeedKey(authorID),
	}

	if circleID != nil {
		keys = append(keys, s.keyGen.CircleFeedKey(*circleID))
	}

	return s.cacheAside.InvalidateMultiple(ctx, keys)
}

// GetEntityCount retrieves entity count from cache or loads from database
func (s *EntityCacheService) GetEntityCount(
	ctx context.Context,
	entityType string,
	entityID int64,
	loader func(ctx context.Context) (*models.EntityCount, error),
) (*models.EntityCount, error) {
	var key string
	if entityType == "post" {
		key = s.keyGen.PostCountKey(entityID)
	} else if entityType == "comment" {
		key = s.keyGen.CommentCountKey(entityID)
	} else {
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}

	var count models.EntityCount
	err := s.cacheAside.GetOrLoad(ctx, key, &count, func(ctx context.Context) (interface{}, error) {
		return loader(ctx)
	}, 10*time.Minute)

	if err != nil {
		return nil, err
	}

	return &count, nil
}

// InvalidateEntityCount removes entity count from cache
func (s *EntityCacheService) InvalidateEntityCount(ctx context.Context, entityType string, entityID int64) error {
	var key string
	if entityType == "post" {
		key = s.keyGen.PostCountKey(entityID)
	} else if entityType == "comment" {
		key = s.keyGen.CommentCountKey(entityID)
	} else {
		return fmt.Errorf("unsupported entity type: %s", entityType)
	}

	return s.cacheAside.Invalidate(ctx, key)
}
