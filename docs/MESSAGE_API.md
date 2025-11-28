# Private Messaging API Documentation

## Overview

The private messaging system allows users to send direct messages to each other through conversations. Each conversation is unique between two users and maintains a history of messages.

## Endpoints

### 1. Get Conversations

Retrieves all conversations for the authenticated user, sorted by last message time.

**Endpoint:** `GET /api/v1/conversations`

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Number of conversations per page (default: 20, max: 100)

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "conversations": [
      {
        "id": 1,
        "user1_id": 1,
        "user2_id": 2,
        "last_message_at": "2024-01-15T10:30:00Z",
        "created_at": "2024-01-10T08:00:00Z",
        "other_user_id": 2,
        "other_user": {
          "id": 2,
          "username": "john_doe",
          "avatar": "https://example.com/avatar.jpg"
        },
        "last_message": {
          "id": 100,
          "conversation_id": 1,
          "sender_id": 2,
          "content_type": "text",
          "content": "Hello!",
          "is_read": false,
          "created_at": "2024-01-15T10:30:00Z"
        },
        "unread_count": 3
      }
    ],
    "total": 10,
    "page": 1,
    "page_size": 20
  }
}
```

### 2. Create or Get Conversation

Creates a new conversation with another user or retrieves an existing one.

**Endpoint:** `POST /api/v1/conversations`

**Request Body:**
```json
{
  "other_user_id": 2
}
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "user1_id": 1,
    "user2_id": 2,
    "last_message_at": "2024-01-15T10:30:00Z",
    "created_at": "2024-01-10T08:00:00Z"
  }
}
```

**Error Responses:**
- `400 Bad Request`: Cannot send message to yourself
- `404 Not Found`: User not found

### 3. Get Messages

Retrieves messages in a conversation with pagination. Automatically marks all messages in the conversation as read for the authenticated user.

**Endpoint:** `GET /api/v1/conversations/:id/messages`

**Path Parameters:**
- `id`: Conversation ID

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Number of messages per page (default: 50, max: 100)

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "messages": [
      {
        "id": 100,
        "conversation_id": 1,
        "sender_id": 2,
        "content_type": "text",
        "content": "Hello!",
        "is_read": true,
        "created_at": "2024-01-15T10:30:00Z"
      },
      {
        "id": 99,
        "conversation_id": 1,
        "sender_id": 1,
        "content_type": "text",
        "content": "Hi there!",
        "is_read": true,
        "created_at": "2024-01-15T10:25:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 50
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid conversation ID
- `403 Forbidden`: Unauthorized to access this conversation
- `404 Not Found`: Conversation not found

### 4. Send Message

Sends a message in a conversation.

**Endpoint:** `POST /api/v1/conversations/:id/messages`

**Path Parameters:**
- `id`: Conversation ID

**Request Body:**
```json
{
  "content_type": "text",
  "content": "Hello, how are you?"
}
```

**Fields:**
- `content_type` (optional): Type of content - "text" or "image" (default: "text")
- `content` (required): Message content

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 101,
    "conversation_id": 1,
    "sender_id": 1,
    "content_type": "text",
    "content": "Hello, how are you?",
    "is_read": false,
    "created_at": "2024-01-15T10:35:00Z"
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid conversation ID or empty message content
- `403 Forbidden`: Unauthorized to send message in this conversation
- `404 Not Found`: Conversation not found

## Features

### Conversation Uniqueness

The system ensures that there is only one conversation between any two users. When creating a conversation, if one already exists between the two users, the existing conversation is returned.

### Message Read Status

When a user opens a conversation (GET messages endpoint), all unread messages in that conversation are automatically marked as read for that user.

### Conversation Sorting

Conversations are sorted by the timestamp of the last message, with the most recent conversations appearing first.

### Authorization

Users can only:
- View conversations they are part of
- Send messages in conversations they are part of
- View messages in conversations they are part of

## Implementation Details

### Service Layer

The `MessageService` implements the following business logic:
- **GetOrCreateConversation**: Ensures conversation uniqueness between two users
- **SendMessage**: Validates sender authorization and creates messages
- **GetConversations**: Retrieves conversations with enriched details (other user info, last message, unread count)
- **GetMessages**: Retrieves paginated messages with authorization checks
- **MarkConversationAsRead**: Marks all messages in a conversation as read for the user

### Repository Layer

The repositories handle data persistence:
- **ConversationRepository**: CRUD operations for conversations
- **MessageRepository**: CRUD operations for messages, including read status management

### Data Models

**Conversation:**
- `id`: Unique identifier
- `user1_id`: First user ID (smaller ID)
- `user2_id`: Second user ID (larger ID)
- `last_message_at`: Timestamp of last message
- `created_at`: Conversation creation timestamp

**Message:**
- `id`: Unique identifier
- `conversation_id`: Reference to conversation
- `sender_id`: User who sent the message
- `content_type`: Type of content (text, image)
- `content`: Message content
- `is_read`: Read status
- `created_at`: Message creation timestamp

## Usage Example

### Starting a Conversation

1. Create or get a conversation with another user:
```bash
curl -X POST http://localhost:8080/api/v1/conversations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"other_user_id": 2}'
```

2. Send a message:
```bash
curl -X POST http://localhost:8080/api/v1/conversations/1/messages \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello!"}'
```

3. Get messages:
```bash
curl -X GET http://localhost:8080/api/v1/conversations/1/messages \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

4. List all conversations:
```bash
curl -X GET http://localhost:8080/api/v1/conversations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Requirements Validation

This implementation satisfies the following requirements from the specification:

- **Requirement 11.1**: Conversation creation/retrieval with uniqueness guarantee
- **Requirement 11.2**: Message sending with content storage
- **Requirement 11.3**: Message type support (text, image)
- **Requirement 11.4**: Conversation list sorted by last message time
- **Requirement 11.5**: Message read status marking when conversation is opened
