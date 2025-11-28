# 圈子

前缀：`/api/v1/circles`

## 获取圈子信息
- `GET /:id`
- 响应：圈子基本信息与统计

## 获取圈子成员
- `GET /:id/members?page=1&page_size=20`

## 创建圈子（需登录）
- `POST /`
- 请求体：
```
{
  "name": "Go Lovers",
  "description": "Go 语言交流圈"
}
```

## 申请加入圈子（需登录）
- `POST /:id/join`

## 审批成员（需圈主/管理员）
- `POST /:id/members/:userId/approve`

## 指定版主（需圈主/管理员）
- `POST /:id/moderators`
- 请求体：
```
{
  "user_id": 123
}
```
