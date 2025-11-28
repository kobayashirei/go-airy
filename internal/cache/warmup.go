package cache

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	appLogger "github.com/kobayashirei/airy/internal/logger"
)

// WarmupConfig holds configuration for cache warming
type WarmupConfig struct {
	// Enabled indicates whether cache warming is enabled
	Enabled bool
	// HotPostsLimit is the number of hot posts to warm up
	HotPostsLimit int
	// HotUsersLimit is the number of active users to warm up
	HotUsersLimit int
	// HotCirclesLimit is the number of popular circles to warm up
	HotCirclesLimit int
	// RefreshInterval is the interval between cache refresh cycles
	RefreshInterval time.Duration
	// Concurrency is the number of concurrent warmup operations
	Concurrency int
}

// DefaultWarmupConfig returns default warmup configuration
func DefaultWarmupConfig() *WarmupConfig {
	return &WarmupConfig{
		Enabled:         true,
		HotPostsLimit:   100,
		HotUsersLimit:   50,
		HotCirclesLimit: 20,
		RefreshInterval: 30 * time.Minute,
		Concurrency:     10,
	}
}

// WarmupService handles cache warming operations
type WarmupService struct {
	cache      Service
	config     *WarmupConfig
	keyGen     *KeyGenerator
	stopCh     chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex
	isRunning  bool
}

// HotDataProvider defines the interface for loading hot data
type HotDataProvider interface {
	// GetHotPostIDs returns IDs of hot/popular posts
	GetHotPostIDs(ctx context.Context, limit int) ([]int64, error)
	// GetActiveUserIDs returns IDs of recently active users
	GetActiveUserIDs(ctx context.Context, limit int) ([]int64, error)
	// GetPopularCircleIDs returns IDs of popular circles
	GetPopularCircleIDs(ctx context.Context, limit int) ([]int64, error)
	// LoadPost loads a post by ID
	LoadPost(ctx context.Context, id int64) (interface{}, error)
	// LoadUser loads a user by ID
	LoadUser(ctx context.Context, id int64) (interface{}, error)
	// LoadCircle loads a circle by ID
	LoadCircle(ctx context.Context, id int64) (interface{}, error)
}

// NewWarmupService creates a new warmup service
func NewWarmupService(cache Service, config *WarmupConfig) *WarmupService {
	if config == nil {
		config = DefaultWarmupConfig()
	}
	return &WarmupService{
		cache:  cache,
		config: config,
		keyGen: NewKeyGenerator(),
		stopCh: make(chan struct{}),
	}
}

// WarmupOnStartup performs initial cache warming at application startup
func (s *WarmupService) WarmupOnStartup(ctx context.Context, provider HotDataProvider) error {
	appLogger.Info("Starting cache warmup on startup",
		zap.Int("hot_posts_limit", s.config.HotPostsLimit),
		zap.Int("hot_users_limit", s.config.HotUsersLimit),
		zap.Int("hot_circles_limit", s.config.HotCirclesLimit),
	)

	startTime := time.Now()

	// Warm up hot posts
	postCount, postErr := s.warmupPosts(ctx, provider)
	if postErr != nil {
		appLogger.Warn("Failed to warm up some posts", zap.Error(postErr))
	}

	// Warm up active users
	userCount, userErr := s.warmupUsers(ctx, provider)
	if userErr != nil {
		appLogger.Warn("Failed to warm up some users", zap.Error(userErr))
	}

	// Warm up popular circles
	circleCount, circleErr := s.warmupCircles(ctx, provider)
	if circleErr != nil {
		appLogger.Warn("Failed to warm up some circles", zap.Error(circleErr))
	}

	duration := time.Since(startTime)
	appLogger.Info("Cache warmup completed",
		zap.Int("posts_warmed", postCount),
		zap.Int("users_warmed", userCount),
		zap.Int("circles_warmed", circleCount),
		zap.Duration("duration", duration),
	)

	return nil
}

// warmupPosts warms up hot posts
func (s *WarmupService) warmupPosts(ctx context.Context, provider HotDataProvider) (int, error) {
	postIDs, err := provider.GetHotPostIDs(ctx, s.config.HotPostsLimit)
	if err != nil {
		return 0, err
	}

	return s.warmupEntities(ctx, postIDs, func(ctx context.Context, id int64) (string, interface{}, error) {
		post, err := provider.LoadPost(ctx, id)
		if err != nil {
			return "", nil, err
		}
		return s.keyGen.PostKey(id), post, nil
	}, 30*time.Minute)
}

