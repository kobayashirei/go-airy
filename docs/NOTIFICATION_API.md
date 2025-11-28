# Notification API Documentation

## Overview

The Notification System provides real-time notification functionality for users. It supports various notification types including comments, votes, mentions, and system announcements.

## Features

- **Pre-rendered Content**: Notifications are created with pre-rendered content for immediate display
- **Unread Priority**: Unread notifications are displayed first in the list
- **Batch Operations**: Mark all notifications as read in a single operation
- **Authorization**: Users can only access their own notifications

## API Endpoints

### 1. Get Notifications

Retrieves a paginated list of notifications for the authenticated user.

**Endpoint**: `GET /api/v1/notifications`

**Authentication**: Required

**Query Parameters**:
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Number of items per page (default: 20, max: 100)

**Response**:
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "notifications": [
      {
        "id": 1,
        "receiver_id": 123,
        "trigger_user_id": 456,
        "type": "comment",
        "entity_type": "post",
        "entity_id": 789,
        "content": "John commented on your post: My First Post",
        "is_read": false,
        "created_at": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20,
    "unread_count": 5
  },
  "request_id": "abc123",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 2. Get Unread Count

Retrieves the count of unread notifications for the authenticated user.

**Endpoint**: `GET /api/v1/notifications/unread-count`

**Authentication**: Required

**Response**:
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "unread_count": 5
  },
  "request_id": "abc123",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 3. Mark Notification as Read

Marks a single notification as read.

**Endpoint**: `PUT /api/v1/notifications/:id/read`

**Authentication**: Required

**URL Parameters**:
- `id`: Notification ID

**Response**:
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "message": "Notification marked as read"
  },
  "request_id": "abc123",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Error Responses**:
- `404 NOT_FOUND`: Notification not found
- `403 FORBIDDEN`: Unauthorized to access this notification

### 4. Mark All Notifications as Read

Marks all notifications for the authenticated user as read.

**Endpoint**: `PUT /api/v1/notifications/read-all`

**Authentication**: Required

**Response**:
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "message": "All notifications marked as read"
  },
  "request_id": "abc123",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Notification Types

The system supports the following notification types:

### 1. Comment Notifications
- **Type**: `comment`
- **Trigger**: When a user comments on a post or replies to a comment
- **Content Format**: "{username} commented on your post: {post_title}"

### 2. Vote Notifications
- **Type**: `vote`
- **Trigger**: When a user upvotes a post or comment
- **Content Format**: "{username} upvoted your post: {post_title}"

### 3. Mention Notifications
- **Type**: `mention`
- **Trigger**: When a user mentions another user in a post or comment
- **Content Format**: "{username} mentioned you in a post: {post_title}"

### 4. System Notifications
- **Type**: `system`
- **Trigger**: System-generated announcements
- **Content Format**: Custom content provided by the system

## Entity Types

Notifications can reference different entity types:

- `post`: References a post
- `comment`: References a comment

## Notification Ordering

Notifications are ordered by:
1. **Read Status**: Unread notifications appear first
2. **Creation Time**: Within each group (read/unread), sorted by newest first

## Implementation Details

### Service Layer

The `NotificationService` provides the following methods:

```go
type NotificationService interface {
    CreateNotification(ctx context.Context, req CreateNotificationRequest) (*models.Notification, error)
    GetNotifications(ctx context.Context, userID int64, page, pageSize int) (*NotificationListResponse, error)
    MarkAsRead(ctx context.Context, userID, notificationID int64) error
    MarkAllAsRead(ctx context.Context, userID int64) error
    GetUnreadCount(ctx context.Context, userID int64) (int64, error)
}
```

### Content Pre-rendering

Notifications are created with pre-rendered content to improve performance:

1. When a notification is created, the system fetches relevant data (user, post, comment)
2. Content is formatted based on the notification type
3. The rendered content is stored in the database
4. No additional queries are needed when displaying notifications

### Authorization

The system enforces strict authorization:

- Users can only view their own notifications
- Attempting to mark another user's notification as read returns a 403 Forbidden error
- All endpoints require authentication via JWT token

## Usage Example

### Creating a Notification (Internal)

```go
// Create a comment notification
notification, err := notificationService.CreateNotification(ctx, service.CreateNotificationRequest{
    ReceiverID:    postAuthorID,
    TriggerUserID: &commenterID,
    Type:          "comment",
    EntityType:    "post",
    EntityID:      &postID,
    Content:       "", // Will be auto-generated
})
```

### Retrieving Notifications (Client)

```bash
# Get first page of notifications
curl -X GET "http://localhost:8080/api/v1/notifications?page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Get unread count
curl -X GET "http://localhost:8080/api/v1/notifications/unread-count" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Mark notification as read
curl -X PUT "http://localhost:8080/api/v1/notifications/123/read" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Mark all as read
curl -X PUT "http://localhost:8080/api/v1/notifications/read-all" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Database Schema

```sql
CREATE TABLE notifications (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    receiver_id BIGINT NOT NULL,
    trigger_user_id BIGINT,
    type VARCHAR(20) NOT NULL,
    entity_type VARCHAR(20),
    entity_id BIGINT,
    content VARCHAR(500) NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    INDEX idx_receiver_id (receiver_id),
    INDEX idx_is_read (is_read),
    INDEX idx_created_at (created_at)
);
```

## Performance Considerations

1. **Indexing**: The `receiver_id`, `is_read`, and `created_at` columns are indexed for efficient queries
2. **Pagination**: Always use pagination to avoid loading too many notifications at once
3. **Caching**: Consider caching unread counts for frequently accessed users
4. **Pre-rendering**: Content is pre-rendered to avoid N+1 query problems

## Future Enhancements

- Real-time notifications via WebSocket
- Push notifications to mobile devices
- Notification preferences and filtering
- Notification grouping (e.g., "John and 5 others liked your post")
- Email digest of notifications
