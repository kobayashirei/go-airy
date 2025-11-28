# Airy Backend

A high-performance community platform backend built with Go.

## Project Information

- **Project Name**: Airy
- **Version**: v0.1.0
- **Author**: Rei
- **GitHub**: [kobayashirei](https://github.com/kobayashirei)
- **Website**: [iqwq.com](https://iqwq.com)
- **Repository**: github.com/kobayashirei/airy

## Features

- RESTful API with Gin framework
- Structured logging with Zap
- Configuration management with Viper
- Environment-based configuration
- Request logging and error handling middleware
- Graceful shutdown
- Health check endpoint

## Project Structure

```
.
├── cmd/
│   └── server/          # Application entry point
│       └── main.go
├── internal/
│   ├── config/          # Configuration management
│   ├── logger/          # Logging utilities
│   ├── middleware/      # HTTP middleware
│   └── response/        # Response utilities
├── .env.example         # Example environment variables
├── .gitignore
├── go.mod
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- MySQL/PostgreSQL (for future database integration)
- Redis (for future caching)
- Elasticsearch (for future search functionality)

### Installation

1. Clone the repository
2. Copy `.env.example` to `.env` and configure your environment variables:

```bash
cp .env.example .env
```

3. Install dependencies:

```bash
go mod download
```

4. Run the server:

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080` by default.

### API Endpoints

- `GET /health` - Health check endpoint
- `GET /api/v1/ping` - Ping endpoint

## Configuration

Configuration is managed through environment variables. See `.env.example` for all available options.

Key configuration areas:
- Server settings (host, port, mode)
- Database connection
- Redis connection
- JWT settings
- Logging configuration
- Goroutine pool size

## Development

### Running in Development Mode

```bash
GIN_MODE=debug go run cmd/server/main.go
```

### Building for Production

```bash
go build -o bin/server cmd/server/main.go
```

## Logging

The application uses structured logging with Zap. All HTTP requests are logged with:
- Request ID (UUID)
- HTTP method and path
- Response status code
- Latency
- Client IP
- User agent
- User ID (if authenticated)

Error logs include:
- Error message
- Stack trace
- Request context

## Documentation / 文档

[中文文档](README_CN.md) | [English Documentation](README.md)

### English
- [API Documentation](docs/API.md)
- [Configuration Guide](docs/CONFIGURATION.md)
- [Database Setup](docs/DATABASE_SETUP.md)
- [Authentication Implementation](docs/AUTH_IMPLEMENTATION.md)
- [Async System](docs/ASYNC_SYSTEM.md)
- [Hotness System](docs/HOTNESS_SYSTEM.md)
- [Monitoring](docs/MONITORING.md)

### 中文
- [API 文档](docs/API_CN.md)
- [配置指南](docs/CONFIGURATION_CN.md)
- [数据库设置](docs/DATABASE_SETUP_CN.md)
- [认证实现](docs/AUTH_IMPLEMENTATION_CN.md)
- [异步系统](docs/ASYNC_SYSTEM_CN.md)
- [热度系统](docs/HOTNESS_SYSTEM_CN.md)
- [监控指南](docs/MONITORING_CN.md)

## License

MIT
