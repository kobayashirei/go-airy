package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kobayashirei/airy/internal/config"
	"github.com/kobayashirei/airy/internal/logger"
)

func init() {
	// Initialize logger for tests
	_ = logger.Init(&config.LogConfig{
		Level:  "error",
		Output: "stdout",
	})
}

// mockCacheService is a mock implementation of the Service interface for testing
type mockCacheService struct {
	data       map[string]interface{}
	getErr     error
	setErr     error
	deleteErr  error
	existsErr  error
}

func newMockCacheService() *mockCacheService {
	return &mockCacheService{
		data: make(map[string]interface{}),
	}
}

func (m *mockCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	if m.getErr != nil {
		return m.getErr
	}
	if _, ok := m.data[key]; !ok {
		return ErrCacheMiss
	}
	return nil
}

func (m *mockCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.data[key] = value
	return nil
}

func (m *mockCacheService) Delete(ctx context.Context, key string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.data, key)
	return nil
}

func (m *mockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	_, ok := m.data[key]
	return ok, nil
}

func (m *mockCacheService) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if _, ok := m.data[key]; ok {
		return false, nil
	}
	m.data[key] = value
	return true, nil
}

func (m *mockCacheService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}

func (m *mockCacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	return time.Hour, nil
}

// mockHotDataProvider is a mock implementation of HotDataProvider for testing
type mockHotDataProvider struct {
	hotPostIDs      []int64
	activeUserIDs   []int64
	popularCircleIDs []int64
	loadPostErr     error
	loadUserErr     error
	loadCircleErr   error
	loadPostCalls   int32
	loadUserCalls   int32
	loadCircleCalls int32
}

func newMockHotDataProvider() *mockHotDataProvider {
	return &mockHotDataProvider{
		hotPostIDs:       []int64{1, 2, 3, 4, 5},
		activeUserIDs:    []int64{10, 20, 30},
		popularCircleIDs: []int64{100, 200},
	}
}

func (m *mockHotDataProvider) GetHotPostIDs(ctx context.Context, limit int) ([]int64, error) {
	if limit > len(m.hotPostIDs) {
		return m.hotPostIDs, nil
	}
	return m.hotPostIDs[:limit], nil
}

func (m *mockHotDataProvider) GetActiveUserIDs(ctx context.Context, limit int) ([]int64, error) {
	if limit > len(m.activeUserIDs) {
		return m.activeUserIDs, nil
	}
	return m.activeUserIDs[:limit], nil
}

func (m *mockHotDataProvider) GetPopularCircleIDs(ctx context.Context, limit int) ([]int64, error) {
	if limit > len(m.popularCircleIDs) {
		return m.popularCircleIDs, nil
	}
	return m.popularCircleIDs[:limit], nil
}

func (m *mockHotDataProvider) LoadPost(ctx context.Context, id int64) (interface{}, error) {
	atomic.AddInt32(&m.loadPostCalls, 1)
	if m.loadPostErr != nil {
		return nil, m.loadPostErr
	}
	return map[string]interface{}{"id": id, "title": "Test Post"}, nil
}

func (m *mockHotDataProvider) LoadUser(ctx context.Context, id int64) (interface{}, error) {
	atomic.AddInt32(&m.loadUserCalls, 1)
	if m.loadUserErr != nil {
		return nil, m.loadUserErr
	}
	return map[string]interface{}{"id": id, "username": "testuser"}, nil
}

func (m *mockHotDataProvider) LoadCircle(ctx context.Context, id int64) (interface{}, error) {
	atomic.AddInt32(&m.loadCircleCalls, 1)
	if m.loadCircleErr != nil {
		return nil, m.loadCircleErr
	}
	return map[string]interface{}{"id": id, "name": "Test Circle"}, nil
}

func TestDefaultWarmupConfig(t *testing.T) {
	config := DefaultWarmupConfig()
	
	assert.True(t, config.Enabled)
	assert.Equal(t, 100, config.HotPostsLimit)
	assert.Equal(t, 50, config.HotUsersLimit)
	assert.Equal(t, 20, config.HotCirclesLimit)
	assert.Equal(t, 30*time.Minute, config.RefreshInterval)
	assert.Equal(t, 10, config.Concurrency)
}

