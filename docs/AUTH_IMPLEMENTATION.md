# Authentication and Authorization Implementation

This document describes the implementation of the authentication and authorization system for the Airy platform.

## Overview

The system implements JWT-based authentication and RBAC (Role-Based Access Control) authorization as specified in the design document.

## Implemented Components

### 1. JWT Service (`internal/auth/jwt.go`)

Provides JWT token operations:
- **GenerateToken**: Creates a new JWT token with user ID and roles
- **ParseToken**: Validates and parses a JWT token
- **RefreshToken**: Refreshes an existing token

**Features:**
- Uses HS256 signing algorithm
- Configurable token expiration
- Proper error handling for expired and invalid tokens

### 2. Authentication Middleware (`internal/middleware/auth.go`)

Provides request authentication:
- **AuthMiddleware**: Requires valid JWT token, aborts on failure
- **OptionalAuthMiddleware**: Extracts user info if token present, continues otherwise
- **Helper functions**: GetUserID, GetRoles for accessing user context

**Features:**
- Extracts token from Authorization header (Bearer scheme)
- Validates token and injects user info into Gin context
- Returns appropriate HTTP status codes (401 for auth failures)

### 3. Permission Service (`internal/service/permission_service.go`)

Implements permission checking logic:
- **CheckPermission**: Checks if roles have a specific permission
- **CheckPermissionWithCircle**: Checks permissions with circle context
- **GetUserPermissions**: Retrieves all permissions for a user

**Features:**
- Aggregates permissions from multiple roles
- Supports circle-specific roles
- Falls back to global roles when no circle-specific roles exist

### 4. RBAC Middleware (`internal/middleware/rbac.go`)

Provides authorization enforcement:
- **RequirePermission**: Requires a specific permission
- **RequireCirclePermission**: Requires permission in a specific circle
- **RequireAnyPermission**: Requires at least one of multiple permissions
- **RequireAllPermissions**: Requires all specified permissions

**Features:**
- Integrates with PermissionService
- Returns 403 Forbidden for insufficient permissions
- Supports flexible permission checking strategies

## Usage Examples

See `examples/auth_example.go` for a complete working example.

### Basic Authentication

```go
router := gin.Default()
jwtService := auth.NewJWTService("secret", 24*time.Hour)

// Protected route
router.Use(middleware.AuthMiddleware(jwtService))
router.GET("/profile", profileHandler)
```

### Permission-Based Authorization

```go
// Require specific permission
router.POST("/posts",
    middleware.AuthMiddleware(jwtService),
    middleware.RequirePermission(permissionService, "post:create"),
    createPostHandler,
)

// Circle-specific permission
router.DELETE("/circles/:circleId/posts/:id",
    middleware.AuthMiddleware(jwtService),
    middleware.RequireCirclePermission(permissionService, "post:delete", "circleId"),
    deletePostHandler,
)
```

## Requirements Validation

This implementation satisfies the following requirements:

### Requirement 2.3 (JWT Token Generation)
✅ JWT tokens are generated with user ID and roles
✅ Tokens are signed using HS256 algorithm
✅ Tokens have configurable expiration

### Requirement 18.1 (Token Parsing)
✅ Tokens are parsed from Authorization header
✅ Bearer scheme is enforced
✅ Invalid tokens return 401 Unauthorized

### Requirement 18.2 (Token Validation)
✅ Token signature is validated
✅ Token expiration is checked
✅ User ID and roles are extracted and injected into context

### Requirement 3.2 (Permission Verification)
✅ User permissions are verified through their roles
✅ Permission checks return boolean results
✅ Errors are properly handled

### Requirement 3.3 (Permission Aggregation)
✅ Permissions from multiple roles are aggregated
✅ Union of all permissions is computed
✅ Duplicate permissions are eliminated

### Requirement 3.5 (Circle-Specific Permissions)
✅ Circle-specific roles are supported
✅ Circle context is considered in permission checks
✅ Falls back to global roles when appropriate

### Requirement 18.3 (Protected Endpoints)
✅ Middleware enforces permission requirements
✅ Returns 403 for insufficient permissions
✅ Returns 401 for missing/invalid authentication

## Design Properties Addressed

The implementation addresses the following correctness properties from the design document:

- **Property 6**: JWT token completeness - Tokens contain user ID and complete role list
- **Property 9**: Permission verification - Users with required permissions succeed
- **Property 10**: Permission aggregation - Multiple roles' permissions are aggregated
- **Property 11**: Circle-specific permissions - Circle context is properly handled
- **Property 42**: JWT token parsing - User ID and roles are extracted correctly
- **Property 43**: Authorization error handling - Proper error codes are returned

## Security Considerations

1. **Secret Key**: Must be strong and kept secure in production
2. **Token Expiration**: Configured to 24 hours by default, should be shorter in production
3. **HTTPS**: All token transmission should be over HTTPS
4. **Token Storage**: Clients should store tokens securely
5. **Refresh Mechanism**: Implemented for extended sessions

## Testing

The implementation:
- ✅ Compiles without errors
- ✅ Passes all existing tests
- ✅ Integrates with existing middleware and services
- ⏭️ Property-based tests are marked as optional (tasks 5.2, 5.4, 5.5)

## Next Steps

To use this system in production:

1. Configure JWT secret in environment variables
2. Set appropriate token expiration times
3. Seed initial roles and permissions in database
4. Assign roles to users during registration/onboarding
5. Apply authentication and authorization middleware to routes
6. Implement token refresh endpoint for extended sessions

## Files Created

- `internal/auth/jwt.go` - JWT service implementation
- `internal/auth/README.md` - Authentication documentation
- `internal/middleware/auth.go` - Authentication middleware
- `internal/middleware/rbac.go` - Authorization middleware
- `internal/service/permission_service.go` - Permission service
- `examples/auth_example.go` - Usage examples
- `docs/AUTH_IMPLEMENTATION.md` - This document
