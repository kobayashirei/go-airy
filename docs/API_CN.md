# Airy API 文档

## 项目信息

- **项目名称**: Airy
- **版本**: v0.1.0
- **基础 URL**: `/api/v1`
- **作者**: Rei (kobayashirei)

## 概述

Airy 是一个基于 Go 语言构建的现代化社区平台后端。本文档描述所有可用的 API 端点。

## 认证

大多数端点需要 JWT 认证。在 Authorization 头中包含令牌：

```
Authorization: Bearer <token>
```

## 响应格式

### 成功响应

```json
{
  "code": "SUCCESS",
  "message": "成功",
  "data": {},
  "request_id": "uuid",
  "timestamp": "2025-01-01T00:00:00Z"
}
```

### 错误响应

```json
{
  "code": "ERROR_CODE",
  "message": "错误消息",
  "details": {},
  "request_id": "uuid",
  "timestamp": "2025-01-01T00:00:00Z"
}
```

## 错误码

| 错误码 | HTTP 状态码 | 描述 |
|--------|-------------|------|
| SUCCESS | 200 | 请求成功 |
| BAD_REQUEST | 400 | 无效的请求参数 |
| UNAUTHORIZED | 401 | 需要认证或认证失败 |
| FORBIDDEN | 403 | 权限不足 |
| NOT_FOUND | 404 | 资源未找到 |
| CONFLICT | 409 | 资源冲突（如重复） |
| INTERNAL_ERROR | 500 | 服务器错误 |

---

## 认证 API

### 用户注册

创建新用户账号。

**端点:** `POST /api/v1/auth/register`

**请求体:**
```json
{
  "username": "用户名",
  "email": "user@example.com",
  "phone": "13800138000",
  "password": "密码"
}
```

**响应:**
```json
{
  "code": "SUCCESS",
  "data": {
    "user_id": 1,
    "username": "用户名",
    "email": "user@example.com",
    "message": "激活邮件已发送"
  }
}
```

**错误:**
- `400` - 无效的邮箱/手机号格式
- `409` - 用户已存在

---

### 激活账号

使用令牌激活用户账号。

**端点:** `POST /api/v1/auth/activate`

**请求体:**
```json
{
  "token": "激活令牌"
}
```

**响应:**
```json
{
  "code": "SUCCESS",
  "data": {
    "message": "账号激活成功"
  }
}
```

---

### 密码登录

**端点:** `POST /api/v1/auth/login`

**请求体:**
```json
{
  "identifier": "邮箱或手机号",
  "password": "密码"
}
```

**响应:**
```json
{
  "code": "SUCCESS",
  "data": {
    "access_token": "jwt-token",
    "refresh_token": "refresh-token",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "username": "用户名",
      "email": "user@example.com"
    }
  }
}
```

---

### 验证码登录

**端点:** `POST /api/v1/auth/login/code`

**请求体:**
```json
{
  "identifier": "邮箱或手机号",
  "code": "123456"
}
```

---

### 刷新令牌

**端点:** `POST /api/v1/auth/refresh`

**请求体:**
```json
{
  "refresh_token": "refresh-token"
}
```

---

## 帖子 API

### 创建帖子

**端点:** `POST /api/v1/posts`  
**需要认证:** 是

**请求体:**
```json
{
  "title": "帖子标题",
  "content": "Markdown 内容",
  "circle_id": 1,
  "category": "tech",
  "tags": ["go", "backend"],
  "is_anonymous": false,
  "allow_comment": true
}
```

