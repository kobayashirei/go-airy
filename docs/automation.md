# 自动化开发与初始化

## 一键本地环境
- 需求：Docker Desktop
- 启动：
```
# 启动 MySQL/Redis
powershell ./scripts/dev-up.ps1

# 运行后端
go run cmd/server/main.go
```
- 停止：
```
powershell ./scripts/dev-down.ps1
```

## 配置说明
- `.env.local` 会在首次运行 `dev-up.ps1` 时生成，包含本地开发所需的变量：
  - 数据库：`ENABLE_DATABASE=true`、`ENABLE_DB_AUTO_CREATE=true`、`ENABLE_DB_AUTO_MIGRATE=true`、`ALLOW_START_WITHOUT_DB=false`
  - MySQL：`DB_HOST=127.0.0.1`、`DB_PORT=3306`、`DB_NAME=airy`、`DB_USER=airy`、`DB_PASSWORD=airy`
  - Redis：`REDIS_HOST=127.0.0.1`、`REDIS_PORT=6379`、`REDIS_PASSWORD=airyredis`
  - JWT：自动生成安全随机 `JWT_SECRET`
- 程序优先加载 `.env.local`，不存在时读取 `.env`

## 初始化与迁移
- 首次启动时会：确保库存在 → 初始化连接 → 运行 `migrations/*.sql` → 写入锁文件
- 锁文件：
  - 建库锁：`DB_INIT_LOCK_FILE`（默认 `./data/db_init.lock`）
  - 迁移锁：`DB_SQL_INIT_LOCK_FILE`（默认 `./data/db_sql_init.lock`）
- 脏版本处理：检测到迁移 dirty 时自动强制到当前版本并重试

## 常见问题
- 端口占用：修改 `.env.local` 中 `SERVER_PORT`，或关闭占用端口的进程
- 远程库握手失败：使用本地 Docker 环境，或设置 `DB_TLS_MODE=skip-verify` 并确认远端白名单与权限
- 需要重跑迁移：删除 `./data/db_sql_init.lock` 后重启
