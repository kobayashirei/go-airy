# 认证与令牌

前缀：`/api/v1/auth`

## 注册
- `POST /register`
- 请求体：
```
{
  "email": "user@example.com",
  "username": "kawaii",
  "password": "******"
}
```
- 响应：`SUCCESS`，返回用户基本信息或激活指引

## 激活
- `POST /activate`
- 传参方式（二选一）：
  - 查询参数：`/activate?token=<激活令牌>`
  - 请求体：
    ```
    { "token": "<激活令牌>" }
    ```
- 响应：`SUCCESS`，`{"message":"Account activated successfully"}`

## 登录（密码）
- `POST /login`
- 请求体：
```
{
  "identifier": "user@example.com | +86130xxxx | username",
  "password": "******"
}
```
- 响应：
```
{
  "code": "SUCCESS",
  "data": {
    "token": "<jwt>",
    "refresh_token": "<jwt_refresh>",
    "user": { ... }
  }
}
```

## 登录（验证码）
- `POST /login/code`
- 请求体：
```
{
  "identifier": "user@example.com | +86130xxxx",
  "code": "123456"
}
```
- 响应：同上，返回 `token`/`refresh_token` 与 `user`

## 刷新令牌
- `POST /refresh`
- 请求头或请求体：
  - `Authorization: Bearer <旧refresh_token>`
  - 或：`{"refresh_token":"<旧refresh_token>"}`
- 响应：返回新的 `token`

## 重新发送激活邮件
- `POST /resend-activation`
- 请求体：
```
{ "identifier": "user@example.com | +86130xxxx | username" }
```
- 响应：`SUCCESS`，`{"message":"Activation email resent"}`
- 说明：用户未激活且已绑定邮箱时生成新的激活令牌并发送邮件；已激活或无邮箱则返回错误

令牌说明：
- 签名算法：`HS256`
- 载荷包含：`user_id`、`roles`
- 有效期由配置 `JWT_EXPIRATION` 控制