**响应:**
```json
{
  "code": "SUCCESS",
  "data": {
    "id": 1,
    "title": "帖子标题",
    "content_markdown": "...",
    "content_html": "...",
    "status": "published",
    "author_id": 1,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

---

### 获取帖子

**端点:** `GET /api/v1/posts/:id`

**响应:**
```json
{
  "code": "SUCCESS",
  "data": {
    "id": 1,
    "title": "帖子标题",
    "content_html": "...",
    "author": {
      "id": 1,
      "username": "作者"
    },
    "vote_count": 10,
    "comment_count": 5,
    "view_count": 100
  }
}
```

---

### 更新帖子

**端点:** `PUT /api/v1/posts/:id`  
**需要认证:** 是（仅作者）

---

### 删除帖子

**端点:** `DELETE /api/v1/posts/:id`  
**需要认证:** 是（仅作者）

---

### 帖子列表

**端点:** `GET /api/v1/posts`

**查询参数:**
| 参数 | 类型 | 描述 |
|------|------|------|
| page | int | 页码（默认: 1） |
| page_size | int | 每页数量（默认: 20） |
| circle_id | int | 按圈子筛选 |
| status | string | 按状态筛选 |
| sort_by | string | 排序字段 |

---

## 评论 API

### 创建评论

**端点:** `POST /api/v1/posts/:id/comments`  
**需要认证:** 是

**请求体:**
```json
{
  "content": "评论内容",
  "parent_id": null
}
```

---

### 获取评论树

**端点:** `GET /api/v1/posts/:id/comments`

**响应:**
```json
{
  "code": "SUCCESS",
  "data": {
    "post_id": 1,
    "comments": [
      {
        "id": 1,
        "content": "...",
        "author": {},
        "level": 0,
        "children": []
      }
    ]
  }
}
```

---

### 删除评论

**端点:** `DELETE /api/v1/comments/:id`  
**需要认证:** 是（仅作者）

---

## 投票 API

### 投票

**端点:** `POST /api/v1/votes`  
**需要认证:** 是

**请求体:**
```json
{
  "entity_type": "post",
  "entity_id": 1,
  "vote_type": "up"
}
```

| 字段 | 可选值 |
|------|--------|
| entity_type | `post`, `comment` |
| vote_type | `up`, `down` |

---

### 取消投票

**端点:** `DELETE /api/v1/votes`  
**需要认证:** 是

**请求体:**
```json
{
  "entity_type": "post",
  "entity_id": 1
}
```

---

### 获取投票

**端点:** `GET /api/v1/votes/:entity_type/:entity_id`  
**需要认证:** 是

---

## 圈子 API

### 创建圈子

**端点:** `POST /api/v1/circles`  
**需要认证:** 是

**请求体:**
```json
{
  "name": "圈子名称",
  "description": "描述",
  "status": "public",
  "join_rule": "free"
}
```

| 字段 | 可选值 |
|------|--------|
| status | `public`（公开）, `semi_public`（半公开）, `private`（私密） |
| join_rule | `free`（自由加入）, `approval`（需要审批） |

---

### 获取圈子

**端点:** `GET /api/v1/circles/:id`

---

### 加入圈子

**端点:** `POST /api/v1/circles/:id/join`  
**需要认证:** 是

---

### 审批成员

**端点:** `POST /api/v1/circles/:id/members/:userId/approve`  
**需要认证:** 是（版主/创建者）

---

### 分配版主

**端点:** `POST /api/v1/circles/:id/moderators`  
**需要认证:** 是（仅创建者）

**请求体:**
```json
{
  "user_id": 1
}
```

---

### 获取圈子成员

**端点:** `GET /api/v1/circles/:id/members`

**查询参数:**
| 参数 | 类型 | 描述 |
|------|------|------|
| role | string | 按角色筛选（member/moderator/pending） |

---

## Feed API

### 获取用户 Feed

**端点:** `GET /api/v1/feed`  
**需要认证:** 是

**查询参数:**
| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| limit | int | 20 | 最大 100 |
| offset | int | 0 | 分页偏移 |
| sort_by | string | created_at | `created_at` 或 `hotness_score` |

---

### 获取圈子 Feed

**端点:** `GET /api/v1/circles/:id/feed`

**查询参数:** 同用户 Feed

---

## 搜索 API

### 搜索帖子

**端点:** `GET /api/v1/search/posts`

**查询参数:**
| 参数 | 类型 | 描述 |
|------|------|------|
| keyword | string | 搜索关键词 |
| circle_id | int | 按圈子筛选 |
| tags | string | 逗号分隔的标签 |
| sort_by | string | `time`, `hotness`, `relevance` |
| page | int | 页码 |
| page_size | int | 每页数量 |

---

### 搜索用户

**端点:** `GET /api/v1/search/users`

**查询参数:**
| 参数 | 类型 | 描述 |
|------|------|------|
| keyword | string | 搜索关键词 |
| page | int | 页码 |
| page_size | int | 每页数量 |

---

## 通知 API

### 获取通知

**端点:** `GET /api/v1/notifications`  
**需要认证:** 是

**查询参数:**
| 参数 | 类型 | 默认值 |
|------|------|--------|
| page | int | 1 |
| page_size | int | 20 |

---

### 获取未读数量

**端点:** `GET /api/v1/notifications/unread-count`  
**需要认证:** 是

---

### 标记已读

**端点:** `PUT /api/v1/notifications/:id/read`  
**需要认证:** 是

---

### 全部标记已读

**端点:** `PUT /api/v1/notifications/read-all`  
**需要认证:** 是

---

## 私信 API

### 获取会话列表

**端点:** `GET /api/v1/conversations`  
**需要认证:** 是

---

### 创建会话

**端点:** `POST /api/v1/conversations`  
**需要认证:** 是

**请求体:**
```json
{
  "other_user_id": 1
}
```

---

### 获取消息

**端点:** `GET /api/v1/conversations/:id/messages`  
**需要认证:** 是

---

### 发送消息

**端点:** `POST /api/v1/conversations/:id/messages`  
**需要认证:** 是

**请求体:**
```json
{
  "content_type": "text",
  "content": "消息内容"
}
```

---

## 用户档案 API

### 获取档案

**端点:** `GET /api/v1/users/:id/profile`

---

### 更新档案

**端点:** `PUT /api/v1/users/profile`  
**需要认证:** 是

**请求体:**
```json
{
  "avatar": "url",
  "bio": "个人简介",
  "gender": "male",
  "birthday": "2000-01-01"
}
```

---

### 获取用户帖子

**端点:** `GET /api/v1/users/:id/posts`

**查询参数:**
| 参数 | 类型 | 默认值 |
|------|------|--------|
| status | string | published |
| sort_by | string | created_at |
| order | string | desc |
| page | int | 1 |
| limit | int | 20 |

---

## 管理后台 API

所有管理端点需要认证和管理员权限。

### 获取仪表盘

**端点:** `GET /api/v1/admin/dashboard`

---

### 用户列表

**端点:** `GET /api/v1/admin/users`

**查询参数:**
| 参数 | 类型 | 描述 |
|------|------|------|
| status | string | 按状态筛选 |
| keyword | string | 搜索关键词 |
| sort_by | string | 排序字段 |
| order | string | asc/desc |
| page | int | 页码 |
| page_size | int | 每页数量 |

---

### 封禁用户

**端点:** `POST /api/v1/admin/users/:id/ban`

**请求体:**
```json
{
  "reason": "封禁原因"
}
```

---

### 解封用户

**端点:** `POST /api/v1/admin/users/:id/unban`

---

### 帖子列表（管理）

**端点:** `GET /api/v1/admin/posts`

---

### 批量审核帖子

**端点:** `POST /api/v1/admin/posts/batch-review`

**请求体:**
```json
{
  "post_ids": [1, 2, 3],
  "action": "approve",
  "reason": "可选原因"
}
```

---

### 管理日志列表

**端点:** `GET /api/v1/admin/logs`

**查询参数:**
| 参数 | 类型 | 描述 |
|------|------|------|
| operator_id | int | 按操作者筛选 |
| action | string | 按操作筛选 |
| entity_type | string | 按实体类型筛选 |
| start_date | string | 开始日期 |
| end_date | string | 结束日期 |

---

## 健康检查

### 监控指标

**端点:** `GET /metrics`

Prometheus 监控指标端点。
