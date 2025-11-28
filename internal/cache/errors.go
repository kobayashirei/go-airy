package cache

import "errors"

var (
	// ErrCacheMiss is returned when a cache key is not found
	ErrCacheMiss = errors.New("cache miss")

	// ErrCacheNotInitialized is returned when cache client is not initialized
	ErrCacheNotInitialized = errors.New("cache client not initialized")

	// ErrInvalidValue is returned when the cached value cannot be unmarshaled
	ErrInvalidValue = errors.New("invalid cached value")
)
