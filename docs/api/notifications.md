# 通知

前缀：`/api/v1/notifications`

所有接口需登录：`Authorization: Bearer <JWT>`

## 获取我的通知
- `GET /`
- 支持分页

## 获取未读数量
- `GET /unread-count`

## 标记单条为已读
- `PUT /:id/read`

## 标记全部为已读
- `PUT /read-all`
