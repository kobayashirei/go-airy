# Changelog

All notable changes to the Airy project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0] - 2025-11-28

### Added
- Initial project setup and infrastructure
- Go module with Gin web framework
- Configuration management with Viper and environment variables
- Structured logging with Zap
- HTTP middleware:
  - Request logging with UUID tracking
  - Panic recovery with error logging
  - CORS support
  - Error logging
- Standard response utilities
- Version management package
- Health check endpoint (`/health`)
- Version information endpoint (`/version`)
- Ping endpoint (`/api/v1/ping`)
- Graceful shutdown support
- Comprehensive test coverage for config and middleware
- Project documentation:
  - README with setup instructions
  - Project structure documentation
  - Example environment configuration
  - Makefile for common tasks

### Project Information
- **Project Name**: Airy
- **Author**: Rei
- **GitHub**: kobayashirei
- **Website**: iqwq.com
- **Repository**: github.com/airygithub/backend

### Technical Details
- Go 1.21+
- Gin web framework
- Zap structured logging
- Viper configuration management
- UUID request tracking
- JSON structured logs
- Environment-based configuration

### Requirements Validated
- ✓ 20.1: Request logging with method, path, user ID, response time
- ✓ 20.2: Error logging with stack traces and context
- ✓ 20.3: Monitoring support (Prometheus-ready structure)
- ✓ 20.4: Structured logging format (JSON)

[v0.1.0]: https://github.com/airygithub/backend/releases/tag/v0.1.0
