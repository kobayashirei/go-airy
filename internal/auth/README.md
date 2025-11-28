# Authentication and Authorization

This package provides JWT-based authentication and RBAC (Role-Based Access Control) authorization for the Airy platform.

## Components

### JWT Service

The JWT service handles token generation, parsing, and refresh operations.

```go
import (
    "github.com/kobayashirei/airy/internal/auth"
    "time"
)

// Create JWT service
jwtService := auth.NewJWTService("your-secret-key", 24*time.Hour)

// Generate token
token, err := jwtService.GenerateToken(userID, []string{"user", "moderator"})

// Parse token
claims, err := jwtService.ParseToken(token)

// Refresh token
newToken, err := jwtService.RefreshToken(oldToken)
```

### Authentication Middleware

The authentication middleware validates JWT tokens and injects user information into the request context.

```go
import (
    "github.com/kobayashirei/airy/internal/middleware"
    "github.com/gin-gonic/gin"
)

router := gin.Default()

// Require authentication for all routes
router.Use(middleware.AuthMiddleware(jwtService))

// Optional authentication (doesn't abort if no token)
router.Use(middleware.OptionalAuthMiddleware(jwtService))

// Get user info in handlers
router.GET("/profile", func(c *gin.Context) {
    userID, _ := middleware.GetUserID(c)
    roles, _ := middleware.GetRoles(c)
    // ...
})
```

### Permission Service

The permission service checks user permissions based on their roles.

```go
import (
    "github.com/kobayashirei/airy/internal/service"
)

permissionService := service.NewPermissionService(
    permissionRepo,
    roleRepo,
    userRoleRepo,
)

// Check if roles have permission
hasPermission, err := permissionService.CheckPermission(ctx, roles, "post:create")

// Check permission with circle context
hasPermission, err := permissionService.CheckPermissionWithCircle(ctx, userID, circleID, "post:delete")

// Get all user permissions
permissions, err := permissionService.GetUserPermissions(ctx, userID, circleID)
```

### RBAC Middleware

The RBAC middleware enforces permission requirements on routes.

```go
import (
    "github.com/kobayashirei/airy/internal/middleware"
)

// Require specific permission
router.POST("/posts", 
    middleware.AuthMiddleware(jwtService),
    middleware.RequirePermission(permissionService, "post:create"),
    postHandler.Create,
)

// Require circle-specific permission
router.DELETE("/circles/:circleId/posts/:id",
    middleware.AuthMiddleware(jwtService),
    middleware.RequireCirclePermission(permissionService, "post:delete", "circleId"),
    postHandler.Delete,
)

// Require any of multiple permissions
router.GET("/admin/dashboard",
    middleware.AuthMiddleware(jwtService),
    middleware.RequireAnyPermission(permissionService, "admin:view", "moderator:view"),
    adminHandler.Dashboard,
)

// Require all permissions
router.POST("/admin/users/ban",
    middleware.AuthMiddleware(jwtService),
    middleware.RequireAllPermissions(permissionService, "admin:users", "admin:ban"),
    adminHandler.BanUser,
)
```

## Permission Naming Convention

Permissions follow the format: `resource:action`

Examples:
- `post:create` - Create posts
- `post:edit` - Edit posts
- `post:delete` - Delete posts
- `user:ban` - Ban users
- `circle:manage` - Manage circles
- `admin:view` - View admin dashboard

## Circle-Specific Permissions

Users can have different roles in different circles. The system supports:

1. **Global roles**: Apply across the entire platform (e.g., super_admin)
2. **Circle-specific roles**: Apply only within a specific circle (e.g., moderator of Circle A)

When checking permissions with a circle context, the system:
1. First checks circle-specific roles
2. Falls back to global roles if no circle-specific roles exist

## Error Handling

The middleware returns appropriate HTTP status codes:

- `401 Unauthorized`: Missing or invalid token
- `403 Forbidden`: Valid token but insufficient permissions

## Security Considerations

1. **JWT Secret**: Use a strong, random secret key in production
2. **Token Expiration**: Set appropriate expiration times (e.g., 15 minutes for access tokens)
3. **Token Refresh**: Implement refresh token mechanism for extended sessions
4. **HTTPS**: Always use HTTPS in production to protect tokens in transit
5. **Token Storage**: Store tokens securely on the client side (e.g., httpOnly cookies)
