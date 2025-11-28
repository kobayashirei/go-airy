# Database Layer Setup

This document describes the database layer implementation for the Airy project.

## Overview

The database layer has been fully implemented with the following components:

1. **Database Connection Management** - GORM-based connection pool with health checks
2. **Data Models** - Complete set of 18 models covering all system entities
3. **Database Migrations** - SQL migration files for schema management

## Components

### 1. Database Connection (`internal/database/database.go`)

Features:
- GORM integration with MySQL driver
- Connection pool configuration (max idle/open connections, connection lifetime)
- Custom GORM logger integration with zap
- Health check functionality
- Graceful connection closing

Key Functions:
- `Init(cfg *config.DatabaseConfig)` - Initialize database connection
- `Close()` - Close database connection
- `HealthCheck(ctx context.Context)` - Check database health
- `GetDB()` - Get database instance

### 2. Data Models (`internal/models/`)

All models are organized by category:

#### User Models (`user.go`)
- `User` - User accounts with authentication info
- `UserProfile` - User profile data (points, level, followers)
- `UserStats` - User statistics (posts, comments, votes)

#### Permission Models (`permission.go`)
- `Role` - User roles (admin, moderator, user)
- `Permission` - System permissions
- `RolePermission` - Role-permission associations
- `UserRole` - User-role associations (with optional circle scope)

#### Content Models (`content.go`)
- `Post` - User posts with markdown/HTML content
- `Comment` - Hierarchical comments with path tracking
- `Vote` - Votes on posts/comments
- `Favorite` - User favorites
- `EntityCount` - Aggregated counts for entities

#### Circle Models (`circle.go`)
- `Circle` - Community circles/groups
- `CircleMember` - Circle membership

#### Notification Models (`notification.go`)
- `Notification` - User notifications
- `Conversation` - Private conversations
- `Message` - Conversation messages

#### Admin Models (`admin.go`)
- `AdminLog` - Administrative action audit logs

### 3. Database Migrations (`migrations/`)

Six migration files covering all tables:

1. **000001_create_users_tables** - User, UserProfile, UserStats
2. **000002_create_permission_tables** - Role, Permission, RolePermission, UserRole
3. **000003_create_circle_tables** - Circle, CircleMember
4. **000004_create_content_tables** - Post, Comment, Vote, Favorite, EntityCount
5. **000005_create_notification_tables** - Notification, Conversation, Message
6. **000006_create_admin_tables** - AdminLog

Each migration includes:
- Up migration (create tables)
- Down migration (drop tables)
- Proper indexes for performance
- Foreign key constraints for referential integrity
- Default values and constraints

### 4. Migration Management (`internal/database/migrate.go`)

Functions for programmatic migration management:
- `RunMigrations()` - Run all pending migrations
- `RollbackMigration()` - Rollback last migration
- `MigrationVersion()` - Get current migration version

## Usage

### Initialize Database Connection

```go
import (
    "github.com/kobayashirei/airy/internal/config"
    "github.com/kobayashirei/airy/internal/database"
)

// Load configuration
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// Initialize database
if err := database.Init(&cfg.Database); err != nil {
    log.Fatal(err)
}
defer database.Close()
```

### Run Migrations

Using the Go API:
```go
err := database.RunMigrations(&cfg.Database, "migrations")
if err != nil {
    log.Fatal(err)
}
```

Using the Makefile:
```bash
# Set environment variables
export DB_USER=root
export DB_PASSWORD=password
export DB_HOST=localhost
export DB_PORT=3306
export DB_NAME=airygithub

# Run migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check migration version
make migrate-version

# Create new migration
make migrate-create NAME=add_new_feature
```

### Use Models

```go
import (
    "github.com/kobayashirei/airy/internal/database"
    "github.com/kobayashirei/airy/internal/models"
)

// Create a user
user := &models.User{
    Username:     "testuser",
    Email:        "test@example.com",
    PasswordHash: "hashed_password",
    Status:       "active",
}

db := database.GetDB()
result := db.Create(user)
if result.Error != nil {
    log.Fatal(result.Error)
}

// Query a user
var foundUser models.User
db.Where("email = ?", "test@example.com").First(&foundUser)
```

### Health Check

```go
import (
    "context"
    "time"
)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := database.HealthCheck(ctx); err != nil {
    log.Printf("Database health check failed: %v", err)
}
```

## Configuration

Database configuration is managed through environment variables:

```env
DB_HOST=localhost
DB_PORT=3306
DB_NAME=airygithub
DB_USER=root
DB_PASSWORD=password
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=3600
```

## Testing

All components include comprehensive tests:

```bash
# Run all tests
go test -v ./internal/database/... ./internal/models/...

# Run with short flag (skip integration tests)
go test -v -short ./internal/database/... ./internal/models/...

# Run with coverage
go test -v -coverprofile=coverage.out ./internal/database/... ./internal/models/...
go tool cover -html=coverage.out
```

## Database Schema

The complete database schema includes:

- **18 tables** covering all system entities
- **Foreign key constraints** for referential integrity
- **Indexes** on frequently queried columns
- **Unique constraints** for data integrity
- **Default values** for optional fields
- **Timestamps** with automatic updates

## Performance Considerations

1. **Connection Pooling**: Configured with max idle/open connections
2. **Prepared Statements**: Enabled in GORM for query caching
3. **Indexes**: Created on all foreign keys and frequently queried columns
4. **Separate Count Table**: `entity_counts` to avoid lock contention
5. **Skip Default Transaction**: Disabled for better performance

## Requirements Validation

This implementation satisfies the following requirements:

- **Requirement 19.1**: Transaction atomicity for data writes
- **Requirement 19.2**: Referential integrity through foreign keys
- **Requirement 19.3**: Separate count table for high-frequency updates
- **Requirement 19.4**: Cascade deletes and soft deletes support
- **Requirement 19.5**: Transaction rollback on errors

## Next Steps

With the database layer complete, you can now:

1. Implement Repository layer (Task 4)
2. Create Service layer business logic
3. Build API handlers
4. Add authentication and authorization
5. Implement caching layer

## References

- [GORM Documentation](https://gorm.io/docs/)
- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [MySQL Documentation](https://dev.mysql.com/doc/)
