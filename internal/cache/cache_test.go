package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRedis creates a test Redis client
func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use a separate DB for testing
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available for testing")
	}

	// Clean up test DB
	client.FlushDB(ctx)

	return client
}

func TestCacheService_SetAndGet(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	service := NewCacheService(client, 1*time.Hour)
	ctx := context.Background()

	type TestData struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	testData := TestData{ID: 1, Name: "test"}

	// Test Set
	err := service.Set(ctx, "test:key", testData, 1*time.Minute)
	require.NoError(t, err)

	// Test Get
	var result TestData
	err = service.Get(ctx, "test:key", &result)
	require.NoError(t, err)
	assert.Equal(t, testData.ID, result.ID)
	assert.Equal(t, testData.Name, result.Name)
}

func TestCacheService_GetCacheMiss(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	service := NewCacheService(client, 1*time.Hour)
	ctx := context.Background()

	var result string
	err := service.Get(ctx, "nonexistent:key", &result)
	assert.ErrorIs(t, err, ErrCacheMiss)
}

func TestCacheService_Delete(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	service := NewCacheService(client, 1*time.Hour)
	ctx := context.Background()

	// Set a value
	err := service.Set(ctx, "test:key", "value", 1*time.Minute)
	require.NoError(t, err)

	// Delete it
	err = service.Delete(ctx, "test:key")
	require.NoError(t, err)

	// Verify it's gone
	var result string
	err = service.Get(ctx, "test:key", &result)
	assert.ErrorIs(t, err, ErrCacheMiss)
}

func TestCacheService_Exists(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	service := NewCacheService(client, 1*time.Hour)
	ctx := context.Background()

	// Check non-existent key
	exists, err := service.Exists(ctx, "test:key")
	require.NoError(t, err)
	assert.False(t, exists)

	// Set a value
	err = service.Set(ctx, "test:key", "value", 1*time.Minute)
	require.NoError(t, err)

	// Check existing key
	exists, err = service.Exists(ctx, "test:key")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestCacheService_SetNX(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	service := NewCacheService(client, 1*time.Hour)
	ctx := context.Background()

	// First SetNX should succeed
	success, err := service.SetNX(ctx, "test:key", "value1", 1*time.Minute)
	require.NoError(t, err)
	assert.True(t, success)

	// Second SetNX should fail (key already exists)
	success, err = service.SetNX(ctx, "test:key", "value2", 1*time.Minute)
	require.NoError(t, err)
	assert.False(t, success)

	// Verify original value is unchanged
	var result string
	err = service.Get(ctx, "test:key", &result)
	require.NoError(t, err)
	assert.Equal(t, "value1", result)
}

func TestCacheService_Expire(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	service := NewCacheService(client, 1*time.Hour)
	ctx := context.Background()

	// Set a value with long expiration
	err := service.Set(ctx, "test:key", "value", 1*time.Hour)
	require.NoError(t, err)

	// Update expiration to 1 second
	err = service.Expire(ctx, "test:key", 1*time.Second)
	require.NoError(t, err)

	// Check TTL
	ttl, err := service.TTL(ctx, "test:key")
	require.NoError(t, err)
	assert.True(t, ttl > 0 && ttl <= 1*time.Second)
}

func TestCacheService_TTL(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	service := NewCacheService(client, 1*time.Hour)
	ctx := context.Background()

	// Set a value with 10 second expiration
	err := service.Set(ctx, "test:key", "value", 10*time.Second)
	require.NoError(t, err)

	// Check TTL
	ttl, err := service.TTL(ctx, "test:key")
	require.NoError(t, err)
	assert.True(t, ttl > 0 && ttl <= 10*time.Second)

	// Check TTL for non-existent key
	ttl, err = service.TTL(ctx, "nonexistent:key")
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-2), ttl) // Redis returns -2 for non-existent keys
}

func TestCacheService_DefaultExpiration(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	defaultExp := 5 * time.Second
	service := NewCacheService(client, defaultExp)
	ctx := context.Background()

	// Set without explicit expiration (should use default)
	err := service.Set(ctx, "test:key", "value", 0)
	require.NoError(t, err)

	// Check TTL is approximately the default
	ttl, err := service.TTL(ctx, "test:key")
	require.NoError(t, err)
	assert.True(t, ttl > 0 && ttl <= defaultExp)
}
