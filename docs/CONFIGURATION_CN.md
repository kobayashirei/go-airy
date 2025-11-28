# Airy 配置指南

## 概述

Airy 使用环境变量进行配置。将 `.env.example` 复制为 `.env` 并根据您的环境修改配置值。

```bash
cp .env.example .env
```

## 配置分类

### 服务器配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `SERVER_HOST` | `0.0.0.0` | 服务器绑定地址 |
| `SERVER_PORT` | `8080` | 服务器端口 |
| `GIN_MODE` | `debug` | Gin 模式: `debug`, `release`, `test` |

### 数据库配置 (MySQL)

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `DB_HOST` | `localhost` | 数据库主机 |
| `DB_PORT` | `3306` | 数据库端口 |
| `DB_NAME` | `airygithub` | 数据库名称 |
| `DB_USER` | `root` | 数据库用户 |
| `DB_PASSWORD` | - | 数据库密码 |
| `DB_MAX_IDLE_CONNS` | `10` | 最大空闲连接数 |
| `DB_MAX_OPEN_CONNS` | `100` | 最大打开连接数 |
| `DB_CONN_MAX_LIFETIME` | `3600` | 连接最大生命周期（秒） |

### Redis 配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `REDIS_HOST` | `localhost` | Redis 主机 |
| `REDIS_PORT` | `6379` | Redis 端口 |
| `REDIS_PASSWORD` | - | Redis 密码 |
| `REDIS_DB` | `0` | Redis 数据库编号 |

### Elasticsearch 配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `ES_HOST` | `localhost` | Elasticsearch 主机 |
| `ES_PORT` | `9200` | Elasticsearch 端口 |

### JWT 配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `JWT_SECRET` | - | **必填。** JWT 签名密钥 |
| `JWT_EXPIRATION` | `86400` | 令牌过期时间（秒） |

生成安全密钥：
```bash
openssl rand -hex 32
```

### 消息队列 (RabbitMQ)

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `MQ_HOST` | `localhost` | RabbitMQ 主机 |
| `MQ_PORT` | `5672` | RabbitMQ 端口 |
| `MQ_USER` | `guest` | RabbitMQ 用户 |
| `MQ_PASSWORD` | `guest` | RabbitMQ 密码 |

### 日志配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `LOG_LEVEL` | `info` | 日志级别: `debug`, `info`, `warn`, `error` |
| `LOG_OUTPUT` | `stdout` | 输出: `stdout`, `file` |
| `LOG_FILE_PATH` | `logs/app.log` | 日志文件路径 |

### 缓存配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `CACHE_DEFAULT_EXPIRATION` | `3600` | 默认 TTL（秒） |
| `CACHE_CLEANUP_INTERVAL` | `600` | 清理间隔（秒） |
| `CACHE_WARMUP_ENABLED` | `true` | 启用缓存预热 |
| `CACHE_WARMUP_HOT_POSTS` | `100` | 预加载热门帖子数量 |
| `CACHE_WARMUP_HOT_USERS` | `50` | 预加载热门用户数量 |
| `CACHE_WARMUP_HOT_CIRCLES` | `20` | 预加载热门圈子数量 |
| `CACHE_WARMUP_REFRESH_INTERVAL` | `30` | 刷新间隔（分钟） |
| `CACHE_WARMUP_CONCURRENCY` | `10` | 预热并发数 |

### Feed 配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `FEED_FANOUT_THRESHOLD` | `1000` | 写扩散策略的粉丝数阈值 |

### 热度配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `HOTNESS_ALGORITHM` | `reddit` | 算法: `reddit`, `hackernews` |

### 限流配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `RATE_LIMIT_ENABLED` | `true` | 启用限流 |
| `RATE_LIMIT_REQUESTS_PER_SECOND` | `100` | 每秒请求数 |
| `RATE_LIMIT_BURST_SIZE` | `200` | 突发大小 |

### TLS/HTTPS 配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `TLS_ENABLED` | `false` | 启用 TLS |
| `TLS_CERT_FILE` | - | 证书文件路径 |
| `TLS_KEY_FILE` | - | 密钥文件路径 |

### 安全配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `ENCRYPTION_KEY` | - | AES 加密密钥（base64） |
| `BCRYPT_COST` | `10` | Bcrypt 成本因子（4-31） |

生成加密密钥：
```bash
openssl rand -base64 32
```

