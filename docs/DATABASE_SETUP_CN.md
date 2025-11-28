# 数据库层设置

本文档描述 Airy 项目的数据库层实现。

## 概述

数据库层已完整实现，包含以下组件：

1. **数据库连接管理** - 基于 GORM 的连接池和健康检查
2. **数据模型** - 覆盖所有系统实体的 18 个完整模型
3. **数据库迁移** - 用于模式管理的 SQL 迁移文件

## 组件

### 1. 数据库连接 (`internal/database/database.go`)

功能特性：
- GORM 集成 MySQL 驱动
- 连接池配置（最大空闲/打开连接数、连接生命周期）
- 自定义 GORM 日志与 zap 集成
- 健康检查功能
- 优雅关闭连接

主要函数：
- `Init(cfg *config.DatabaseConfig)` - 初始化数据库连接
- `Close()` - 关闭数据库连接
- `HealthCheck(ctx context.Context)` - 检查数据库健康状态
- `GetDB()` - 获取数据库实例

### 2. 数据模型 (`internal/models/`)

所有模型按类别组织：

#### 用户模型 (`user.go`)
- `User` - 用户账号和认证信息
- `UserProfile` - 用户档案数据（积分、等级、粉丝）
- `UserStats` - 用户统计（帖子、评论、投票）

#### 权限模型 (`permission.go`)
- `Role` - 用户角色（管理员、版主、用户）
- `Permission` - 系统权限
- `RolePermission` - 角色权限关联
- `UserRole` - 用户角色关联（可选圈子范围）

#### 内容模型 (`content.go`)
- `Post` - 用户帖子（Markdown/HTML 内容）
- `Comment` - 层级评论（带路径追踪）
- `Vote` - 帖子/评论投票
- `Favorite` - 用户收藏
- `EntityCount` - 实体聚合计数

#### 圈子模型 (`circle.go`)
- `Circle` - 社区圈子/群组
- `CircleMember` - 圈子成员

#### 通知模型 (`notification.go`)
- `Notification` - 用户通知
- `Conversation` - 私信会话
- `Message` - 会话消息

#### 管理模型 (`admin.go`)
- `AdminLog` - 管理操作审计日志

### 3. 数据库迁移 (`migrations/`)

六个迁移文件覆盖所有表：

1. **000001_create_users_tables** - User, UserProfile, UserStats
2. **000002_create_permission_tables** - Role, Permission, RolePermission, UserRole
3. **000003_create_circle_tables** - Circle, CircleMember
4. **000004_create_content_tables** - Post, Comment, Vote, Favorite, EntityCount
5. **000005_create_notification_tables** - Notification, Conversation, Message
6. **000006_create_admin_tables** - AdminLog

每个迁移包含：
- 升级迁移（创建表）
- 降级迁移（删除表）
- 性能索引
- 外键约束保证引用完整性
- 默认值和约束

### 4. 迁移管理 (`internal/database/migrate.go`)

程序化迁移管理函数：
- `RunMigrations()` - 运行所有待处理迁移
- `RollbackMigration()` - 回滚上一次迁移
- `MigrationVersion()` - 获取当前迁移版本

## 使用方法

### 初始化数据库连接

```go
import (
    "github.com/kobayashirei/airy/internal/config"
    "github.com/kobayashirei/airy/internal/database"
)

// 加载配置
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// 初始化数据库
if err := database.Init(&cfg.Database); err != nil {
    log.Fatal(err)
}
defer database.Close()
```

### 运行迁移

使用 Go API：
```go
err := database.RunMigrations(&cfg.Database, "migrations")
if err != nil {
    log.Fatal(err)
}
```

使用 Makefile：
```bash
# 设置环境变量
export DB_USER=root
export DB_PASSWORD=password
export DB_HOST=localhost
export DB_PORT=3306
export DB_NAME=airygithub

# 运行迁移
make migrate-up

# 回滚上一次迁移
make migrate-down

# 检查迁移版本
make migrate-version

# 创建新迁移
make migrate-create NAME=add_new_feature
```

### 使用模型

```go
import (
    "github.com/kobayashirei/airy/internal/database"
    "github.com/kobayashirei/airy/internal/models"
)

// 创建用户
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

// 查询用户
var foundUser models.User
db.Where("email = ?", "test@example.com").First(&foundUser)
```

### 健康检查

```go
import (
    "context"
    "time"
)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := database.HealthCheck(ctx); err != nil {
    log.Printf("数据库健康检查失败: %v", err)
}
```

## 配置

数据库配置通过环境变量管理：

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

## 测试

所有组件包含完整测试：

```bash
# 运行所有测试
go test -v ./internal/database/... ./internal/models/...

# 使用 short 标志（跳过集成测试）
go test -v -short ./internal/database/... ./internal/models/...

# 运行并生成覆盖率报告
go test -v -coverprofile=coverage.out ./internal/database/... ./internal/models/...
go tool cover -html=coverage.out
```

## 数据库模式

完整的数据库模式包括：

- **18 个表** 覆盖所有系统实体
- **外键约束** 保证引用完整性
- **索引** 在频繁查询的列上
- **唯一约束** 保证数据完整性
- **默认值** 用于可选字段
- **时间戳** 自动更新

## 性能考虑

1. **连接池**: 配置最大空闲/打开连接数
2. **预处理语句**: 在 GORM 中启用查询缓存
3. **索引**: 在所有外键和频繁查询的列上创建
4. **独立计数表**: `entity_counts` 避免锁竞争
5. **跳过默认事务**: 禁用以提高性能

## 需求验证

此实现满足以下需求：

- **需求 19.1**: 数据写入的事务原子性
- **需求 19.2**: 通过外键保证引用完整性
- **需求 19.3**: 高频更新使用独立计数表
- **需求 19.4**: 级联删除和软删除支持
- **需求 19.5**: 错误时事务回滚

## 后续步骤

数据库层完成后，您可以：

1. 实现 Repository 层（任务 4）
2. 创建 Service 层业务逻辑
3. 构建 API 处理器
4. 添加认证和授权
5. 实现缓存层

## 参考资料

- [GORM 文档](https://gorm.io/zh_CN/docs/)
- [golang-migrate 文档](https://github.com/golang-migrate/migrate)
- [MySQL 文档](https://dev.mysql.com/doc/)
