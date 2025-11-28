# 认证与授权实现

本文档描述 Airy 平台的认证和授权系统实现。

## 概述

系统实现了基于 JWT 的认证和 RBAC（基于角色的访问控制）授权，符合设计文档的规范。

## 已实现组件

### 1. JWT 服务 (`internal/auth/jwt.go`)

提供 JWT 令牌操作：
- **GenerateToken**: 创建包含用户 ID 和角色的新 JWT 令牌
- **ParseToken**: 验证和解析 JWT 令牌
- **RefreshToken**: 刷新现有令牌

**功能特性:**
- 使用 HS256 签名算法
- 可配置的令牌过期时间
- 对过期和无效令牌的正确错误处理

### 2. 认证中间件 (`internal/middleware/auth.go`)

提供请求认证：
- **AuthMiddleware**: 需要有效的 JWT 令牌，失败时中止请求
- **OptionalAuthMiddleware**: 如果存在令牌则提取用户信息，否则继续
- **辅助函数**: GetUserID, GetRoles 用于访问用户上下文

**功能特性:**
- 从 Authorization 头提取令牌（Bearer 方案）
- 验证令牌并将用户信息注入 Gin 上下文
- 返回适当的 HTTP 状态码（401 表示认证失败）

### 3. 权限服务 (`internal/service/permission_service.go`)

实现权限检查逻辑：
- **CheckPermission**: 检查角色是否具有特定权限
- **CheckPermissionWithCircle**: 检查带圈子上下文的权限
- **GetUserPermissions**: 获取用户的所有权限

**功能特性:**
- 聚合多个角色的权限
- 支持圈子特定角色
- 当没有圈子特定角色时回退到全局角色

### 4. RBAC 中间件 (`internal/middleware/rbac.go`)

提供授权执行：
- **RequirePermission**: 需要特定权限
- **RequireCirclePermission**: 需要特定圈子中的权限
- **RequireAnyPermission**: 需要多个权限中的至少一个
- **RequireAllPermissions**: 需要所有指定的权限

**功能特性:**
- 与 PermissionService 集成
- 权限不足时返回 403 Forbidden
- 支持灵活的权限检查策略

## 使用示例

参见 `examples/auth_example.go` 获取完整的工作示例。

### 基本认证

```go
router := gin.Default()
jwtService := auth.NewJWTService("secret", 24*time.Hour)

// 受保护的路由
router.Use(middleware.AuthMiddleware(jwtService))
router.GET("/profile", profileHandler)
```

### 基于权限的授权

```go
// 需要特定权限
router.POST("/posts",
    middleware.AuthMiddleware(jwtService),
    middleware.RequirePermission(permissionService, "post:create"),
    createPostHandler,
)

// 圈子特定权限
router.DELETE("/circles/:circleId/posts/:id",
    middleware.AuthMiddleware(jwtService),
    middleware.RequireCirclePermission(permissionService, "post:delete", "circleId"),
    deletePostHandler,
)
```

## 需求验证

此实现满足以下需求：

### 需求 2.3（JWT 令牌生成）
✅ JWT 令牌包含用户 ID 和角色
✅ 令牌使用 HS256 算法签名
✅ 令牌具有可配置的过期时间

### 需求 18.1（令牌解析）
✅ 从 Authorization 头解析令牌
✅ 强制使用 Bearer 方案
✅ 无效令牌返回 401 Unauthorized

### 需求 18.2（令牌验证）
✅ 验证令牌签名
✅ 检查令牌过期
✅ 提取用户 ID 和角色并注入上下文

### 需求 3.2（权限验证）
✅ 通过用户角色验证权限
✅ 权限检查返回布尔结果
✅ 正确处理错误

### 需求 3.3（权限聚合）
✅ 聚合多个角色的权限
✅ 计算所有权限的并集
✅ 消除重复权限

### 需求 3.5（圈子特定权限）
✅ 支持圈子特定角色
✅ 在权限检查中考虑圈子上下文
✅ 适当时回退到全局角色

### 需求 18.3（受保护端点）
✅ 中间件强制执行权限要求
✅ 权限不足返回 403
✅ 缺少/无效认证返回 401

## 设计属性

实现解决了设计文档中的以下正确性属性：

- **属性 6**: JWT 令牌完整性 - 令牌包含用户 ID 和完整角色列表
- **属性 9**: 权限验证 - 具有所需权限的用户成功
- **属性 10**: 权限聚合 - 聚合多个角色的权限
- **属性 11**: 圈子特定权限 - 正确处理圈子上下文
- **属性 42**: JWT 令牌解析 - 正确提取用户 ID 和角色
- **属性 43**: 授权错误处理 - 返回正确的错误码

## 安全考虑

1. **密钥**: 生产环境必须使用强密钥并安全保存
2. **令牌过期**: 默认配置为 24 小时，生产环境应更短
3. **HTTPS**: 所有令牌传输应通过 HTTPS
4. **令牌存储**: 客户端应安全存储令牌
5. **刷新机制**: 已实现用于延长会话

## 测试

实现：
- ✅ 编译无错误
- ✅ 通过所有现有测试
- ✅ 与现有中间件和服务集成
- ⏭️ 属性测试标记为可选（任务 5.2, 5.4, 5.5）

## 后续步骤

在生产环境使用此系统：

1. 在环境变量中配置 JWT 密钥
2. 设置适当的令牌过期时间
3. 在数据库中初始化角色和权限
4. 在注册/入职时为用户分配角色
5. 将认证和授权中间件应用到路由
6. 实现令牌刷新端点用于延长会话

## 创建的文件

- `internal/auth/jwt.go` - JWT 服务实现
- `internal/auth/README.md` - 认证文档
- `internal/middleware/auth.go` - 认证中间件
- `internal/middleware/rbac.go` - 授权中间件
- `internal/service/permission_service.go` - 权限服务
- `examples/auth_example.go` - 使用示例
- `docs/AUTH_IMPLEMENTATION.md` - 本文档