### CSRF 配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `CSRF_TOKEN_LENGTH` | `32` | 令牌长度（字节） |
| `CSRF_TOKEN_EXPIRATION` | `86400` | 令牌过期时间（秒） |
| `CSRF_COOKIE_NAME` | `_csrf` | Cookie 名称 |
| `CSRF_HEADER_NAME` | `X-CSRF-Token` | 头部名称 |
| `CSRF_FORM_FIELD_NAME` | `_csrf` | 表单字段名称 |
| `CSRF_SECURE` | `false` | 安全 Cookie 标志 |
| `CSRF_SAME_SITE` | `Strict` | SameSite 属性 |

### CORS 配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000,http://localhost:8080` | 允许的源（逗号分隔） |
| `CORS_ALLOWED_METHODS` | `GET,POST,PUT,DELETE,OPTIONS,PATCH` | 允许的方法 |
| `CORS_ALLOWED_HEADERS` | `Origin,Content-Type,Accept,Authorization,X-CSRF-Token` | 允许的头部 |
| `CORS_EXPOSED_HEADERS` | `Content-Length,Content-Type` | 暴露的头部 |
| `CORS_ALLOW_CREDENTIALS` | `true` | 允许凭证 |
| `CORS_MAX_AGE` | `86400` | 预检缓存最大时间（秒） |

### 应用 URL

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `APP_BASE_URL` | `http://localhost:8080` | 应用基础 URL |
| `FRONTEND_URL` | `http://localhost:3000` | 前端 URL |

### 令牌过期时间

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `ACTIVATION_TOKEN_EXPIRATION` | `86400` | 激活令牌 TTL（秒） |
| `PASSWORD_RESET_TOKEN_EXPIRATION` | `3600` | 密码重置令牌 TTL（秒） |
| `VERIFICATION_CODE_EXPIRATION` | `300` | 验证码 TTL（秒） |

### 邮件配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `SMTP_HOST` | - | SMTP 服务器主机 |
| `SMTP_PORT` | `587` | SMTP 端口 |
| `SMTP_USER` | - | SMTP 用户名 |
| `SMTP_PASSWORD` | - | SMTP 密码 |
| `SMTP_FROM` | - | 发件人邮箱地址 |

### 内容审核

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `MODERATION_API_URL` | - | 审核 API URL |
| `MODERATION_API_KEY` | - | 审核 API 密钥 |

### 协程池

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `GOROUTINE_POOL_SIZE` | `10000` | 协程池大小 |

## 环境示例

### 开发环境

```bash
GIN_MODE=debug
LOG_LEVEL=debug
DB_HOST=localhost
REDIS_HOST=localhost
JWT_SECRET=dev-secret-key-change-in-production
```

### 生产环境

```bash
GIN_MODE=release
LOG_LEVEL=info
LOG_OUTPUT=file
TLS_ENABLED=true
RATE_LIMIT_ENABLED=true
CACHE_WARMUP_ENABLED=true

# 安全设置
CSRF_SECURE=true
CSRF_SAME_SITE=Strict
CORS_ALLOWED_ORIGINS=https://yourdomain.com
CORS_ALLOW_CREDENTIALS=true

# 应用 URL
APP_BASE_URL=https://api.yourdomain.com
FRONTEND_URL=https://yourdomain.com

# 使用强密钥！
JWT_SECRET=<生成的密钥>
ENCRYPTION_KEY=<生成的密钥>
```

## Docker Compose

示例 `docker-compose.yml`：

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    env_file:
      - .env
    depends_on:
      - mysql
      - redis
      - elasticsearch
      - rabbitmq

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: airygithub
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

  elasticsearch:
    image: elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    volumes:
      - es_data:/usr/share/elasticsearch/data

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "15672:15672"

volumes:
  mysql_data:
  redis_data:
  es_data:
```

## 安全最佳实践

### 密钥管理

1. **永远不要将密钥提交到版本控制**
   - 本地使用 `.env` 文件（添加到 `.gitignore`）
   - 生产环境使用环境变量或密钥管理器

2. **生成强密钥**
   ```bash
   # JWT 密钥（64 个十六进制字符）
   openssl rand -hex 32
   
   # 加密密钥（base64，32 字节）
   openssl rand -base64 32
   ```

3. **定期轮换密钥**
   - JWT 密钥应定期轮换
   - 加密密钥轮换时需要谨慎迁移

### 生产环境检查清单

- [ ] `GIN_MODE=release`
- [ ] `TLS_ENABLED=true` 并配置有效证书
- [ ] `JWT_SECRET` 是强唯一值
- [ ] `ENCRYPTION_KEY` 已设置用于敏感数据加密
- [ ] `CSRF_SECURE=true`（需要 HTTPS）
- [ ] `CORS_ALLOWED_ORIGINS` 限制为您的域名
- [ ] `RATE_LIMIT_ENABLED=true`
- [ ] 数据库凭证不是默认值
- [ ] 如果 Redis 暴露则设置密码
- [ ] 日志级别为 `info` 或更高（不是 `debug`）

### 环境特定设置

| 设置 | 开发环境 | 生产环境 |
|------|----------|----------|
| `GIN_MODE` | `debug` | `release` |
| `LOG_LEVEL` | `debug` | `info` |
| `TLS_ENABLED` | `false` | `true` |
| `CSRF_SECURE` | `false` | `true` |
| `RATE_LIMIT_ENABLED` | `false` | `true` |
| `CORS_ALLOWED_ORIGINS` | `*` 或 localhost | 特定域名 |

## 快速开始

1. 复制环境文件：
   ```bash
   cp .env.example .env
   ```

2. 设置必需值：
   ```bash
   # 生成 JWT 密钥
   JWT_SECRET=$(openssl rand -hex 32)
   
   # 设置数据库密码
   DB_PASSWORD=your-password
   ```

3. 运行迁移：
   ```bash
   make migrate-up
   ```

4. 启动服务器：
   ```bash
   make run
   ```

## 故障排除

### 常见问题

#### JWT 密钥错误
```
Error: JWT secret must be set and changed from default
```
**解决方案**: 设置唯一的 `JWT_SECRET` 值：
```bash
export JWT_SECRET=$(openssl rand -hex 32)
```

#### 数据库连接失败
```
Error: dial tcp: connect: connection refused
```
**解决方案**: 
- 验证 `DB_HOST` 和 `DB_PORT` 是否正确
- 确保 MySQL/PostgreSQL 正在运行
- 检查防火墙规则

#### Redis 连接失败
```
Error: dial tcp: connect: connection refused
```
**解决方案**:
- 验证 `REDIS_HOST` 和 `REDIS_PORT` 是否正确
- 确保 Redis 正在运行
- 如果启用了认证，检查 `REDIS_PASSWORD`

#### TLS 证书错误
```
Error: TLS certificate file is required when TLS is enabled
```
**解决方案**:
- 设置 `TLS_CERT_FILE` 和 `TLS_KEY_FILE` 路径
- 或在开发环境禁用 TLS：`TLS_ENABLED=false`

#### CORS 错误
```
Access-Control-Allow-Origin header missing
```
**解决方案**:
- 将前端 URL 添加到 `CORS_ALLOWED_ORIGINS`
- 如果使用 Cookie，确保 `CORS_ALLOW_CREDENTIALS=true`

#### 限流问题
```
Error: Too many requests
```
**解决方案**:
- 增加 `RATE_LIMIT_REQUESTS_PER_SECOND` 和 `RATE_LIMIT_BURST_SIZE`
- 或在开发环境禁用：`RATE_LIMIT_ENABLED=false`

### 调试技巧

1. **启用调试日志**：
   ```bash
   LOG_LEVEL=debug
   GIN_MODE=debug
   ```

2. **检查配置加载**：
   应用在启动时会记录配置值（敏感值会被掩码）。

3. **验证环境文件**：
   ```bash
   # 检查语法错误
   source .env && echo "环境文件有效"
   ```

4. **测试数据库连接**：
   ```bash
   mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "SELECT 1"
   ```

5. **测试 Redis 连接**：
   ```bash
   redis-cli -h $REDIS_HOST -p $REDIS_PORT ping
   ```

## 配置参考

有关所有配置选项及其默认值的完整列表，请参阅项目根目录中的 `.env.example` 文件。

### 必需 vs 可选

| 类别 | 必需 | 可选 |
|------|------|------|
| 服务器 | `SERVER_PORT` | `SERVER_HOST`, `GIN_MODE` |
| 数据库 | `DB_NAME`, `DB_USER` | `DB_PASSWORD`, 连接池设置 |
| JWT | `JWT_SECRET` | `JWT_EXPIRATION` |
| Redis | - | 全部（使用默认值） |
| Elasticsearch | - | 全部（使用默认值） |
| 安全 | - | `ENCRYPTION_KEY`, `BCRYPT_COST` |
| TLS | `TLS_CERT_FILE`, `TLS_KEY_FILE`（如果启用） | `TLS_ENABLED` |