func TestNewWarmupService(t *testing.T) {
	cache := newMockCacheService()
	
	t.Run("with nil config uses defaults", func(t *testing.T) {
		service := NewWarmupService(cache, nil)
		require.NotNil(t, service)
		assert.Equal(t, 100, service.config.HotPostsLimit)
	})
	
	t.Run("with custom config", func(t *testing.T) {
		config := &WarmupConfig{
			HotPostsLimit: 50,
			Concurrency:   5,
		}
		service := NewWarmupService(cache, config)
		require.NotNil(t, service)
		assert.Equal(t, 50, service.config.HotPostsLimit)
		assert.Equal(t, 5, service.config.Concurrency)
	})
}

func TestWarmupOnStartup(t *testing.T) {
	cache := newMockCacheService()
	config := &WarmupConfig{
		Enabled:         true,
		HotPostsLimit:   5,
		HotUsersLimit:   3,
		HotCirclesLimit: 2,
		Concurrency:     2,
	}
	service := NewWarmupService(cache, config)
	provider := newMockHotDataProvider()
	
	ctx := context.Background()
	err := service.WarmupOnStartup(ctx, provider)
	
	require.NoError(t, err)
	
	// Verify posts were cached
	assert.Equal(t, int32(5), atomic.LoadInt32(&provider.loadPostCalls))
	
	// Verify users were cached
	assert.Equal(t, int32(3), atomic.LoadInt32(&provider.loadUserCalls))
	
	// Verify circles were cached
	assert.Equal(t, int32(2), atomic.LoadInt32(&provider.loadCircleCalls))
	
	// Verify cache contains the data
	assert.Len(t, cache.data, 10) // 5 posts + 3 users + 2 circles
}

func TestWarmupWithLoadErrors(t *testing.T) {
	cache := newMockCacheService()
	config := &WarmupConfig{
		HotPostsLimit:   3,
		HotUsersLimit:   2,
		HotCirclesLimit: 1,
		Concurrency:     2,
	}
	service := NewWarmupService(cache, config)
	provider := newMockHotDataProvider()
	provider.loadPostErr = errors.New("load error")
	
	ctx := context.Background()
	err := service.WarmupOnStartup(ctx, provider)
	
	// Should not return error, just log warnings
	require.NoError(t, err)
	
	// Posts should not be cached due to error
	// But users and circles should be cached
	assert.Equal(t, int32(3), atomic.LoadInt32(&provider.loadPostCalls))
	assert.Equal(t, int32(2), atomic.LoadInt32(&provider.loadUserCalls))
}

func TestWarmupWithCacheSetError(t *testing.T) {
	cache := newMockCacheService()
	cache.setErr = errors.New("cache set error")
	
	config := &WarmupConfig{
		HotPostsLimit:   2,
		HotUsersLimit:   1,
		HotCirclesLimit: 1,
		Concurrency:     1,
	}
	service := NewWarmupService(cache, config)
	provider := newMockHotDataProvider()
	
	ctx := context.Background()
	err := service.WarmupOnStartup(ctx, provider)
	
	// Should not return error, just log warnings
	require.NoError(t, err)
	
	// Data should be empty due to cache set errors
	assert.Empty(t, cache.data)
}

