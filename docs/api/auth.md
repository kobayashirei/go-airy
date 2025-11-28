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
- 请求体：
```
{
  "email": "user@example.com",
  "code": "123456"
}
```
- 响应：`SUCCESS`

## 登录（密码）
- `POST /login`
- 请求体：
```
{
  "email": "user@example.com",
  "password": "******"
}
```
- 响应：
```
{
  "code": "SUCCESS",
  "data": {
    "access_token": "<jwt>",
    "expires_in": 86400
  }
}
```

## 登录（验证码）
- `POST /login/code`
- 请求体：
```
{
  "email": "user@example.com",
  "code": "123456"
}
```
- 响应：同上，返回 `access_token`

## 刷新令牌
- `POST /refresh`
- 请求头：`Authorization: Bearer <旧JWT>`
- 响应：返回新的 `access_token`

令牌说明：
- 签名算法：`HS256`
- 载荷包含：`user_id`、`roles`
- 有效期由配置 `JWT_EXPIRATION` 控制