// warmupUsers warms up active users
func (s *WarmupService) warmupUsers(ctx context.Context, provider HotDataProvider) (int, error) {
	userIDs, err := provider.GetActiveUserIDs(ctx, s.config.HotUsersLimit)
	if err != nil {
		return 0, err
	}

	return s.warmupEntities(ctx, userIDs, func(ctx context.Context, id int64) (string, interface{}, error) {
		user, err := provider.LoadUser(ctx, id)
		if err != nil {
			return "", nil, err
		}
		return s.keyGen.UserKey(id), user, nil
	}, 1*time.Hour)
}

// warmupCircles warms up popular circles
func (s *WarmupService) warmupCircles(ctx context.Context, provider HotDataProvider) (int, error) {
	circleIDs, err := provider.GetPopularCircleIDs(ctx, s.config.HotCirclesLimit)
	if err != nil {
		return 0, err
	}

	return s.warmupEntities(ctx, circleIDs, func(ctx context.Context, id int64) (string, interface{}, error) {
		circle, err := provider.LoadCircle(ctx, id)
		if err != nil {
			return "", nil, err
		}
		return s.keyGen.CircleKey(id), circle, nil
	}, 1*time.Hour)
}

// warmupEntities warms up a list of entities concurrently
func (s *WarmupService) warmupEntities(
	ctx context.Context,
	ids []int64,
	loader func(ctx context.Context, id int64) (string, interface{}, error),
	expiration time.Duration,
) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	var (
		successCount int
		mu           sync.Mutex
		lastErr      error
	)

	// Create a semaphore for concurrency control
	sem := make(chan struct{}, s.config.Concurrency)
	var wg sync.WaitGroup

	for _, id := range ids {
		select {
		case <-ctx.Done():
			return successCount, ctx.Err()
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(entityID int64) {
			defer wg.Done()
			defer func() { <-sem }()

			key, value, err := loader(ctx, entityID)
			if err != nil {
				mu.Lock()
				lastErr = err
				mu.Unlock()
				appLogger.Debug("Failed to load entity for warmup",
					zap.Int64("id", entityID),
					zap.Error(err),
				)
				return
			}

			if value == nil {
				return
			}

			if err := s.cache.Set(ctx, key, value, expiration); err != nil {
				mu.Lock()
				lastErr = err
				mu.Unlock()
				appLogger.Debug("Failed to cache entity during warmup",
					zap.String("key", key),
					zap.Error(err),
				)
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
		}(id)
	}

	wg.Wait()
	return successCount, lastErr
}

// StartPeriodicRefresh starts a background goroutine that periodically refreshes the cache
func (s *WarmupService) StartPeriodicRefresh(provider HotDataProvider) {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ticker := time.NewTicker(s.config.RefreshInterval)
		defer ticker.Stop()

		appLogger.Info("Started periodic cache refresh",
			zap.Duration("interval", s.config.RefreshInterval),
		)

		for {
			select {
			case <-s.stopCh:
				appLogger.Info("Stopping periodic cache refresh")
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				if err := s.WarmupOnStartup(ctx, provider); err != nil {
					appLogger.Warn("Periodic cache refresh failed", zap.Error(err))
				}
				cancel()
			}
		}
	}()
}

// Stop stops the periodic refresh goroutine
func (s *WarmupService) Stop() {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = false
	close(s.stopCh)
	s.mu.Unlock()

	s.wg.Wait()
	appLogger.Info("Cache warmup service stopped")
}

// RefreshEntity refreshes a single entity in the cache
func (s *WarmupService) RefreshEntity(
	ctx context.Context,
	key string,
	loader func(ctx context.Context) (interface{}, error),
	expiration time.Duration,
) error {
	value, err := loader(ctx)
	if err != nil {
		return err
	}

	if value == nil {
		// Entity doesn't exist, remove from cache
		return s.cache.Delete(ctx, key)
	}

	return s.cache.Set(ctx, key, value, expiration)
}

// RefreshPost refreshes a post in the cache
func (s *WarmupService) RefreshPost(ctx context.Context, postID int64, loader func(ctx context.Context) (interface{}, error)) error {
	key := s.keyGen.PostKey(postID)
	return s.RefreshEntity(ctx, key, loader, 30*time.Minute)
}

// RefreshUser refreshes a user in the cache
func (s *WarmupService) RefreshUser(ctx context.Context, userID int64, loader func(ctx context.Context) (interface{}, error)) error {
	key := s.keyGen.UserKey(userID)
	return s.RefreshEntity(ctx, key, loader, 1*time.Hour)
}

// RefreshCircle refreshes a circle in the cache
func (s *WarmupService) RefreshCircle(ctx context.Context, circleID int64, loader func(ctx context.Context) (interface{}, error)) error {
	key := s.keyGen.CircleKey(circleID)
	return s.RefreshEntity(ctx, key, loader, 1*time.Hour)
}

// IsRunning returns whether the periodic refresh is running
func (s *WarmupService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}
