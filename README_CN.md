# Airy 后端

基于 Go 语言构建的高性能社区平台后端系统。

## 项目信息

- **项目名称**: Airy
- **版本**: v0.1.0
- **作者**: Rei
- **GitHub**: [kobayashirei](https://github.com/kobayashirei)
- **网站**: [iqwq.com](https://iqwq.com)
- **仓库**: github.com/kobayashirei/airy

## 功能特性

- 基于 Gin 框架的 RESTful API
- 使用 Zap 的结构化日志
- 使用 Viper 的配置管理
- 基于环境变量的配置
- 请求日志和错误处理中间件
- 优雅关闭
- 健康检查端点
- JWT 认证与 RBAC 权限系统
- Redis 缓存与 Cache-Aside 模式
- Elasticsearch 全文搜索
- RabbitMQ 消息队列
- ants 协程池异步任务处理
- 内容审核系统
- Feed 流推送（支持写扩散和读扩散）
- 热度排序算法（Reddit/Hacker News 风格）

## 项目结构

```
.
├── cmd/
│   └── server/              # 应用程序入口
│       └── main.go
├── internal/
│   ├── auth/                # JWT 认证
│   ├── cache/               # Redis 缓存服务
│   ├── config/              # 配置管理
│   ├── database/            # 数据库连接和迁移
│   ├── handler/             # HTTP 处理器
│   ├── logger/              # 日志工具
│   ├── middleware/          # HTTP 中间件
│   ├── models/              # 数据模型
│   ├── mq/                  # 消息队列
│   ├── repository/          # 数据访问层
│   ├── response/            # 响应工具
│   ├── router/              # 路由配置
│   ├── search/              # Elasticsearch 搜索
│   ├── security/            # 安全工具
│   ├── service/             # 业务逻辑层
│   ├── taskpool/            # 协程池
│   └── version/             # 版本信息
├── migrations/              # 数据库迁移文件
├── docs/                    # 文档
├── examples/                # 示例代码
├── .env.example             # 环境变量示例
├── .gitignore
├── go.mod
└── README.md
```

## 快速开始

### 前置要求

- Go 1.21 或更高版本
- MySQL 8.0+ 或 PostgreSQL
- Redis 7.0+
- Elasticsearch 8.x
- RabbitMQ 3.x

### 安装步骤

1. 克隆仓库：
```bash
git clone https://github.com/kobayashirei/airy.git
cd airy
```

2. 复制 `.env.example` 到 `.env` 并配置环境变量：
```bash
cp .env.example .env
```

3. 安装依赖：
```bash
go mod download
```

4. 运行数据库迁移：
```bash
# 使用 golang-migrate
migrate -path migrations -database "mysql://user:password@tcp(localhost:3306)/airygithub" up
```

5. 启动服务器：
```bash
go run cmd/server/main.go
```

服务器默认在 `http://localhost:8080` 启动。

### API 端点

#### 健康检查
- `GET /health` - 健康检查端点
- `GET /metrics` - Prometheus 监控指标

#### 认证
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/activate` - 账号激活
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/refresh` - 刷新令牌

#### 帖子
- `POST /api/v1/posts` - 创建帖子
- `GET /api/v1/posts/:id` - 获取帖子
- `PUT /api/v1/posts/:id` - 更新帖子
- `DELETE /api/v1/posts/:id` - 删除帖子
- `GET /api/v1/posts` - 帖子列表

#### 评论
- `POST /api/v1/posts/:id/comments` - 创建评论
- `GET /api/v1/posts/:id/comments` - 获取评论树
- `DELETE /api/v1/comments/:id` - 删除评论

#### 投票
- `POST /api/v1/votes` - 投票
- `DELETE /api/v1/votes` - 取消投票

#### 圈子
- `POST /api/v1/circles` - 创建圈子
- `GET /api/v1/circles/:id` - 获取圈子
- `POST /api/v1/circles/:id/join` - 加入圈子

#### Feed
- `GET /api/v1/feed` - 获取个人 Feed
- `GET /api/v1/circles/:id/feed` - 获取圈子 Feed

#### 搜索
- `GET /api/v1/search/posts` - 搜索帖子
- `GET /api/v1/search/users` - 搜索用户

#### 通知
- `GET /api/v1/notifications` - 获取通知列表
- `PUT /api/v1/notifications/:id/read` - 标记已读

#### 私信
- `GET /api/v1/conversations` - 获取会话列表
- `POST /api/v1/conversations/:id/messages` - 发送消息

## 配置说明

配置通过环境变量管理。查看 `.env.example` 了解所有可用选项。

主要配置项：
- 服务器设置（主机、端口、模式）
- 数据库连接
- Redis 连接
- JWT 设置
- 日志配置
- 协程池大小
- 缓存配置
- Feed 配置
- 热度算法配置

详细配置说明请参阅 [配置指南](docs/CONFIGURATION_CN.md)。

## 开发指南

### 开发模式运行

```bash
GIN_MODE=debug go run cmd/server/main.go
```

### 生产环境构建

```bash
go build -o bin/server cmd/server/main.go
```

### 运行测试

```bash
# 运行所有测试
go test ./... -v

# 运行特定包的测试
go test ./internal/service/... -v

# 运行测试并生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 日志系统

应用使用 Zap 进行结构化日志记录。所有 HTTP 请求都会记录：
- 请求 ID（UUID）
- HTTP 方法和路径
- 响应状态码
- 延迟时间
- 客户端 IP
- User Agent
- 用户 ID（如果已认证）

错误日志包含：
- 错误消息
- 堆栈跟踪
- 请求上下文

## 架构设计

### 分层架构

```
┌─────────────────────────────────────────┐
│              Handler 层                  │
│         (HTTP 请求处理)                  │
├─────────────────────────────────────────┤
│              Service 层                  │
│         (业务逻辑处理)                   │
├─────────────────────────────────────────┤
│            Repository 层                 │
│         (数据访问层)                     │
├─────────────────────────────────────────┤
│              数据库/缓存                 │
│      (MySQL/Redis/Elasticsearch)        │
└─────────────────────────────────────────┘
```

### 异步处理

- 使用 ants 协程池处理计算密集型任务
- 使用 RabbitMQ 消息队列处理 IO 密集型任务
- 支持搜索索引同步、通知推送、Feed 更新等异步操作

### 缓存策略

- 采用 Cache-Aside 模式
- 支持缓存预热
- 支持缓存失效和刷新

## 文档

- [API 文档](docs/API_CN.md)
- [配置指南](docs/CONFIGURATION_CN.md)
- [数据库设置](docs/DATABASE_SETUP_CN.md)
- [认证实现](docs/AUTH_IMPLEMENTATION_CN.md)
- [异步系统](docs/ASYNC_SYSTEM_CN.md)
- [热度系统](docs/HOTNESS_SYSTEM_CN.md)
- [监控指南](docs/MONITORING_CN.md)

## 贡献指南

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m '添加某个功能'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

MIT

## 联系方式

- 作者: Rei
- GitHub: [@kobayashirei](https://github.com/kobayashirei)
- 网站: [iqwq.com](https://iqwq.com)
