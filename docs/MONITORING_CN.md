# 监控与可观测性

本文档描述 Airy 后端系统实现的监控和可观测性功能。

## 概述

系统实现了全面的日志和监控功能，以确保运营可见性并便于故障排除。

## 日志

### 请求日志

**中间件**: `middleware.RequestLogger()`

**功能特性**:
- 生成唯一请求 ID 用于追踪
- 记录 HTTP 方法、路径、查询参数
- 记录响应状态码和延迟
- 捕获客户端 IP 和 User Agent
- 包含已认证用户 ID（如果可用）
- 测量请求体大小

**日志格式**: 使用 zap 日志器的结构化 JSON 日志

**示例日志条目**:
```json
{
  "timestamp": "2025-11-28T10:26:12.618+0800",
  "level": "info",
  "message": "HTTP Request",
  "request_id": "3b2939ee-0f4a-41c5-9d42-ede279de4be7",
  "method": "GET",
  "path": "/api/v1/posts",
  "query": "page=1&limit=10",
  "status": 200,
  "latency": 0.025,
  "client_ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "user_id": 123,
  "body_size": 1024
}
```

**验证**: 需求 20.1

### 错误日志

**中间件**: `middleware.Recovery()` 和 `middleware.ErrorLogger()`

**功能特性**:
- 从 panic 恢复并记录堆栈跟踪
- 记录请求处理器的所有错误
- 包含请求上下文（方法、路径、请求 ID）
- 捕获错误类型和元数据
- 使用结构化日志格式

**示例错误日志**:
```json
{
  "timestamp": "2025-11-28T10:26:12.618+0800",
  "level": "error",
  "message": "Panic recovered",
  "request_id": "3b2939ee-0f4a-41c5-9d42-ede279de4be7",
  "method": "POST",
  "path": "/api/v1/posts",
  "error": "runtime error: invalid memory address",
  "stack_trace": "goroutine 1 [running]:\n...",
  "client_ip": "192.168.1.100"
}
```

**验证**: 需求 20.2, 20.4

## Prometheus 指标

**中间件**: `middleware.PrometheusMetrics()`

**端点**: `GET /metrics`

### 收集的指标

#### 1. HTTP 请求计数
**指标**: `http_requests_total`
**类型**: Counter
**标签**: method, path, status
**描述**: 处理的 HTTP 请求总数

#### 2. HTTP 请求持续时间
**指标**: `http_request_duration_seconds`
**类型**: Histogram
**标签**: method, path, status
**描述**: HTTP 请求持续时间（秒）
**桶**: 默认 Prometheus 桶 (0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10)

#### 3. 进行中的 HTTP 请求
**指标**: `http_requests_in_flight`
**类型**: Gauge
**描述**: 当前正在处理的 HTTP 请求数

#### 4. HTTP 错误计数
**指标**: `http_errors_total`
**类型**: Counter
**标签**: method, path, status
**描述**: HTTP 错误总数（状态码 >= 400）

### 示例指标输出

```
# HELP http_requests_total HTTP 请求总数
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/v1/posts",status="200"} 1523

# HELP http_request_duration_seconds HTTP 请求持续时间（秒）
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",path="/api/v1/posts",status="200",le="0.005"} 120
http_request_duration_seconds_bucket{method="GET",path="/api/v1/posts",status="200",le="0.01"} 450
http_request_duration_seconds_bucket{method="GET",path="/api/v1/posts",status="200",le="0.025"} 1200
http_request_duration_seconds_sum{method="GET",path="/api/v1/posts",status="200"} 15.234
http_request_duration_seconds_count{method="GET",path="/api/v1/posts",status="200"} 1523

# HELP http_requests_in_flight 当前正在处理的 HTTP 请求数
# TYPE http_requests_in_flight gauge
http_requests_in_flight 5

# HELP http_errors_total HTTP 错误总数
# TYPE http_errors_total counter
http_errors_total{method="GET",path="/api/v1/posts",status="404"} 23
http_errors_total{method="POST",path="/api/v1/posts",status="500"} 2
```

**验证**: 需求 20.3

## Prometheus 集成

### 抓取配置

将以下内容添加到您的 Prometheus 配置 (`prometheus.yml`)：

```yaml
scrape_configs:
  - job_name: 'airy-backend'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### Grafana 仪表盘

您可以使用以下查询创建 Grafana 仪表盘：

**请求速率**:
```promql
rate(http_requests_total[5m])
```

**错误率**:
```promql
rate(http_errors_total[5m])
```

**请求延迟 (p95)**:
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

**进行中的请求**:
```promql
http_requests_in_flight
```

## 日志配置

通过环境变量或配置文件配置日志：

```yaml
log:
  level: "info"        # debug, info, warn, error, fatal
  output: "stdout"     # stdout 或 file
  file_path: "logs/app.log"  # 如果 output 是 file
```

## 最佳实践

1. **请求追踪**: 使用 `request_id` 字段在日志中追踪请求
2. **错误调查**: 检查带堆栈跟踪的错误日志进行调试
3. **性能监控**: 监控请求持续时间直方图以发现性能问题
4. **告警**: 为高错误率或慢响应时间设置告警
5. **日志轮转**: 为基于文件的日志配置日志轮转以防止磁盘空间问题

## 健康检查

系统提供健康检查端点：

- `GET /health` - 整体系统健康状态（包括数据库和缓存状态）
- `GET /version` - 应用版本信息
- `GET /metrics` - Prometheus 指标

## 未来增强

- [ ] 使用 OpenTelemetry 的分布式追踪
- [ ] 自定义业务指标（用户注册、帖子创建等）
- [ ] 使用 ELK 栈或类似工具的日志聚合
- [ ] 实时告警集成
- [ ] 性能分析端点
