package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
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
	gin.SetMode(gin.TestMode)
}

func TestDefaultRateLimitConfig(t *testing.T) {
	config := DefaultRateLimitConfig()
	
	assert.True(t, config.Enabled)
	assert.Equal(t, 100, config.RequestsPerSecond)
	assert.Equal(t, 200, config.BurstSize)
	assert.Equal(t, time.Second, config.WindowSize)
	assert.Equal(t, "ratelimit", config.KeyPrefix)
}

func TestInMemoryRateLimiter_Allow(t *testing.T) {
	config := &RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10,
		BurstSize:         10,
		WindowSize:        time.Second,
		KeyPrefix:         "test",
	}
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	t.Run("allows requests within limit", func(t *testing.T) {
		key := "test-allow"
		for i := 0; i < 10; i++ {
			allowed, err := limiter.Allow(ctx, key)
			require.NoError(t, err)
			assert.True(t, allowed, "request %d should be allowed", i)
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		key := "test-block"
		// Exhaust the bucket
		for i := 0; i < 10; i++ {
			limiter.Allow(ctx, key)
		}
		
		// Next request should be blocked
		allowed, err := limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("refills tokens over time", func(t *testing.T) {
		key := "test-refill"
		// Exhaust the bucket
		for i := 0; i < 10; i++ {
			limiter.Allow(ctx, key)
		}
		
		// Wait for refill
		time.Sleep(200 * time.Millisecond)
		
		// Should have some tokens now
		allowed, err := limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.True(t, allowed)
	})
}

func TestInMemoryRateLimiter_AllowN(t *testing.T) {
	config := &RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10,
		BurstSize:         10,
		WindowSize:        time.Second,
		KeyPrefix:         "test",
	}
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	t.Run("allows batch requests within limit", func(t *testing.T) {
		key := "test-batch-allow"
		allowed, err := limiter.AllowN(ctx, key, 5)
		require.NoError(t, err)
		assert.True(t, allowed)
		
		allowed, err = limiter.AllowN(ctx, key, 5)
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("blocks batch requests over limit", func(t *testing.T) {
		key := "test-batch-block"
		allowed, err := limiter.AllowN(ctx, key, 5)
		require.NoError(t, err)
		assert.True(t, allowed)
		
		// This should fail as we only have 5 tokens left
		allowed, err = limiter.AllowN(ctx, key, 6)
		require.NoError(t, err)
		assert.False(t, allowed)
	})
}

func TestInMemoryRateLimiter_Reset(t *testing.T) {
	config := &RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 5,
		BurstSize:         5,
		WindowSize:        time.Second,
		KeyPrefix:         "test",
	}
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	key := "test-reset"
	
	// Exhaust the bucket
	for i := 0; i < 5; i++ {
		limiter.Allow(ctx, key)
	}
	
	// Should be blocked
	allowed, _ := limiter.Allow(ctx, key)
	assert.False(t, allowed)
	
	// Reset
	err := limiter.Reset(ctx, key)
	require.NoError(t, err)
	
	// Should be allowed again
	allowed, err = limiter.Allow(ctx, key)
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestInMemoryRateLimiter_GetRemaining(t *testing.T) {
	config := &RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10,
		BurstSize:         10,
		WindowSize:        time.Second,
		KeyPrefix:         "test",
	}
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	key := "test-remaining"
	
	// Initially should have full bucket
	remaining, err := limiter.GetRemaining(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 10, remaining)
	
	// Use some tokens
	limiter.AllowN(ctx, key, 3)
	
	remaining, err = limiter.GetRemaining(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 7, remaining)
}

func TestInMemoryRateLimiter_Cleanup(t *testing.T) {
	config := &RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10,
		BurstSize:         10,
		WindowSize:        10 * time.Millisecond,
		KeyPrefix:         "test",
	}
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	// Create some buckets
	limiter.Allow(ctx, "key1")
	limiter.Allow(ctx, "key2")
	
	assert.Len(t, limiter.buckets, 2)
	
	// Wait for cleanup threshold
	time.Sleep(150 * time.Millisecond)
	
	// Run cleanup
	limiter.Cleanup()
	
	// Buckets should be removed
	assert.Empty(t, limiter.buckets)
}

func TestRateLimitMiddleware(t *testing.T) {
	config := &RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 2,
		BurstSize:         2,
		WindowSize:        time.Second,
		KeyPrefix:         "test",
	}
	limiter := NewInMemoryRateLimiter(config)

	router := gin.New()
	// Add request_id middleware for tests
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		c.Next()
	})
	router.Use(RateLimitMiddleware(limiter, func(c *gin.Context) string {
		return "test-key"
	}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	t.Run("allows requests within limit", func(t *testing.T) {
		limiter.Reset(context.Background(), "test-key")
		
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		limiter.Reset(context.Background(), "test-key")
		
		// Exhaust limit
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)
		}
		
		// This should be blocked
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, w.Header().Get("Retry-After"))
	})
}

func TestIPRateLimitMiddleware(t *testing.T) {
	config := &RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 5,
		BurstSize:         5,
		WindowSize:        time.Second,
		KeyPrefix:         "test",
	}
	limiter := NewInMemoryRateLimiter(config)

	router := gin.New()
	router.Use(IPRateLimitMiddleware(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserRateLimitMiddleware(t *testing.T) {
	config := &RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 5,
		BurstSize:         5,
		WindowSize:        time.Second,
		KeyPrefix:         "test",
	}
	limiter := NewInMemoryRateLimiter(config)

	t.Run("uses user ID when authenticated", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", int64(123))
			c.Next()
		})
		router.Use(UserRateLimitMiddleware(limiter))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("falls back to IP when not authenticated", func(t *testing.T) {
		router := gin.New()
		router.Use(UserRateLimitMiddleware(limiter))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestEndpointRateLimitMiddleware(t *testing.T) {
	config := &RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 2,
		BurstSize:         2,
		WindowSize:        time.Second,
		KeyPrefix:         "test",
	}
	limiter := NewInMemoryRateLimiter(config)

	router := gin.New()
	// Add request_id middleware for tests
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		c.Next()
	})
	router.POST("/login", EndpointRateLimitMiddleware(limiter, "login"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First two requests should succeed
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Third request should be rate limited
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
