# 用户与资料

前缀：`/api/v1/users`

## 获取用户资料
- `GET /:id/profile`
- 响应示例：
```
{
  "code": "SUCCESS",
  "data": {
    "user": { "id": 1, "username": "kawaii" },
    "profile": { "points": 120, "level": 3 },
    "stats": { "posts": 10, "comments": 32 }
  }
}
```

## 获取用户文章列表
- `GET /:id/posts?page=1&page_size=20`
- 响应包含文章数组与分页信息

## 更新我的资料（需登录）
- `PUT /profile`
- 请求头：`Authorization: Bearer <JWT>`
- 请求体：
```
{
  "bio": "关于我...",
  "avatar": "https://..."
}
```
- 响应：`SUCCESS`