func TestWarmupWithContextCancellation(t *testing.T) {
	cache := newMockCacheService()
	config := &WarmupConfig{
		HotPostsLimit:   100,
		HotUsersLimit:   50,
		HotCirclesLimit: 20,
		Concurrency:     1,
	}
	service := NewWarmupService(cache, config)
	provider := newMockHotDataProvider()
	provider.hotPostIDs = make([]int64, 100)
	for i := range provider.hotPostIDs {
		provider.hotPostIDs[i] = int64(i + 1)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	err := service.WarmupOnStartup(ctx, provider)
	
	// Should handle context cancellation gracefully
	require.NoError(t, err)
}

func TestRefreshEntity(t *testing.T) {
	cache := newMockCacheService()
	service := NewWarmupService(cache, DefaultWarmupConfig())
	ctx := context.Background()
	
	t.Run("refresh existing entity", func(t *testing.T) {
		loader := func(ctx context.Context) (interface{}, error) {
			return map[string]string{"name": "test"}, nil
		}
		
		err := service.RefreshEntity(ctx, "test:key", loader, time.Hour)
		require.NoError(t, err)
		assert.Contains(t, cache.data, "test:key")
	})
	
	t.Run("refresh with nil value removes from cache", func(t *testing.T) {
		cache.data["test:nil"] = "old value"
		
		loader := func(ctx context.Context) (interface{}, error) {
			return nil, nil
		}
		
		err := service.RefreshEntity(ctx, "test:nil", loader, time.Hour)
		require.NoError(t, err)
		assert.NotContains(t, cache.data, "test:nil")
	})
	
	t.Run("refresh with loader error", func(t *testing.T) {
		loader := func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("load error")
		}
		
		err := service.RefreshEntity(ctx, "test:error", loader, time.Hour)
		require.Error(t, err)
	})
}

func TestRefreshPost(t *testing.T) {
	cache := newMockCacheService()
	service := NewWarmupService(cache, DefaultWarmupConfig())
	ctx := context.Background()
	
	loader := func(ctx context.Context) (interface{}, error) {
		return map[string]interface{}{"id": 1, "title": "Test"}, nil
	}
	
	err := service.RefreshPost(ctx, 1, loader)
	require.NoError(t, err)
	assert.Contains(t, cache.data, "post:1")
}

func TestRefreshUser(t *testing.T) {
	cache := newMockCacheService()
	service := NewWarmupService(cache, DefaultWarmupConfig())
	ctx := context.Background()
	
	loader := func(ctx context.Context) (interface{}, error) {
		return map[string]interface{}{"id": 10, "username": "test"}, nil
	}
	
	err := service.RefreshUser(ctx, 10, loader)
	require.NoError(t, err)
	assert.Contains(t, cache.data, "user:10")
}

func TestRefreshCircle(t *testing.T) {
	cache := newMockCacheService()
	service := NewWarmupService(cache, DefaultWarmupConfig())
	ctx := context.Background()
	
	loader := func(ctx context.Context) (interface{}, error) {
		return map[string]interface{}{"id": 100, "name": "Test Circle"}, nil
	}
	
	err := service.RefreshCircle(ctx, 100, loader)
	require.NoError(t, err)
	assert.Contains(t, cache.data, "circle:100")
}

func TestStartAndStopPeriodicRefresh(t *testing.T) {
	cache := newMockCacheService()
	config := &WarmupConfig{
		HotPostsLimit:   2,
		HotUsersLimit:   1,
		HotCirclesLimit: 1,
		RefreshInterval: 100 * time.Millisecond,
		Concurrency:     2,
	}
	service := NewWarmupService(cache, config)
	provider := newMockHotDataProvider()
	
	// Start periodic refresh
	service.StartPeriodicRefresh(provider)
	assert.True(t, service.IsRunning())
	
	// Starting again should be a no-op
	service.StartPeriodicRefresh(provider)
	assert.True(t, service.IsRunning())
	
	// Wait for at least one refresh cycle
	time.Sleep(150 * time.Millisecond)
	
	// Stop the service
	service.Stop()
	assert.False(t, service.IsRunning())
	
	// Stopping again should be a no-op
	service.Stop()
	assert.False(t, service.IsRunning())
}

func TestWarmupEmptyIDs(t *testing.T) {
	cache := newMockCacheService()
	config := &WarmupConfig{
		HotPostsLimit:   10,
		HotUsersLimit:   10,
		HotCirclesLimit: 10,
		Concurrency:     2,
	}
	service := NewWarmupService(cache, config)
	provider := &mockHotDataProvider{
		hotPostIDs:       []int64{},
		activeUserIDs:    []int64{},
		popularCircleIDs: []int64{},
	}
	
	ctx := context.Background()
	err := service.WarmupOnStartup(ctx, provider)
	
	require.NoError(t, err)
	assert.Empty(t, cache.data)
}
