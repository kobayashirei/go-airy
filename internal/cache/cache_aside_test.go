package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheAsideService_GetOrLoad_CacheHit(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	cache := NewCacheService(client, 1*time.Hour)
	cacheAside := NewCacheAsideService(cache, 1*time.Hour)
	ctx := context.Background()

	// Pre-populate cache
	testData := map[string]string{"id": "123", "name": "test"}
	err := cache.Set(ctx, "test:key", testData, 1*time.Minute)
	require.NoError(t, err)

	// GetOrLoad should hit cache and not call loader
	loaderCalled := false
	loader := func(ctx context.Context) (interface{}, error) {
		loaderCalled = true
		return nil, errors.New("should not be called")
	}

	var result map[string]string
	err = cacheAside.GetOrLoad(ctx, "test:key", &result, loader, 1*time.Minute)
	require.NoError(t, err)
	assert.False(t, loaderCalled, "Loader should not be called on cache hit")
	assert.Equal(t, "123", result["id"])
	assert.Equal(t, "test", result["name"])
}

func TestCacheAsideService_GetOrLoad_CacheMiss(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	cache := NewCacheService(client, 1*time.Hour)
	cacheAside := NewCacheAsideService(cache, 1*time.Hour)
	ctx := context.Background()

	// Loader should be called on cache miss
	testData := map[string]string{"id": "456", "name": "loaded"}
	loaderCalled := false
	loader := func(ctx context.Context) (interface{}, error) {
		loaderCalled = true
		return testData, nil
	}

	var result map[string]string
	err := cacheAside.GetOrLoad(ctx, "test:key", &result, loader, 1*time.Minute)
	require.NoError(t, err)
	assert.True(t, loaderCalled, "Loader should be called on cache miss")
	assert.Equal(t, "456", result["id"])
	assert.Equal(t, "loaded", result["name"])

	// Verify data was cached
	var cached map[string]string
	err = cache.Get(ctx, "test:key", &cached)
	require.NoError(t, err)
	assert.Equal(t, testData, cached)
}

func TestCacheAsideService_GetOrLoad_LoaderError(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	cache := NewCacheService(client, 1*time.Hour)
	cacheAside := NewCacheAsideService(cache, 1*time.Hour)
	ctx := context.Background()

	// Loader returns error
	expectedErr := errors.New("database error")
	loader := func(ctx context.Context) (interface{}, error) {
		return nil, expectedErr
	}

	var result map[string]string
	err := cacheAside.GetOrLoad(ctx, "test:key", &result, loader, 1*time.Minute)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load data")
}

func TestCacheAsideService_Invalidate(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	cache := NewCacheService(client, 1*time.Hour)
	cacheAside := NewCacheAsideService(cache, 1*time.Hour)
	ctx := context.Background()

	// Set a value
	err := cache.Set(ctx, "test:key", "value", 1*time.Minute)
	require.NoError(t, err)

	// Invalidate it
	err = cacheAside.Invalidate(ctx, "test:key")
	require.NoError(t, err)

	// Verify it's gone
	var result string
	err = cache.Get(ctx, "test:key", &result)
	assert.ErrorIs(t, err, ErrCacheMiss)
}

func TestCacheAsideService_InvalidateMultiple(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	cache := NewCacheService(client, 1*time.Hour)
	cacheAside := NewCacheAsideService(cache, 1*time.Hour)
	ctx := context.Background()

	// Set multiple values
	keys := []string{"test:key1", "test:key2", "test:key3"}
	for _, key := range keys {
		err := cache.Set(ctx, key, "value", 1*time.Minute)
		require.NoError(t, err)
	}

	// Invalidate all
	err := cacheAside.InvalidateMultiple(ctx, keys)
	require.NoError(t, err)

	// Verify all are gone
	for _, key := range keys {
		var result string
		err := cache.Get(ctx, key, &result)
		assert.ErrorIs(t, err, ErrCacheMiss)
	}
}

func TestCacheAsideService_Warmup(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	cache := NewCacheService(client, 1*time.Hour)
	cacheAside := NewCacheAsideService(cache, 1*time.Hour)
	ctx := context.Background()

	// Warmup should load data into cache
	testData := map[string]string{"id": "789", "name": "warmed"}
	loader := func(ctx context.Context) (interface{}, error) {
		return testData, nil
	}

	err := cacheAside.Warmup(ctx, "test:key", loader, 1*time.Minute)
	require.NoError(t, err)

	// Verify data is in cache
	var result map[string]string
	err = cache.Get(ctx, "test:key", &result)
	require.NoError(t, err)
	assert.Equal(t, testData, result)
}

func TestCacheAsideService_Warmup_SkipExisting(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	cache := NewCacheService(client, 1*time.Hour)
	cacheAside := NewCacheAsideService(cache, 1*time.Hour)
	ctx := context.Background()

	// Pre-populate cache
	existingData := map[string]string{"id": "existing", "name": "data"}
	err := cache.Set(ctx, "test:key", existingData, 1*time.Minute)
	require.NoError(t, err)

	// Warmup should skip existing key
	loaderCalled := false
	loader := func(ctx context.Context) (interface{}, error) {
		loaderCalled = true
		return map[string]string{"id": "new", "name": "data"}, nil
	}

	err = cacheAside.Warmup(ctx, "test:key", loader, 1*time.Minute)
	require.NoError(t, err)
	assert.False(t, loaderCalled, "Loader should not be called for existing key")

	// Verify original data is unchanged
	var result map[string]string
	err = cache.Get(ctx, "test:key", &result)
	require.NoError(t, err)
	assert.Equal(t, existingData, result)
}

func TestCacheAsideService_WarmupBatch(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	cache := NewCacheService(client, 1*time.Hour)
	cacheAside := NewCacheAsideService(cache, 1*time.Hour)
	ctx := context.Background()

	// Create batch warmup items
	items := []WarmupItem{
		{
			Key: "test:key1",
			Loader: func(ctx context.Context) (interface{}, error) {
				return map[string]string{"id": "1"}, nil
			},
			Expiration: 1 * time.Minute,
		},
		{
			Key: "test:key2",
			Loader: func(ctx context.Context) (interface{}, error) {
				return map[string]string{"id": "2"}, nil
			},
			Expiration: 1 * time.Minute,
		},
		{
			Key: "test:key3",
			Loader: func(ctx context.Context) (interface{}, error) {
				return map[string]string{"id": "3"}, nil
			},
			Expiration: 1 * time.Minute,
		},
	}

	err := cacheAside.WarmupBatch(ctx, items)
	require.NoError(t, err)

	// Verify all items are in cache
	for i, item := range items {
		var result map[string]string
		err := cache.Get(ctx, item.Key, &result)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"id": string(rune('1' + i))}, result)
	}
}

func TestCacheAsideService_RefreshCache(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	cache := NewCacheService(client, 1*time.Hour)
	cacheAside := NewCacheAsideService(cache, 1*time.Hour)
	ctx := context.Background()

	// Set initial value
	err := cache.Set(ctx, "test:key", map[string]string{"id": "old"}, 1*time.Minute)
	require.NoError(t, err)

	// Refresh with new value
	newData := map[string]string{"id": "new"}
	loader := func(ctx context.Context) (interface{}, error) {
		return newData, nil
	}

	err = cacheAside.RefreshCache(ctx, "test:key", loader, 1*time.Minute)
	require.NoError(t, err)

	// Verify cache has new value
	var result map[string]string
	err = cache.Get(ctx, "test:key", &result)
	require.NoError(t, err)
	assert.Equal(t, newData, result)
}
