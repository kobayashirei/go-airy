# Project Structure

## Project Information

- **Project Name**: Airy
- **Version**: v0.1.0
- **Author**: Rei
- **GitHub**: kobayashirei
- **Website**: iqwq.com
- **Repository**: github.com/kobayashirei/airy

This document describes the organization of the Airy backend codebase.

## Directory Layout

```
.
├── cmd/                    # Application entry points
│   └── server/            # Main server application
│       └── main.go        # Server initialization and startup
│
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   │   ├── config.go     # Configuration structures and loading
│   │   └── config_test.go
│   │
│   ├── logger/           # Logging utilities
│   │   └── logger.go     # Zap logger initialization and helpers
│   │
│   ├── middleware/       # HTTP middleware
│   │   ├── cors.go       # CORS middleware
│   │   ├── logger.go     # Request logging middleware
│   │   ├── logger_test.go
│   │   └── recovery.go   # Panic recovery and error logging
│   │
│   ├── response/         # Response utilities
│   │   └── response.go   # Standard response formats
│   │
│   └── version/          # Version information
│       └── version.go    # Project metadata and version constants
│
├── docs/                 # Documentation
│   └── PROJECT_STRUCTURE.md
│
├── .env.example          # Example environment variables
├── .gitignore           # Git ignore rules
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── Makefile             # Build and development tasks
└── README.md            # Project overview
```

## Package Descriptions

### cmd/server

The main application entry point. Responsible for:
- Loading configuration
- Initializing logger
- Setting up the Gin router
- Configuring middleware
- Starting the HTTP server
- Graceful shutdown handling

### internal/config

Configuration management using Viper and environment variables. Features:
- Structured configuration with validation
- Support for .env files
- Default values for all settings
- Helper methods for connection strings

Configuration sections:
- Server (host, port, mode)
- Database (MySQL/PostgreSQL)
- Redis
- Elasticsearch
- JWT
- Message Queue
- Logging
- Goroutine Pool
- Cache
- Feed

### internal/logger

Structured logging using Zap. Features:
- JSON and console output formats
- Configurable log levels
- File and stdout output
- Automatic caller information
- Stack traces for errors

### internal/middleware

HTTP middleware components:

**RequestLogger**: Logs all HTTP requests with:
- Request ID (UUID)
- Method, path, query
- Response status and size
- Latency
- Client IP and user agent
- User ID (if authenticated)

**Recovery**: Recovers from panics and logs errors with:
- Stack traces
- Request context
- Error details

**ErrorLogger**: Logs errors from handlers with context

**CORS**: Handles Cross-Origin Resource Sharing

### internal/response

Standard response formats for API endpoints:
- Success responses with data
- Error responses with codes and details
- Helper functions for common HTTP status codes
- Request ID tracking
- Timestamps

### internal/version

Project metadata and version information:
- Project name, version, author
- GitHub username and website
- Repository path
- Centralized version management
- Version info endpoint support

## Design Principles

### Layered Architecture

The application follows a layered architecture pattern:
1. **Handler Layer**: HTTP request/response handling
2. **Service Layer**: Business logic (to be implemented)
3. **Repository Layer**: Data access (to be implemented)

### Configuration Management

- All configuration through environment variables
- Validation at startup
- Type-safe configuration structures
- No hardcoded values

### Logging

- Structured logging for machine parsing
- Request ID tracking across the request lifecycle
- Comprehensive error logging with context
- Performance metrics (latency, response size)

### Error Handling

- Panic recovery at the top level
- Structured error responses
- Request ID in all error responses
- No sensitive information in error messages

## Future Additions

The following directories will be added as development progresses:

```
internal/
├── handler/          # HTTP handlers
├── service/          # Business logic
├── repository/       # Data access layer
├── model/            # Data models
├── cache/            # Cache service
├── queue/            # Message queue
├── auth/             # Authentication and authorization
└── util/             # Utility functions
```

## Testing

Tests are co-located with the code they test:
- Unit tests: `*_test.go` files
- Test coverage: Run `make test-coverage`
- Integration tests: To be added in `test/` directory

## Building and Running

See the main README.md for build and run instructions.

Common commands:
```bash
make run          # Run the application
make build        # Build binary
make test         # Run tests
make fmt          # Format code
```
