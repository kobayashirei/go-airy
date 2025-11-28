# Database Migrations

This directory contains database migration files for the Airy project.

## Migration Files

Migrations are numbered sequentially and include both "up" and "down" files:

- `000001_create_users_tables.up.sql` / `000001_create_users_tables.down.sql` - User, UserProfile, UserStats tables
- `000002_create_permission_tables.up.sql` / `000002_create_permission_tables.down.sql` - Role, Permission, RolePermission, UserRole tables
- `000003_create_circle_tables.up.sql` / `000003_create_circle_tables.down.sql` - Circle, CircleMember tables
- `000004_create_content_tables.up.sql` / `000004_create_content_tables.down.sql` - Post, Comment, Vote, Favorite, EntityCount tables
- `000005_create_notification_tables.up.sql` / `000005_create_notification_tables.down.sql` - Notification, Conversation, Message tables
- `000006_create_admin_tables.up.sql` / `000006_create_admin_tables.down.sql` - AdminLog table

## Running Migrations

### Using the migrate CLI tool

Install the migrate tool:
```bash
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Run migrations:
```bash
migrate -path migrations -database "mysql://user:password@tcp(localhost:3306)/airygithub" up
```

Rollback migrations:
```bash
migrate -path migrations -database "mysql://user:password@tcp(localhost:3306)/airygithub" down
```

Check migration version:
```bash
migrate -path migrations -database "mysql://user:password@tcp(localhost:3306)/airygithub" version
```

### Using the Go API

The `internal/database` package provides helper functions:

```go
import "github.com/kobayashirei/airy/internal/database"

// Run all pending migrations
err := database.RunMigrations(cfg, "migrations")

// Rollback the last migration
err := database.RollbackMigration(cfg, "migrations")

// Get current migration version
version, dirty, err := database.MigrationVersion(cfg, "migrations")
```

## Creating New Migrations

To create a new migration:

1. Create two files with the next sequential number:
   - `00000X_description.up.sql` - Contains the migration
   - `00000X_description.down.sql` - Contains the rollback

2. Write your SQL in the up file
3. Write the reverse SQL in the down file

Example:
```sql
-- 000007_add_user_email_verified.up.sql
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;

-- 000007_add_user_email_verified.down.sql
ALTER TABLE users DROP COLUMN email_verified;
```

## Database Schema

The migrations create the following tables:

### User Tables
- `users` - User accounts
- `user_profiles` - User profile information
- `user_stats` - User statistics

### Permission Tables
- `roles` - User roles
- `permissions` - System permissions
- `role_permissions` - Role-permission associations
- `user_roles` - User-role associations

### Circle Tables
- `circles` - Community circles
- `circle_members` - Circle membership

### Content Tables
- `posts` - User posts
- `comments` - Post comments
- `votes` - Votes on posts/comments
- `favorites` - User favorites
- `entity_counts` - Aggregated counts

### Notification Tables
- `notifications` - User notifications
- `conversations` - Private conversations
- `messages` - Conversation messages

### Admin Tables
- `admin_logs` - Administrative action logs

## Notes

- All tables use `utf8mb4` character set with `utf8mb4_unicode_ci` collation
- Foreign keys are set up with appropriate CASCADE rules
- Indexes are created for frequently queried columns
- Timestamps use `DATETIME` type with automatic updates
