package cache

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	appLogger "github.com/kobayashirei/airy/internal/logger"
)

// CacheAsideService implements the Cache-Aside pattern
type CacheAsideService struct {
	cache     Service
	keyGen    *KeyGenerator
	expiration time.Duration
}

// NewCacheAsideService creates a new cache-aside service
func NewCacheAsideService(cache Service, expiration time.Duration) *CacheAsideService {
	return &CacheAsideService{
		cache:      cache,
		keyGen:     NewKeyGenerator(),
		expiration: expiration,
	}
}

// GetOrLoad implements the Cache-Aside pattern:
// 1. Try to get from cache
// 2. If cache miss, load from database using loader function
// 3. Store the loaded value in cache
// 4. Return the value
func (s *CacheAsideService) GetOrLoad(
	ctx context.Context,
	key string,
	dest interface{},
	loader func(ctx context.Context) (interface{}, error),
	expiration time.Duration,
) error {
	// Try to get from cache first
	err := s.cache.Get(ctx, key, dest)
	if err == nil {
		// Cache hit
		return nil
	}

	if err != ErrCacheMiss {
		// Log cache error but continue to load from database
		appLogger.Warn("Cache get error, falling back to database",
			zap.String("key", key),
			zap.Error(err),
		)
	}

	// Cache miss - load from database
	value, err := loader(ctx)
	if err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}

	// Store in cache for next time
	if expiration == 0 {
		expiration = s.expiration
	}

	if err := s.cache.Set(ctx, key, value, expiration); err != nil {
		// Log cache error but don't fail the request
		appLogger.Warn("Failed to set cache after loading from database",
			zap.String("key", key),
			zap.Error(err),
		)
	}

	// Copy the loaded value to dest
	// This is a workaround since we can't directly assign interface{} to dest
	// In practice, the loader should return the correct type
	if err := s.cache.Get(ctx, key, dest); err != nil {
		// If we just set it and can't get it, something is wrong
		// But we have the value from the loader, so we can still return it
		appLogger.Warn("Failed to get value from cache after setting it",
			zap.String("key", key),
			zap.Error(err),
		)
		// Try to use the value directly (this requires type assertion in practice)
		// For now, we'll return an error and let the caller handle it
		return fmt.Errorf("cache inconsistency after load: %w", err)
	}

	return nil
}

// Invalidate removes a key from cache (used when data is updated)
func (s *CacheAsideService) Invalidate(ctx context.Context, key string) error {
	if err := s.cache.Delete(ctx, key); err != nil {
		appLogger.Warn("Failed to invalidate cache",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	appLogger.Debug("Cache invalidated", zap.String("key", key))
	return nil
}

// InvalidateMultiple removes multiple keys from cache
func (s *CacheAsideService) InvalidateMultiple(ctx context.Context, keys []string) error {
	var lastErr error
	for _, key := range keys {
		if err := s.Invalidate(ctx, key); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Warmup pre-loads data into cache
func (s *CacheAsideService) Warmup(
	ctx context.Context,
	key string,
	loader func(ctx context.Context) (interface{}, error),
	expiration time.Duration,
) error {
	// Check if already in cache
	exists, err := s.cache.Exists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check cache existence: %w", err)
	}

	if exists {
		appLogger.Debug("Cache key already exists, skipping warmup", zap.String("key", key))
		return nil
	}

	// Load from database
	value, err := loader(ctx)
	if err != nil {
		return fmt.Errorf("failed to load data for warmup: %w", err)
	}

	// Store in cache
	if expiration == 0 {
		expiration = s.expiration
	}

	if err := s.cache.Set(ctx, key, value, expiration); err != nil {
		return fmt.Errorf("failed to set cache during warmup: %w", err)
	}

	appLogger.Debug("Cache warmed up", zap.String("key", key))
	return nil
}

// WarmupBatch pre-loads multiple items into cache
func (s *CacheAsideService) WarmupBatch(
	ctx context.Context,
	items []WarmupItem,
) error {
	var lastErr error
	successCount := 0

	for _, item := range items {
		if err := s.Warmup(ctx, item.Key, item.Loader, item.Expiration); err != nil {
			appLogger.Warn("Failed to warmup cache item",
				zap.String("key", item.Key),
				zap.Error(err),
			)
			lastErr = err
		} else {
			successCount++
		}
	}

	appLogger.Info("Cache warmup completed",
		zap.Int("success", successCount),
		zap.Int("total", len(items)),
	)

	return lastErr
}

// WarmupItem represents an item to be warmed up in cache
type WarmupItem struct {
	Key        string
	Loader     func(ctx context.Context) (interface{}, error)
	Expiration time.Duration
}

// RefreshCache updates the cache with fresh data from database
func (s *CacheAsideService) RefreshCache(
	ctx context.Context,
	key string,
	loader func(ctx context.Context) (interface{}, error),
	expiration time.Duration,
) error {
	// Load fresh data from database
	value, err := loader(ctx)
	if err != nil {
		return fmt.Errorf("failed to load data for refresh: %w", err)
	}

	// Update cache
	if expiration == 0 {
		expiration = s.expiration
	}

	if err := s.cache.Set(ctx, key, value, expiration); err != nil {
		return fmt.Errorf("failed to refresh cache: %w", err)
	}

	appLogger.Debug("Cache refreshed", zap.String("key", key))
	return nil
}

// GetKeyGenerator returns the key generator instance
func (s *CacheAsideService) GetKeyGenerator() *KeyGenerator {
	return s.keyGen
}

// GetCache returns the underlying cache service
func (s *CacheAsideService) GetCache() Service {
	return s.cache
}
