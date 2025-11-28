# Airy 接口文档总览

- 基础地址：`http://localhost:${SERVER_PORT}`（默认 `8081`）
- 版本前缀：`/api/v1`
- 认证方式：`Authorization: Bearer <JWT>`（登录后获得）
- 响应结构：统一使用标准响应结构

```
{
  "code": "SUCCESS",
  "message": "Success",
  "data": { ... },
  "request_id": "<uuid>",
  "timestamp": "2025-01-01T12:00:00Z"
}
```

错误响应：
```
{
  "code": "BAD_REQUEST",
  "message": "参数错误",
  "details": { ... },
  "request_id": "<uuid>",
  "timestamp": "2025-01-01T12:00:00Z"
}
```

- 分页规范：`page`（默认 1）、`page_size`（默认 20，最大 100）
- 速率限制：默认启用，具体阈值见配置（`RATE_LIMIT_*`）

模块与文档：
- 认证与令牌：`docs/api/auth.md`
- 用户与资料：`docs/api/users.md`
- 圈子：`docs/api/circles.md`
- 私信与会话：`docs/api/messages.md`
- 通知：`docs/api/notifications.md`
- 管理后台：`docs/api/admin.md`
- 系统与健康：`docs/api/system.md`

接口稳定性说明：
- 所有需要数据库的接口在数据库不可用时可能返回 `SERVICE_UNAVAILABLE`
- 在开发模式下，部分受保护接口未启用鉴权中间件，生产环境必须启用
