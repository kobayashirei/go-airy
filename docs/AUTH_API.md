# Authentication API Documentation

## Overview

This document describes the authentication endpoints implemented in the Airy backend system.

## Endpoints

### 1. User Registration

**Endpoint:** `POST /api/v1/auth/register`

**Description:** Register a new user account.

**Request Body:**
```json
{
  "username": "testuser",
  "email": "user@example.com",
  "phone": "+1234567890",
  "password": "securepassword123"
}
```

**Notes:**
- Either `email` or `phone` is required (or both)
- Password must be at least 6 characters
- Username must be 3-50 characters

**Success Response (200):**
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "user_id": 1,
    "message": "Registration successful. Please check your email to activate your account."
  },
  "request_id": "...",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid email/phone format, or validation error
- `409 Conflict`: User already exists
- `500 Internal Server Error`: Server error

---

### 2. Account Activation

**Endpoint:** `POST /api/v1/auth/activate`

**Description:** Activate a user account using the activation token sent via email.

**Request Body:**
```json
{
  "token": "activation-token-from-email"
}
```

**Alternative:** Token can also be passed as a query parameter:
```
POST /api/v1/auth/activate?token=activation-token-from-email
```

**Success Response (200):**
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "message": "Account activated successfully"
  },
  "request_id": "...",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid or expired activation token
- `404 Not Found`: User not found
- `500 Internal Server Error`: Server error

---

### 3. Login with Password

**Endpoint:** `POST /api/v1/auth/login`

**Description:** Authenticate a user with email/phone/username and password.

**Request Body:**
```json
{
  "identifier": "user@example.com",
  "password": "securepassword123"
}
```

**Notes:**
- `identifier` can be email, phone number, or username
- System automatically detects the identifier type

**Success Response (200):**
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "token": "jwt-access-token",
    "refresh_token": "jwt-refresh-token",
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "user@example.com",
      "phone": "+1234567890",
      "avatar": "",
      "gender": "",
      "birthday": "0001-01-01T00:00:00Z",
      "bio": "",
      "status": "active",
      "last_login_at": "2024-01-01T00:00:00Z",
      "last_login_ip": "192.168.1.1",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  },
  "request_id": "...",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `401 Unauthorized`: Invalid credentials
- `403 Forbidden`: Account is not active
- `500 Internal Server Error`: Server error

---

### 4. Login with Verification Code

**Endpoint:** `POST /api/v1/auth/login/code`

**Description:** Authenticate a user with email/phone and verification code.

**Request Body:**
```json
{
  "identifier": "user@example.com",
  "code": "123456"
}
```

**Notes:**
- `identifier` must be email or phone number
- Verification code must be obtained separately (not implemented in this task)

**Success Response (200):**
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "token": "jwt-access-token",
    "refresh_token": "jwt-refresh-token",
    "user": { ... }
  },
  "request_id": "...",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid or expired verification code, or invalid identifier format
- `403 Forbidden`: Account is not active
- `404 Not Found`: User not found
- `500 Internal Server Error`: Server error

---

### 5. Refresh Token

**Endpoint:** `POST /api/v1/auth/refresh`

**Description:** Refresh an access token using a refresh token.

**Request Body:**
```json
{
  "refresh_token": "jwt-refresh-token"
}
```

**Alternative:** Token can also be passed in Authorization header:
```
Authorization: Bearer jwt-refresh-token
```

**Success Response (200):**
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "token": "new-jwt-access-token"
  },
  "request_id": "...",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `401 Unauthorized`: Invalid or expired refresh token
- `500 Internal Server Error`: Server error

---

## Authentication Flow

### Registration Flow
1. User submits registration form
2. System validates email/phone format
3. System checks for duplicate users
4. Password is hashed using bcrypt
5. User record is created with status "inactive"
6. Activation token is generated and stored in Redis (24h expiration)
7. Activation email is sent to user
8. User clicks activation link
9. System validates token and updates user status to "active"

### Login Flow
1. User submits login credentials
2. System finds user by identifier (email/phone/username)
3. System verifies password using bcrypt
4. System checks user status is "active"
5. JWT tokens are generated (access + refresh)
6. Login time and IP are recorded
7. Tokens and user info are returned

### Token Refresh Flow
1. Client sends refresh token
2. System validates refresh token
3. New access token is generated with same claims
4. New token is returned

---

## Security Features

- **Password Hashing**: bcrypt with default cost factor (10)
- **JWT Tokens**: HS256 signing algorithm
- **Token Expiration**: Configurable via JWT_EXPIRATION env variable
- **Activation Tokens**: 24-hour expiration, stored in Redis
- **Login Logging**: IP address and timestamp recorded
- **Error Messages**: Generic messages to prevent user enumeration

---

## Environment Variables

- `JWT_SECRET`: Secret key for JWT signing (required)
- `JWT_EXPIRATION`: Token expiration in seconds (default: 86400 = 24 hours)
- `REDIS_HOST`: Redis host for token storage
- `REDIS_PORT`: Redis port
- `REDIS_PASSWORD`: Redis password (optional)

---

## Testing

Run the handler tests:
```bash
go test ./internal/handler -v
```

Run all tests:
```bash
go test ./... -v
```
