# 系统与健康

## 版本信息
- `GET /version`
- 响应：项目名、版本、作者、仓库等信息

## 健康检查
- `GET /health`
- 响应字段：
  - `status`: `healthy` / `degraded` / `unhealthy`
  - `database`: `healthy` / `disabled` / `unavailable` / `error`
  - `cache`: `healthy` / `disabled` / `error`
  - `time`: ISO8601 格式当前时间

## 指标监控
- `GET /metrics`
- Prometheus 指标拉取端点
