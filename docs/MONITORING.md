# Monitoring and Observability

This document describes the monitoring and observability features implemented in the Airy backend system.

## Overview

The system implements comprehensive logging and monitoring capabilities to ensure operational visibility and facilitate troubleshooting.

## Logging

### Request Logging

**Middleware**: `middleware.RequestLogger()`

**Features**:
- Generates unique request ID for tracing
- Logs HTTP method, path, query parameters
- Records response status code and latency
- Captures client IP and user agent
- Includes authenticated user ID when available
- Measures request body size

**Log Format**: Structured JSON logs using zap logger

**Example Log Entry**:
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

**Validates**: Requirements 20.1

### Error Logging

**Middleware**: `middleware.Recovery()` and `middleware.ErrorLogger()`

**Features**:
- Recovers from panics and logs stack traces
- Logs all errors from request handlers
- Includes request context (method, path, request ID)
- Captures error type and metadata
- Uses structured logging format

**Example Error Log**:
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

**Validates**: Requirements 20.2, 20.4

## Prometheus Metrics

**Middleware**: `middleware.PrometheusMetrics()`

**Endpoint**: `GET /metrics`

### Metrics Collected

#### 1. HTTP Request Count
**Metric**: `http_requests_total`
**Type**: Counter
**Labels**: method, path, status
**Description**: Total number of HTTP requests processed

#### 2. HTTP Request Duration
**Metric**: `http_request_duration_seconds`
**Type**: Histogram
**Labels**: method, path, status
**Description**: Duration of HTTP requests in seconds
**Buckets**: Default Prometheus buckets (0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10)

#### 3. HTTP Requests In-Flight
**Metric**: `http_requests_in_flight`
**Type**: Gauge
**Description**: Number of HTTP requests currently being processed

#### 4. HTTP Error Count
**Metric**: `http_errors_total`
**Type**: Counter
**Labels**: method, path, status
**Description**: Total number of HTTP errors (status >= 400)

### Example Metrics Output

```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/v1/posts",status="200"} 1523

# HELP http_request_duration_seconds Duration of HTTP requests in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",path="/api/v1/posts",status="200",le="0.005"} 120
http_request_duration_seconds_bucket{method="GET",path="/api/v1/posts",status="200",le="0.01"} 450
http_request_duration_seconds_bucket{method="GET",path="/api/v1/posts",status="200",le="0.025"} 1200
http_request_duration_seconds_sum{method="GET",path="/api/v1/posts",status="200"} 15.234
http_request_duration_seconds_count{method="GET",path="/api/v1/posts",status="200"} 1523

# HELP http_requests_in_flight Number of HTTP requests currently being processed
# TYPE http_requests_in_flight gauge
http_requests_in_flight 5

# HELP http_errors_total Total number of HTTP errors
# TYPE http_errors_total counter
http_errors_total{method="GET",path="/api/v1/posts",status="404"} 23
http_errors_total{method="POST",path="/api/v1/posts",status="500"} 2
```

**Validates**: Requirements 20.3

## Prometheus Integration

### Scrape Configuration

Add the following to your Prometheus configuration (`prometheus.yml`):

```yaml
scrape_configs:
  - job_name: 'airy-backend'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### Grafana Dashboard

You can create Grafana dashboards using the following queries:

**Request Rate**:
```promql
rate(http_requests_total[5m])
```

**Error Rate**:
```promql
rate(http_errors_total[5m])
```

**Request Latency (p95)**:
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

**In-Flight Requests**:
```promql
http_requests_in_flight
```

## Log Configuration

Configure logging through environment variables or config file:

```yaml
log:
  level: "info"        # debug, info, warn, error, fatal
  output: "stdout"     # stdout or file
  file_path: "logs/app.log"  # if output is file
```

## Best Practices

1. **Request Tracing**: Use the `request_id` field to trace requests across logs
2. **Error Investigation**: Check error logs with stack traces for debugging
3. **Performance Monitoring**: Monitor request duration histograms for performance issues
4. **Alerting**: Set up alerts for high error rates or slow response times
5. **Log Rotation**: Configure log rotation for file-based logging to prevent disk space issues

## Health Checks

The system provides health check endpoints:

- `GET /health` - Overall system health (includes database and cache status)
- `GET /version` - Application version information
- `GET /metrics` - Prometheus metrics

## Future Enhancements

- [ ] Distributed tracing with OpenTelemetry
- [ ] Custom business metrics (user registrations, post creations, etc.)
- [ ] Log aggregation with ELK stack or similar
- [ ] Real-time alerting integration
- [ ] Performance profiling endpoints
