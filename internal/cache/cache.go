package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/kobayashirei/airy/internal/config"
	appLogger "github.com/kobayashirei/airy/internal/logger"
)

// Client is the global Redis client instance
var Client *redis.Client

// Service defines the cache service interface
type Service interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
}

// CacheService implements the Service interface
type CacheService struct {
	client            *redis.Client
	defaultExpiration time.Duration
}

// Init initializes the Redis client connection
func Init(cfg *config.RedisConfig, cacheConfig *config.CacheConfig) error {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.GetAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	Client = client

	appLogger.Info("Redis connection established",
		zap.String("addr", cfg.GetAddr()),
		zap.Int("db", cfg.DB),
		zap.Duration("default_expiration", cacheConfig.DefaultExpiration),
	)

	return nil
}

// Close closes the Redis client connection
func Close() error {
	if Client == nil {
		return nil
	}

	if err := Client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis client: %w", err)
	}

	appLogger.Info("Redis connection closed")
	return nil
}

// HealthCheck checks if Redis is healthy
func HealthCheck(ctx context.Context) error {
	if Client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis ping failed: %w", err)
	}

	return nil
}

// GetClient returns the Redis client instance
func GetClient() *redis.Client {
	return Client
}

// NewCacheService creates a new cache service instance
func NewCacheService(client *redis.Client, defaultExpiration time.Duration) Service {
	return &CacheService{
		client:            client,
		defaultExpiration: defaultExpiration,
	}
}

// Get retrieves a value from cache and unmarshals it into dest
func (s *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get cache key %s: %w", key, err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache value for key %s: %w", key, err)
	}

	appLogger.Debug("Cache hit", zap.String("key", key))
	return nil
}

// Set stores a value in cache with the specified expiration
func (s *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value for key %s: %w", key, err)
	}

	if expiration == 0 {
		expiration = s.defaultExpiration
	}

	if err := s.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set cache key %s: %w", key, err)
	}

	appLogger.Debug("Cache set", zap.String("key", key), zap.Duration("expiration", expiration))
	return nil
}

// Delete removes a key from cache
func (s *CacheService) Delete(ctx context.Context, key string) error {
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete cache key %s: %w", key, err)
	}

	appLogger.Debug("Cache deleted", zap.String("key", key))
	return nil
}

// Exists checks if a key exists in cache
func (s *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	count, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache key existence %s: %w", key, err)
	}

	return count > 0, nil
}

// SetNX sets a key only if it doesn't exist (atomic operation)
func (s *CacheService) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal cache value for key %s: %w", key, err)
	}

	if expiration == 0 {
		expiration = s.defaultExpiration
	}

	success, err := s.client.SetNX(ctx, key, data, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set cache key %s with NX: %w", key, err)
	}

	if success {
		appLogger.Debug("Cache set with NX", zap.String("key", key), zap.Duration("expiration", expiration))
	}

	return success, nil
}

// Expire sets a timeout on a key
func (s *CacheService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if err := s.client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set expiration for cache key %s: %w", key, err)
	}

	appLogger.Debug("Cache expiration set", zap.String("key", key), zap.Duration("expiration", expiration))
	return nil
}

// TTL returns the remaining time to live of a key
func (s *CacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for cache key %s: %w", key, err)
	}

	return ttl, nil
}
