# Airy API Documentation

## Project Information

- **Project Name**: Airy
- **Version**: v0.1.0
- **Base URL**: `/api/v1`
- **Author**: Rei (kobayashirei)

## Overview

Airy is a modern community platform backend built with Go. This document describes all available API endpoints.

## Authentication

Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <token>
```

## Response Format

### Success Response

```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {},
  "request_id": "uuid",
  "timestamp": "2025-01-01T00:00:00Z"
}
```

### Error Response

```json
{
  "code": "ERROR_CODE",
  "message": "Error message",
  "details": {},
  "request_id": "uuid",
  "timestamp": "2025-01-01T00:00:00Z"
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| SUCCESS | 200 | Request successful |
| BAD_REQUEST | 400 | Invalid request parameters |
| UNAUTHORIZED | 401 | Authentication required or failed |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Resource not found |
| CONFLICT | 409 | Resource conflict (e.g., duplicate) |
| INTERNAL_ERROR | 500 | Server error |

---

## Authentication APIs

### Register User

Create a new user account.

**Endpoint:** `POST /api/v1/auth/register`

**Request Body:**
```json
{
  "username": "string",
  "email": "user@example.com",
  "phone": "13800138000",
  "password": "string"
}
```

**Response:**
```json
{
  "code": "SUCCESS",
  "data": {
    "user_id": 1,
    "username": "string",
    "email": "user@example.com",
    "message": "Activation email sent"
  }
}
```

**Errors:**
- `400` - Invalid email/phone format
- `409` - User already exists

---

### Activate Account

Activate user account with token.

**Endpoint:** `POST /api/v1/auth/activate`

**Request Body:**
```json
{
  "token": "activation-token"
}
```

**Response:**
```json
{
  "code": "SUCCESS",
  "data": {
    "message": "Account activated successfully"
  }
}
```

---

### Login with Password

**Endpoint:** `POST /api/v1/auth/login`

**Request Body:**
```json
{
  "identifier": "email or phone",
  "password": "string"
}
```

**Response:**
```json
{
  "code": "SUCCESS",
  "data": {
    "access_token": "jwt-token",
    "refresh_token": "refresh-token",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "username": "string",
      "email": "user@example.com"
    }
  }
}
```

---

### Login with Verification Code

**Endpoint:** `POST /api/v1/auth/login/code`

**Request Body:**
```json
{
  "identifier": "email or phone",
  "code": "123456"
}
```

---

### Refresh Token

**Endpoint:** `POST /api/v1/auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "refresh-token"
}
```

---

## Post APIs

### Create Post

**Endpoint:** `POST /api/v1/posts`  
**Auth Required:** Yes

**Request Body:**
```json
{
  "title": "Post Title",
  "content": "Markdown content",
  "circle_id": 1,
  "category": "tech",
  "tags": ["go", "backend"],
  "is_anonymous": false,
  "allow_comment": true
}
```

**Response:**
```json
{
  "code": "SUCCESS",
  "data": {
    "id": 1,
    "title": "Post Title",
    "content_markdown": "...",
    "content_html": "...",
    "status": "published",
    "author_id": 1,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

---

### Get Post

**Endpoint:** `GET /api/v1/posts/:id`

**Response:**
```json
{
  "code": "SUCCESS",
  "data": {
    "id": 1,
    "title": "Post Title",
    "content_html": "...",
    "author": {
      "id": 1,
      "username": "author"
    },
    "vote_count": 10,
    "comment_count": 5,
    "view_count": 100
  }
}
```

---

### Update Post

**Endpoint:** `PUT /api/v1/posts/:id`  
**Auth Required:** Yes (owner only)

---

### Delete Post

**Endpoint:** `DELETE /api/v1/posts/:id`  
**Auth Required:** Yes (owner only)

---

### List Posts

**Endpoint:** `GET /api/v1/posts`

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| page | int | Page number (default: 1) |
| page_size | int | Items per page (default: 20) |
| circle_id | int | Filter by circle |
| status | string | Filter by status |
| sort_by | string | Sort field |

---

## Comment APIs

### Create Comment

**Endpoint:** `POST /api/v1/posts/:id/comments`  
**Auth Required:** Yes

**Request Body:**
```json
{
  "content": "Comment content",
  "parent_id": null
}
```

---

### Get Comment Tree

**Endpoint:** `GET /api/v1/posts/:id/comments`

**Response:**
```json
{
  "code": "SUCCESS",
  "data": {
    "post_id": 1,
    "comments": [
      {
        "id": 1,
        "content": "...",
        "author": {},
        "level": 0,
        "children": []
      }
    ]
  }
}
```

---

### Delete Comment

**Endpoint:** `DELETE /api/v1/comments/:id`  
**Auth Required:** Yes (owner only)

---

## Vote APIs

### Vote

**Endpoint:** `POST /api/v1/votes`  
**Auth Required:** Yes

**Request Body:**
```json
{
  "entity_type": "post",
  "entity_id": 1,
  "vote_type": "up"
}
```

| Field | Values |
|-------|--------|
| entity_type | `post`, `comment` |
| vote_type | `up`, `down` |

---

### Cancel Vote

**Endpoint:** `DELETE /api/v1/votes`  
**Auth Required:** Yes

**Request Body:**
```json
{
  "entity_type": "post",
  "entity_id": 1
}
```

---

### Get Vote

**Endpoint:** `GET /api/v1/votes/:entity_type/:entity_id`  
**Auth Required:** Yes

---

## Circle APIs

### Create Circle

**Endpoint:** `POST /api/v1/circles`  
**Auth Required:** Yes

**Request Body:**
```json
{
  "name": "Circle Name",
  "description": "Description",
  "status": "public",
  "join_rule": "free"
}
```

| Field | Values |
|-------|--------|
| status | `public`, `semi_public`, `private` |
| join_rule | `free`, `approval` |

---

### Get Circle

**Endpoint:** `GET /api/v1/circles/:id`

---

### Join Circle

**Endpoint:** `POST /api/v1/circles/:id/join`  
**Auth Required:** Yes

---

### Approve Member

**Endpoint:** `POST /api/v1/circles/:id/members/:userId/approve`  
**Auth Required:** Yes (moderator/creator)

---

### Assign Moderator

**Endpoint:** `POST /api/v1/circles/:id/moderators`  
**Auth Required:** Yes (creator only)

**Request Body:**
```json
{
  "user_id": 1
}
```

---

### Get Circle Members

**Endpoint:** `GET /api/v1/circles/:id/members`

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| role | string | Filter by role (member/moderator/pending) |

---

## Feed APIs

### Get User Feed

**Endpoint:** `GET /api/v1/feed`  
**Auth Required:** Yes

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| limit | int | 20 | Max 100 |
| offset | int | 0 | Pagination offset |
| sort_by | string | created_at | `created_at` or `hotness_score` |

---

### Get Circle Feed

**Endpoint:** `GET /api/v1/circles/:id/feed`

**Query Parameters:** Same as User Feed

---

## Search APIs

### Search Posts

**Endpoint:** `GET /api/v1/search/posts`

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| keyword | string | Search keyword |
| circle_id | int | Filter by circle |
| tags | string | Comma-separated tags |
| sort_by | string | `time`, `hotness`, `relevance` |
| page | int | Page number |
| page_size | int | Items per page |

---

### Search Users

**Endpoint:** `GET /api/v1/search/users`

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| keyword | string | Search keyword |
| page | int | Page number |
| page_size | int | Items per page |

---

## Notification APIs

### Get Notifications

**Endpoint:** `GET /api/v1/notifications`  
**Auth Required:** Yes

**Query Parameters:**
| Parameter | Type | Default |
|-----------|------|---------|
| page | int | 1 |
| page_size | int | 20 |

---

### Get Unread Count

**Endpoint:** `GET /api/v1/notifications/unread-count`  
**Auth Required:** Yes

---

### Mark as Read

**Endpoint:** `PUT /api/v1/notifications/:id/read`  
**Auth Required:** Yes

---

### Mark All as Read

**Endpoint:** `PUT /api/v1/notifications/read-all`  
**Auth Required:** Yes

---

## Message APIs

### Get Conversations

**Endpoint:** `GET /api/v1/conversations`  
**Auth Required:** Yes

---

### Create Conversation

**Endpoint:** `POST /api/v1/conversations`  
**Auth Required:** Yes

**Request Body:**
```json
{
  "other_user_id": 1
}
```

---

### Get Messages

**Endpoint:** `GET /api/v1/conversations/:id/messages`  
**Auth Required:** Yes

---

### Send Message

**Endpoint:** `POST /api/v1/conversations/:id/messages`  
**Auth Required:** Yes

**Request Body:**
```json
{
  "content_type": "text",
  "content": "Message content"
}
```

---

## User Profile APIs

### Get Profile

**Endpoint:** `GET /api/v1/users/:id/profile`

---

### Update Profile

**Endpoint:** `PUT /api/v1/users/profile`  
**Auth Required:** Yes

**Request Body:**
```json
{
  "avatar": "url",
  "bio": "Bio text",
  "gender": "male",
  "birthday": "2000-01-01"
}
```

---

### Get User Posts

**Endpoint:** `GET /api/v1/users/:id/posts`

**Query Parameters:**
| Parameter | Type | Default |
|-----------|------|---------|
| status | string | published |
| sort_by | string | created_at |
| order | string | desc |
| page | int | 1 |
| limit | int | 20 |

---

## Admin APIs

All admin endpoints require authentication and admin permissions.

### Get Dashboard

**Endpoint:** `GET /api/v1/admin/dashboard`

---

### List Users

**Endpoint:** `GET /api/v1/admin/users`

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| status | string | Filter by status |
| keyword | string | Search keyword |
| sort_by | string | Sort field |
| order | string | asc/desc |
| page | int | Page number |
| page_size | int | Items per page |

---

### Ban User

**Endpoint:** `POST /api/v1/admin/users/:id/ban`

**Request Body:**
```json
{
  "reason": "Ban reason"
}
```

---

### Unban User

**Endpoint:** `POST /api/v1/admin/users/:id/unban`

---

### List Posts (Admin)

**Endpoint:** `GET /api/v1/admin/posts`

---

### Batch Review Posts

**Endpoint:** `POST /api/v1/admin/posts/batch-review`

**Request Body:**
```json
{
  "post_ids": [1, 2, 3],
  "action": "approve",
  "reason": "Optional reason"
}
```

---

### List Admin Logs

**Endpoint:** `GET /api/v1/admin/logs`

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| operator_id | int | Filter by operator |
| action | string | Filter by action |
| entity_type | string | Filter by entity type |
| start_date | string | Start date |
| end_date | string | End date |

---

## Health Check

### Metrics

**Endpoint:** `GET /metrics`

Prometheus metrics endpoint for monitoring.
