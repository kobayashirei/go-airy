# Cache Package

This package provides Redis-based caching functionality with Cache-Aside pattern implementation for the Airy backend system.

## Features

- **Redis Client Management**: Connection pooling, health checks, and graceful shutdown
- **Cache Service Interface**: Generic cache operations (Get, Set, Delete, Exists, SetNX, Expire, TTL)
- **Cache-Aside Pattern**: Automatic cache miss handling with database fallback
- **Key Generator**: Consistent cache key generation for different entity types
- **Entity Cache Service**: High-level cache operations for specific entities (User, Post, Circle, etc.)
- **Cache Warmup**: Pre-load frequently accessed data into cache
- **Cache Invalidation**: Automatic cache invalidation on data updates

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│  (Handlers, Services)                                        │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              EntityCacheService                              │
│  (High-level entity-specific cache operations)              │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              CacheAsideService                               │
│  (Cache-Aside pattern implementation)                       │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              CacheService                                    │
│  (Low-level Redis operations)                               │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              Redis Client                                    │
└─────────────────────────────────────────────────────────────┘
```

## Usage

### Initialization

```go
import (
    "github.com/kobayashirei/airy/internal/cache"
    "github.com/kobayashirei/airy/internal/config"
)

// Initialize Redis connection
cfg := &config.RedisConfig{
    Host:     "localhost",
    Port:     6379,
    Password: "",
    DB:       0,
}

cacheConfig := &config.CacheConfig{
    DefaultExpiration: 1 * time.Hour,
    CleanupInterval:   10 * time.Minute,
}

if err := cache.Init(cfg, cacheConfig); err != nil {
    log.Fatal(err)
}
defer cache.Close()
```

### Basic Cache Operations

```go
// Create cache service
client := cache.GetClient()
cacheService := cache.NewCacheService(client, 1*time.Hour)

// Set a value
ctx := context.Background()
err := cacheService.Set(ctx, "user:123", userData, 30*time.Minute)

// Get a value
var user User
err := cacheService.Get(ctx, "user:123", &user)
if err == cache.ErrCacheMiss {
    // Cache miss - load from database
}

// Delete a value
err := cacheService.Delete(ctx, "user:123")

// Check if key exists
exists, err := cacheService.Exists(ctx, "user:123")

// Set only if not exists (atomic)
success, err := cacheService.SetNX(ctx, "lock:resource", "locked", 10*time.Second)
```

### Cache-Aside Pattern

```go
// Create cache-aside service
cacheAside := cache.NewCacheAsideService(cacheService, 1*time.Hour)

// GetOrLoad automatically handles cache miss
var user User
err := cacheAside.GetOrLoad(
    ctx,
    "user:123",
    &user,
    func(ctx context.Context) (interface{}, error) {
        // This function is only called on cache miss
        return userRepo.FindByID(ctx, 123)
    },
    30*time.Minute,
)

// Invalidate cache when data is updated
err := cacheAside.Invalidate(ctx, "user:123")

// Invalidate multiple keys
keys := []string{"user:123", "user_profile:123", "user_stats:123"}
err := cacheAside.InvalidateMultiple(ctx, keys)
```

### Entity Cache Service

```go
// Create entity cache service
entityCache := cache.NewEntityCacheService(cacheService, 1*time.Hour)

// Get user with automatic cache-aside
user, err := entityCache.GetUser(ctx, 123, func(ctx context.Context) (*models.User, error) {
    return userRepo.FindByID(ctx, 123)
})

// Invalidate user and related caches
err := entityCache.InvalidateUser(ctx, 123)

// Get post with automatic cache-aside
post, err := entityCache.GetPost(ctx, 456, func(ctx context.Context) (*models.Post, error) {
    return postRepo.FindByID(ctx, 456)
})

// Session management
err := entityCache.SetSession(ctx, "token123", 123, 24*time.Hour)
userID, err := entityCache.GetSession(ctx, "token123")
err := entityCache.DeleteSession(ctx, "token123")

// Verification code management
err := entityCache.SetVerificationCode(ctx, "user@example.com", "123456", 5*time.Minute)
code, err := entityCache.GetVerificationCode(ctx, "user@example.com")
err := entityCache.DeleteVerificationCode(ctx, "user@example.com")
```

### Cache Warmup

```go
// Warmup single item
err := cacheAside.Warmup(
    ctx,
    "user:123",
    func(ctx context.Context) (interface{}, error) {
        return userRepo.FindByID(ctx, 123)
    },
    1*time.Hour,
)

// Warmup batch
items := []cache.WarmupItem{
    {
        Key: "user:123",
        Loader: func(ctx context.Context) (interface{}, error) {
            return userRepo.FindByID(ctx, 123)
        },
        Expiration: 1 * time.Hour,
    },
    // ... more items
}
err := cacheAside.WarmupBatch(ctx, items)

