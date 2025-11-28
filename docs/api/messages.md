# 私信与会话

前缀：`/api/v1/conversations`

所有接口需登录：`Authorization: Bearer <JWT>`

## 获取会话列表
- `GET /`
- 支持分页参数

## 创建会话
- `POST /`
- 请求体：
```
{
  "target_user_id": 123
}
```
- 响应：返回会话信息

## 获取会话消息
- `GET /:id/messages?page=1&page_size=20`

## 发送消息
- `POST /:id/messages`
- 请求体：
```
{
  "content": "你好呀~"
}
```
- 响应：消息对象
