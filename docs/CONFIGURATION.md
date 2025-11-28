# Airy Configuration Guide

## Overview

Airy uses environment variables for configuration. Copy `.env.example` to `.env` and modify values for your environment.

```bash
cp .env.example .env
```

## Configuration Categories

### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Server bind address |
| `SERVER_PORT` | `8080` | Server port |
| `GIN_MODE` | `debug` | Gin mode: `debug`, `release`, `test` |

### Database Configuration (MySQL)

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `3306` | Database port |
| `DB_NAME` | `airygithub` | Database name |
| `DB_USER` | `root` | Database user |
| `DB_PASSWORD` | - | Database password |
| `DB_MAX_IDLE_CONNS` | `10` | Max idle connections |
| `DB_MAX_OPEN_CONNS` | `100` | Max open connections |
| `DB_CONN_MAX_LIFETIME` | `3600` | Connection max lifetime (seconds) |

### Redis Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | - | Redis password |
| `REDIS_DB` | `0` | Redis database number |

### Elasticsearch Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ES_HOST` | `localhost` | Elasticsearch host |
| `ES_PORT` | `9200` | Elasticsearch port |

### JWT Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | - | **Required.** JWT signing secret |
| `JWT_EXPIRATION` | `86400` | Token expiration (seconds) |

Generate a secure secret:
```bash
openssl rand -hex 32
```

### Message Queue (RabbitMQ)

| Variable | Default | Description |
|----------|---------|-------------|
| `MQ_HOST` | `localhost` | RabbitMQ host |
| `MQ_PORT` | `5672` | RabbitMQ port |
| `MQ_USER` | `guest` | RabbitMQ user |
| `MQ_PASSWORD` | `guest` | RabbitMQ password |

### Logging Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `LOG_OUTPUT` | `stdout` | Output: `stdout`, `file` |
| `LOG_FILE_PATH` | `logs/app.log` | Log file path |

### Cache Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CACHE_DEFAULT_EXPIRATION` | `3600` | Default TTL (seconds) |
| `CACHE_CLEANUP_INTERVAL` | `600` | Cleanup interval (seconds) |
| `CACHE_WARMUP_ENABLED` | `true` | Enable cache warmup |
| `CACHE_WARMUP_HOT_POSTS` | `100` | Hot posts to preload |
| `CACHE_WARMUP_HOT_USERS` | `50` | Hot users to preload |
| `CACHE_WARMUP_HOT_CIRCLES` | `20` | Hot circles to preload |
| `CACHE_WARMUP_REFRESH_INTERVAL` | `30` | Refresh interval (minutes) |
| `CACHE_WARMUP_CONCURRENCY` | `10` | Warmup concurrency |

### Feed Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `FEED_FANOUT_THRESHOLD` | `1000` | Follower threshold for fan-out strategy |

### Hotness Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HOTNESS_ALGORITHM` | `reddit` | Algorithm: `reddit`, `hackernews` |

### Rate Limiting

| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMIT_ENABLED` | `true` | Enable rate limiting |
| `RATE_LIMIT_REQUESTS_PER_SECOND` | `100` | Requests per second |
| `RATE_LIMIT_BURST_SIZE` | `200` | Burst size |

### TLS/HTTPS Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `TLS_ENABLED` | `false` | Enable TLS |
| `TLS_CERT_FILE` | - | Certificate file path |
| `TLS_KEY_FILE` | - | Key file path |

### Security Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ENCRYPTION_KEY` | - | AES encryption key (base64) |
| `BCRYPT_COST` | `10` | Bcrypt cost factor (4-31) |

Generate encryption key:
```bash
openssl rand -base64 32
```

### Email Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SMTP_HOST` | - | SMTP server host |
| `SMTP_PORT` | `587` | SMTP port |
| `SMTP_USER` | - | SMTP username |
| `SMTP_PASSWORD` | - | SMTP password |
| `SMTP_FROM` | - | From email address |

### Content Moderation

| Variable | Default | Description |
|----------|---------|-------------|
| `MODERATION_API_URL` | - | Moderation API URL |
| `MODERATION_API_KEY` | - | Moderation API key |

### Goroutine Pool

| Variable | Default | Description |
|----------|---------|-------------|
| `GOROUTINE_POOL_SIZE` | `10000` | Pool size |

## Environment Examples

### Development

```bash
GIN_MODE=debug
LOG_LEVEL=debug
DB_HOST=localhost
REDIS_HOST=localhost
JWT_SECRET=dev-secret-key-change-in-production
```

### Production

```bash
GIN_MODE=release
LOG_LEVEL=info
LOG_OUTPUT=file
TLS_ENABLED=true
RATE_LIMIT_ENABLED=true
CACHE_WARMUP_ENABLED=true
# Use strong secrets!
JWT_SECRET=<generated-secret>
ENCRYPTION_KEY=<generated-key>
```

## Docker Compose

Example `docker-compose.yml`:

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    env_file:
      - .env
    depends_on:
      - mysql
      - redis
      - elasticsearch
      - rabbitmq

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: airygithub
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

  elasticsearch:
    image: elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    volumes:
      - es_data:/usr/share/elasticsearch/data

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "15672:15672"

volumes:
  mysql_data:
  redis_data:
  es_data:
```

## Quick Start

1. Copy environment file:
   ```bash
   cp .env.example .env
   ```

2. Set required values:
   ```bash
   # Generate JWT secret
   JWT_SECRET=$(openssl rand -hex 32)
   
   # Set database password
   DB_PASSWORD=your-password
   ```

3. Run migrations:
   ```bash
   make migrate-up
   ```

4. Start server:
   ```bash
   make run
   ```
