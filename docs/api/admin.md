# 管理后台

前缀：`/api/v1/admin`

所有接口需登录并具备管理员权限。

## 仪表盘
- `GET /dashboard`

## 用户管理
- `GET /users`（分页）
- `POST /users/:id/ban`
- `POST /users/:id/unban`

## 内容管理
- `GET /posts`（分页）
- `POST /posts/batch-review`（批量审核）

## 审计日志
- `GET /logs`（分页）
