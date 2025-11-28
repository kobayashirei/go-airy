package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	appLogger "github.com/kobayashirei/airy/internal/logger"
	"github.com/kobayashirei/airy/internal/response"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	// Enabled indicates whether rate limiting is enabled
	Enabled bool
	// RequestsPerSecond is the maximum number of requests per second
	RequestsPerSecond int
	// BurstSize is the maximum burst size
	BurstSize int
	// WindowSize is the time window for rate limiting
	WindowSize time.Duration
	// KeyPrefix is the prefix for rate limit keys in Redis
	KeyPrefix string
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 100,
		BurstSize:         200,
		WindowSize:        time.Second,
		KeyPrefix:         "ratelimit",
	}
}

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	// Allow checks if a request is allowed for the given key
	Allow(ctx context.Context, key string) (bool, error)
	// AllowN checks if n requests are allowed for the given key
	AllowN(ctx context.Context, key string, n int) (bool, error)
	// Reset resets the rate limit for the given key
	Reset(ctx context.Context, key string) error
	// GetRemaining returns the remaining requests for the given key
	GetRemaining(ctx context.Context, key string) (int, error)
}

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	client     *redis.Client
	config     *RateLimitConfig
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client, config *RateLimitConfig) *RedisRateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}
	return &RedisRateLimiter{
		client: client,
		config: config,
	}
}

// Allow checks if a request is allowed for the given key using sliding window algorithm
func (r *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return r.AllowN(ctx, key, 1)
}

// AllowN checks if n requests are allowed for the given key
func (r *RedisRateLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	fullKey := fmt.Sprintf("%s:%s", r.config.KeyPrefix, key)
	now := time.Now().UnixNano()
	windowStart := now - r.config.WindowSize.Nanoseconds()

	// Use Redis pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, fullKey, "0", fmt.Sprintf("%d", windowStart))

	// Count current entries in the window
	countCmd := pipe.ZCard(ctx, fullKey)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("failed to execute rate limit pipeline: %w", err)
	}

	count := countCmd.Val()
	limit := int64(r.config.RequestsPerSecond)

	// Check if we're over the limit
	if count+int64(n) > limit {
		return false, nil
	}

	// Add new entries
	pipe2 := r.client.Pipeline()
	for i := 0; i < n; i++ {
		member := fmt.Sprintf("%d-%d", now, i)
		pipe2.ZAdd(ctx, fullKey, redis.Z{Score: float64(now), Member: member})
	}
	pipe2.Expire(ctx, fullKey, r.config.WindowSize*2)

	_, err = pipe2.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to add rate limit entries: %w", err)
	}

	return true, nil
}

// Reset resets the rate limit for the given key
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := fmt.Sprintf("%s:%s", r.config.KeyPrefix, key)
	return r.client.Del(ctx, fullKey).Err()
}

// GetRemaining returns the remaining requests for the given key
func (r *RedisRateLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	fullKey := fmt.Sprintf("%s:%s", r.config.KeyPrefix, key)
	now := time.Now().UnixNano()
	windowStart := now - r.config.WindowSize.Nanoseconds()

	// Remove old entries and count current
	pipe := r.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, fullKey, "0", fmt.Sprintf("%d", windowStart))
	countCmd := pipe.ZCard(ctx, fullKey)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return 0, err
	}

	count := int(countCmd.Val())
	remaining := r.config.RequestsPerSecond - count
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

// InMemoryRateLimiter implements rate limiting using in-memory storage
// Useful for single-instance deployments or testing
type InMemoryRateLimiter struct {
	config   *RateLimitConfig
	buckets  map[string]*tokenBucket
	mu       sync.RWMutex
}

type tokenBucket struct {
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(config *RateLimitConfig) *InMemoryRateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}
	return &InMemoryRateLimiter{
		config:  config,
		buckets: make(map[string]*tokenBucket),
	}
}

// Allow checks if a request is allowed for the given key
func (r *InMemoryRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return r.AllowN(ctx, key, 1)
}

