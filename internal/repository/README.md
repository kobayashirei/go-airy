# Repository Layer

This package contains the repository layer implementations for the Airy backend system. The repository layer provides a clean abstraction over database operations using GORM.

## Architecture

The repository layer follows the Repository pattern, providing:
- Clean separation between business logic and data access
- Interface-based design for easy testing and mocking
- Context-aware operations for cancellation and timeouts
- Consistent error handling

## Implemented Repositories

### User Repositories

#### UserRepository
Manages user account data operations.

**Methods:**
- `Create` - Create a new user
- `FindByID` - Find user by ID
- `FindByEmail` - Find user by email
- `FindByPhone` - Find user by phone number
- `FindByUsername` - Find user by username
- `Update` - Update user information
- `Delete` - Soft delete a user
- `UpdateLoginInfo` - Update last login time and IP

#### UserProfileRepository
Manages user profile data including points, level, and follower counts.

**Methods:**
- `Create` - Create a new user profile
- `FindByUserID` - Find profile by user ID
- `Update` - Update profile information
- `IncrementFollowerCount` - Increment/decrement follower count
- `IncrementFollowingCount` - Increment/decrement following count
- `UpdatePoints` - Update user points

#### UserStatsRepository
Manages user statistics including post, comment, and vote counts.

**Methods:**
- `Create` - Create a new user stats record
- `FindByUserID` - Find stats by user ID
- `Update` - Update stats
- `IncrementPostCount` - Increment/decrement post count
- `IncrementCommentCount` - Increment/decrement comment count
- `IncrementVoteReceivedCount` - Increment/decrement vote received count

### Permission Repositories

#### RoleRepository
Manages user roles in the system.

**Methods:**
- `Create` - Create a new role
- `FindByID` - Find role by ID
- `FindByName` - Find role by name
- `FindAll` - Retrieve all roles
- `Update` - Update role information
- `Delete` - Delete a role

#### PermissionRepository
Manages permissions in the system.

**Methods:**
- `Create` - Create a new permission
- `FindByID` - Find permission by ID
- `FindByName` - Find permission by name
- `FindAll` - Retrieve all permissions
- `FindByRoleID` - Find all permissions for a role
- `Update` - Update permission information
- `Delete` - Delete a permission

#### UserRoleRepository
Manages user-role associations including circle-specific roles.

**Methods:**
- `Create` - Create a new user-role association
- `FindByUserID` - Find all roles for a user
- `FindByUserIDAndCircleID` - Find roles for a user in a specific circle
- `FindRolesByUserID` - Find all role objects for a user
- `Delete` - Remove a user-role association
- `DeleteByUserID` - Remove all roles for a user

### Content Repositories

#### PostRepository
Manages post/article data operations.

**Methods:**
- `Create` - Create a new post
- `FindByID` - Find post by ID
- `Update` - Update post information
- `Delete` - Soft delete a post
- `List` - List posts with filtering and pagination
- `Count` - Count posts matching criteria
- `IncrementViewCount` - Increment view count
- `UpdateHotnessScore` - Update hotness score
- `UpdateStatus` - Update post status

#### CommentRepository
Manages comment data operations including hierarchical comments.

**Methods:**
- `Create` - Create a new comment
- `FindByID` - Find comment by ID
- `FindByPostID` - Find all comments for a post
- `FindByParentID` - Find direct replies to a comment
- `FindRootComments` - Find root-level comments
- `Update` - Update comment information
- `Delete` - Soft delete a comment
- `UpdateStatus` - Update comment status
- `CountByPostID` - Count comments for a post

#### VoteRepository
Manages vote data operations with idempotency support.

**Methods:**
- `Create` - Create a new vote
- `FindByID` - Find vote by ID
- `FindByUserAndEntity` - Find vote by user and entity
- `Update` - Update vote information
- `Upsert` - Create or update vote (idempotent)
- `Delete` - Delete a vote
- `DeleteByUserAndEntity` - Delete vote by user and entity
- `CountByEntity` - Count votes for an entity

#### EntityCountRepository
Manages aggregated counts for posts and comments.

**Methods:**
- `Create` - Create a new entity count record
- `FindByEntity` - Find count by entity type and ID
- `Update` - Update count information
- `Upsert` - Create or update count record
- `IncrementUpvoteCount` - Increment/decrement upvote count
- `IncrementDownvoteCount` - Increment/decrement downvote count
- `IncrementCommentCount` - Increment/decrement comment count
- `IncrementFavoriteCount` - Increment/decrement favorite count