// Warmup hot data (entity cache service)
err := entityCache.WarmupHotData(
    ctx,
    []int64{1, 2, 3}, // hot user IDs
    []int64{10, 20, 30}, // hot post IDs
    []int64{100, 200}, // hot circle IDs
    userLoader,
    postLoader,
    circleLoader,
)
```

### Key Generation

```go
keyGen := cache.NewKeyGenerator()

// Generate cache keys
userKey := keyGen.UserKey(123)                    // "user:123"
postKey := keyGen.PostKey(456)                    // "post:456"
countKey := keyGen.PostCountKey(456)              // "count:post:456"
feedKey := keyGen.UserFeedKey(123)                // "timeline:user:123"
sessionKey := keyGen.SessionKey("token123")       // "session:token123"
codeKey := keyGen.VerificationCodeKey("email")    // "code:email"
```

## Cache Key Patterns

| Entity Type | Key Pattern | Example | Expiration |
|------------|-------------|---------|------------|
| User | `user:{id}` | `user:123` | 1 hour |
| User Profile | `user_profile:{id}` | `user_profile:123` | 1 hour |
| User Stats | `user_stats:{id}` | `user_stats:123` | 1 hour |
| Post | `post:{id}` | `post:456` | 30 minutes |
| Comment | `comment:{id}` | `comment:789` | 30 minutes |
| Circle | `circle:{id}` | `circle:100` | 1 hour |
| Post Count | `count:post:{id}` | `count:post:456` | 10 minutes |
| Comment Count | `count:comment:{id}` | `count:comment:789` | 10 minutes |
| User Feed | `timeline:user:{id}` | `timeline:user:123` | 5 minutes |
| Circle Feed | `timeline:circle:{id}` | `timeline:circle:100` | 5 minutes |
| Session | `session:{token}` | `session:abc123` | 24 hours |
| Verification Code | `code:{identifier}` | `code:user@example.com` | 5 minutes |
| Activation Token | `token:activation:{token}` | `token:activation:xyz` | 24 hours |

## Cache Invalidation Strategy

When data is updated, invalidate related cache entries:

### User Update
- `user:{id}`
- `user_profile:{id}`
- `user_stats:{id}`

### Post Update
- `post:{id}`
- `count:post:{id}`
- `timeline:user:{author_id}`
- `timeline:circle:{circle_id}` (if applicable)

### Comment Update
- `comment:{id}`
- `count:comment:{id}`
- `count:post:{post_id}` (parent post count)

### Vote Update
- `count:post:{id}` or `count:comment:{id}`

## Error Handling

```go
err := cacheService.Get(ctx, "key", &dest)
if err != nil {
    if err == cache.ErrCacheMiss {
        // Cache miss - load from database
    } else {
        // Other cache error - log and continue
        log.Warn("Cache error", zap.Error(err))
    }
}
```

## Testing

Run tests with Redis available:

```bash
# Run all cache tests
go test ./internal/cache/... -v

# Run specific test
go test ./internal/cache -run TestCacheService_SetAndGet -v

# Run tests with coverage
go test ./internal/cache/... -cover
```

Tests require Redis running on `localhost:6379`. Tests use DB 15 to avoid conflicts with production data.

## Best Practices

1. **Always handle cache misses gracefully** - Cache should be transparent to the application
2. **Use appropriate expiration times** - Balance between freshness and cache hit rate
3. **Invalidate cache on updates** - Maintain cache consistency
4. **Use Cache-Aside pattern** - Simplifies cache management
5. **Warmup hot data on startup** - Improve initial response times
6. **Monitor cache hit rate** - Optimize cache strategy based on metrics
7. **Use consistent key patterns** - Makes debugging and monitoring easier
8. **Handle cache errors gracefully** - Don't fail requests due to cache errors

## Configuration

Environment variables:

```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
CACHE_DEFAULT_EXPIRATION=3600  # seconds
CACHE_CLEANUP_INTERVAL=600     # seconds
```

## Performance Considerations

- **Connection Pooling**: Redis client uses connection pooling (default: 10 connections)
- **Serialization**: JSON serialization for complex objects (consider msgpack for better performance)
- **Batch Operations**: Use batch operations when possible to reduce round trips
- **Pipeline**: Consider using Redis pipeline for multiple operations
- **Compression**: Consider compressing large values before caching

## Monitoring

Monitor these metrics:

- Cache hit rate
- Cache miss rate
- Average response time
- Connection pool usage
- Memory usage
- Eviction rate

## Future Enhancements

- [ ] Add support for Redis Cluster
- [ ] Implement cache compression for large values
- [ ] Add cache statistics and metrics
- [ ] Implement distributed locking
- [ ] Add support for Redis Streams
- [ ] Implement cache versioning for schema changes