// AllowN checks if n requests are allowed for the given key
func (r *InMemoryRateLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	r.mu.Lock()
	bucket, exists := r.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     float64(r.config.BurstSize),
			lastUpdate: time.Now(),
		}
		r.buckets[key] = bucket
	}
	r.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	bucket.lastUpdate = now

	// Refill tokens based on elapsed time
	bucket.tokens += elapsed * float64(r.config.RequestsPerSecond)
	if bucket.tokens > float64(r.config.BurstSize) {
		bucket.tokens = float64(r.config.BurstSize)
	}

	// Check if we have enough tokens
	if bucket.tokens < float64(n) {
		return false, nil
	}

	bucket.tokens -= float64(n)
	return true, nil
}

// Reset resets the rate limit for the given key
func (r *InMemoryRateLimiter) Reset(ctx context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.buckets, key)
	return nil
}

// GetRemaining returns the remaining requests for the given key
func (r *InMemoryRateLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	r.mu.RLock()
	bucket, exists := r.buckets[key]
	r.mu.RUnlock()

	if !exists {
		return r.config.BurstSize, nil
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	tokens := bucket.tokens + elapsed*float64(r.config.RequestsPerSecond)
	if tokens > float64(r.config.BurstSize) {
		tokens = float64(r.config.BurstSize)
	}

	return int(tokens), nil
}

// Cleanup removes expired buckets from memory
func (r *InMemoryRateLimiter) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	threshold := time.Now().Add(-r.config.WindowSize * 10)
	for key, bucket := range r.buckets {
		bucket.mu.Lock()
		if bucket.lastUpdate.Before(threshold) {
			delete(r.buckets, key)
		}
		bucket.mu.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter RateLimiter, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		
		allowed, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			appLogger.Warn("Rate limit check failed",
				zap.String("key", key),
				zap.Error(err),
			)
			// On error, allow the request but log the issue
			c.Next()
			return
		}

		if !allowed {
			remaining, _ := limiter.GetRemaining(c.Request.Context(), key)
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("Retry-After", "1")
			response.Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests, please try again later", nil)
			c.Abort()
			return
		}

		remaining, _ := limiter.GetRemaining(c.Request.Context(), key)
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Next()
	}
}

// IPRateLimitMiddleware creates a rate limiting middleware based on client IP
func IPRateLimitMiddleware(limiter RateLimiter) gin.HandlerFunc {
	return RateLimitMiddleware(limiter, func(c *gin.Context) string {
		return "ip:" + c.ClientIP()
	})
}

// UserRateLimitMiddleware creates a rate limiting middleware based on user ID
// Falls back to IP-based limiting if user is not authenticated
func UserRateLimitMiddleware(limiter RateLimiter) gin.HandlerFunc {
	return RateLimitMiddleware(limiter, func(c *gin.Context) string {
		userID, exists := c.Get("userID")
		if exists {
			return fmt.Sprintf("user:%v", userID)
		}
		return "ip:" + c.ClientIP()
	})
}

// EndpointRateLimitMiddleware creates a rate limiting middleware for specific endpoints
func EndpointRateLimitMiddleware(limiter RateLimiter, endpoint string) gin.HandlerFunc {
	return RateLimitMiddleware(limiter, func(c *gin.Context) string {
		userID, exists := c.Get("userID")
		if exists {
			return fmt.Sprintf("endpoint:%s:user:%v", endpoint, userID)
		}
		return fmt.Sprintf("endpoint:%s:ip:%s", endpoint, c.ClientIP())
	})
}

// CompositeRateLimitMiddleware applies multiple rate limiters
func CompositeRateLimitMiddleware(limiters ...struct {
	Limiter RateLimiter
	KeyFunc func(*gin.Context) string
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, l := range limiters {
			key := l.KeyFunc(c)
			allowed, err := l.Limiter.Allow(c.Request.Context(), key)
			if err != nil {
				appLogger.Warn("Rate limit check failed",
					zap.String("key", key),
					zap.Error(err),
				)
				continue
			}

			if !allowed {
				remaining, _ := l.Limiter.GetRemaining(c.Request.Context(), key)
				c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
				c.Header("Retry-After", "1")
				response.Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests, please try again later", nil)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