### Community Repositories

#### CircleRepository
Manages circle/community data operations.

**Methods:**
- `Create` - Create a new circle
- `FindByID` - Find circle by ID
- `FindByName` - Find circle by name
- `FindAll` - List all circles with pagination
- `FindByCreatorID` - Find circles created by a user
- `Update` - Update circle information
- `Delete` - Delete a circle
- `IncrementMemberCount` - Increment/decrement member count
- `IncrementPostCount` - Increment/decrement post count
- `Count` - Count total circles

#### CircleMemberRepository
Manages circle membership data operations.

**Methods:**
- `Create` - Create a new membership record
- `FindByID` - Find membership by ID
- `FindByCircleAndUser` - Find membership by circle and user
- `FindByCircleID` - Find all members of a circle
- `FindByUserID` - Find all circles a user is member of
- `Update` - Update membership information
- `UpdateRole` - Update member role
- `Delete` - Delete a membership
- `DeleteByCircleAndUser` - Delete membership by circle and user
- `CountByCircleID` - Count members in a circle
- `IsMember` - Check if user is a member

### Notification and Message Repositories

#### NotificationRepository
Manages notification data operations.

**Methods:**
- `Create` - Create a new notification
- `FindByID` - Find notification by ID
- `FindByReceiverID` - Find all notifications for a receiver
- `FindUnreadByReceiverID` - Find unread notifications
- `Update` - Update notification information
- `MarkAsRead` - Mark notification as read
- `MarkAllAsRead` - Mark all notifications as read
- `Delete` - Delete a notification
- `CountUnreadByReceiverID` - Count unread notifications

#### ConversationRepository
Manages private conversation data operations.

**Methods:**
- `Create` - Create a new conversation
- `FindByID` - Find conversation by ID
- `FindByUsers` - Find conversation between two users
- `FindByUserID` - Find all conversations for a user
- `Update` - Update conversation information
- `UpdateLastMessageAt` - Update last message timestamp
- `Delete` - Delete a conversation

#### MessageRepository
Manages message data operations.

**Methods:**
- `Create` - Create a new message
- `FindByID` - Find message by ID
- `FindByConversationID` - Find all messages in a conversation
- `FindUnreadByConversationAndReceiver` - Find unread messages
- `Update` - Update message information
- `MarkAsRead` - Mark message as read
- `MarkConversationAsRead` - Mark all messages in conversation as read
- `Delete` - Delete a message
- `CountUnreadByReceiver` - Count unread messages for a receiver

## Usage Example

```go
package main

import (
    "context"
    "github.com/kobayashirei/airy/internal/repository"
    "gorm.io/gorm"
)

func example(db *gorm.DB) {
    // Create repository instances
    userRepo := repository.NewUserRepository(db)
    postRepo := repository.NewPostRepository(db)
    
    ctx := context.Background()
    
    // Find a user
    user, err := userRepo.FindByEmail(ctx, "user@example.com")
    if err != nil {
        // handle error
    }
    
    // List posts
    posts, err := postRepo.List(ctx, repository.PostListOptions{
        Status: "published",
        Limit:  10,
        Offset: 0,
        SortBy: "created_at",
        Order:  "DESC",
    })
    if err != nil {
        // handle error
    }
}
```

## Design Principles

1. **Interface-based**: All repositories define interfaces for easy mocking and testing
2. **Context-aware**: All methods accept context for cancellation and timeout support
3. **Error handling**: Consistent error handling with GORM error checking
4. **Null safety**: Returns nil for not found cases instead of errors
5. **Atomic operations**: Uses GORM expressions for atomic counter updates
6. **Pagination support**: List methods support limit and offset parameters

## Requirements Validation

This implementation satisfies the following requirements from the design document:

- **Requirements 1.1, 1.2**: User registration and authentication data operations
- **Requirements 3.1, 3.2**: Role and permission management
- **Requirements 4.1, 5.1, 6.1, 6.4**: Content management (posts, comments, votes, counts)
- **Requirements 7.1, 7.2**: Circle/community management
- **Requirements 10.1, 11.1, 11.2**: Notification and messaging operations
- **Requirements 17.1, 17.2**: User profile and statistics management

## Next Steps

The repository layer is now complete. The next steps in the implementation plan are:

1. Implement the Service layer that uses these repositories
2. Implement authentication and authorization middleware
3. Implement Handler/Controller layer for HTTP endpoints
4. Add unit tests for repository operations
5. Add integration tests with test database
