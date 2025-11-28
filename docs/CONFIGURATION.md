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

### CSRF Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CSRF_TOKEN_LENGTH` | `32` | Token length in bytes |
| `CSRF_TOKEN_EXPIRATION` | `86400` | Token expiration (seconds) |
| `CSRF_COOKIE_NAME` | `_csrf` | Cookie name |
| `CSRF_HEADER_NAME` | `X-CSRF-Token` | Header name |
| `CSRF_FORM_FIELD_NAME` | `_csrf` | Form field name |
| `CSRF_SECURE` | `false` | Secure cookie flag |
| `CSRF_SAME_SITE` | `Strict` | SameSite attribute |

### CORS Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000,http://localhost:8080` | Allowed origins (comma-separated) |
| `CORS_ALLOWED_METHODS` | `GET,POST,PUT,DELETE,OPTIONS,PATCH` | Allowed methods |
| `CORS_ALLOWED_HEADERS` | `Origin,Content-Type,Accept,Authorization,X-CSRF-Token` | Allowed headers |
| `CORS_EXPOSED_HEADERS` | `Content-Length,Content-Type` | Exposed headers |
| `CORS_ALLOW_CREDENTIALS` | `true` | Allow credentials |
| `CORS_MAX_AGE` | `86400` | Preflight cache max age (seconds) |

### Application URLs

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_BASE_URL` | `http://localhost:8080` | Application base URL |
| `FRONTEND_URL` | `http://localhost:3000` | Frontend URL |

### Token Expiration

| Variable | Default | Description |
|----------|---------|-------------|
| `ACTIVATION_TOKEN_EXPIRATION` | `86400` | Activation token TTL (seconds) |
| `PASSWORD_RESET_TOKEN_EXPIRATION` | `3600` | Password reset token TTL (seconds) |
| `VERIFICATION_CODE_EXPIRATION` | `300` | Verification code TTL (seconds) |

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

# Security settings
CSRF_SECURE=true
CSRF_SAME_SITE=Strict
CORS_ALLOWED_ORIGINS=https://yourdomain.com
CORS_ALLOW_CREDENTIALS=true

# Application URLs
APP_BASE_URL=https://api.yourdomain.com
FRONTEND_URL=https://yourdomain.com

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

## Security Best Practices

### Secrets Management

1. **Never commit secrets to version control**
   - Use `.env` files locally (add to `.gitignore`)
   - Use environment variables or secret managers in production

2. **Generate strong secrets**
   ```bash
   # JWT Secret (64 hex characters)
   openssl rand -hex 32
   
   # Encryption Key (base64, 32 bytes)
   openssl rand -base64 32
   ```

3. **Rotate secrets regularly**
   - JWT secrets should be rotated periodically
   - Encryption keys require careful migration when rotated

### Production Checklist

- [ ] `GIN_MODE=release`
- [ ] `TLS_ENABLED=true` with valid certificates
- [ ] `JWT_SECRET` is a strong, unique value
- [ ] `ENCRYPTION_KEY` is set for sensitive data encryption
- [ ] `CSRF_SECURE=true` (requires HTTPS)
- [ ] `CORS_ALLOWED_ORIGINS` is restricted to your domains
- [ ] `RATE_LIMIT_ENABLED=true`
- [ ] Database credentials are not default values
- [ ] Redis password is set if exposed
- [ ] Log level is `info` or higher (not `debug`)

### Environment-Specific Settings

| Setting | Development | Production |
|---------|-------------|------------|
| `GIN_MODE` | `debug` | `release` |
| `LOG_LEVEL` | `debug` | `info` |
| `TLS_ENABLED` | `false` | `true` |
| `CSRF_SECURE` | `false` | `true` |
| `RATE_LIMIT_ENABLED` | `false` | `true` |
| `CORS_ALLOWED_ORIGINS` | `*` or localhost | Specific domains |

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

## Troubleshooting

### Common Issues

#### JWT Secret Error
```
Error: JWT secret must be set and changed from default
```
**Solution**: Set a unique `JWT_SECRET` value:
```bash
export JWT_SECRET=$(openssl rand -hex 32)
```

#### Database Connection Failed
```
Error: dial tcp: connect: connection refused
```
**Solution**: 
- Verify `DB_HOST` and `DB_PORT` are correct
- Ensure MySQL/PostgreSQL is running
- Check firewall rules

#### Redis Connection Failed
```
Error: dial tcp: connect: connection refused
```
**Solution**:
- Verify `REDIS_HOST` and `REDIS_PORT` are correct
- Ensure Redis is running
- Check `REDIS_PASSWORD` if authentication is enabled

#### TLS Certificate Error
```
Error: TLS certificate file is required when TLS is enabled
```
**Solution**:
- Set `TLS_CERT_FILE` and `TLS_KEY_FILE` paths
- Or disable TLS with `TLS_ENABLED=false` for development

#### CORS Errors
```
Access-Control-Allow-Origin header missing
```
**Solution**:
- Add your frontend URL to `CORS_ALLOWED_ORIGINS`
- Ensure `CORS_ALLOW_CREDENTIALS=true` if using cookies

#### Rate Limiting Issues
```
Error: Too many requests
```
**Solution**:
- Increase `RATE_LIMIT_REQUESTS_PER_SECOND` and `RATE_LIMIT_BURST_SIZE`
- Or disable with `RATE_LIMIT_ENABLED=false` for development

### Debugging Tips

1. **Enable debug logging**:
   ```bash
   LOG_LEVEL=debug
   GIN_MODE=debug
   ```

2. **Check configuration loading**:
   The application logs configuration values at startup (sensitive values are masked).

3. **Validate environment file**:
   ```bash
   # Check for syntax errors
   source .env && echo "Environment file is valid"
   ```

4. **Test database connection**:
   ```bash
   mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "SELECT 1"
   ```

5. **Test Redis connection**:
   ```bash
   redis-cli -h $REDIS_HOST -p $REDIS_PORT ping
   ```

## Configuration Reference

For a complete list of all configuration options with their default values, see the `.env.example` file in the project root.

### Required vs Optional

| Category | Required | Optional |
|----------|----------|----------|
| Server | `SERVER_PORT` | `SERVER_HOST`, `GIN_MODE` |
| Database | `DB_NAME`, `DB_USER` | `DB_PASSWORD`, connection pool settings |
| JWT | `JWT_SECRET` | `JWT_EXPIRATION` |
| Redis | - | All (uses defaults) |
| Elasticsearch | - | All (uses defaults) |
| Security | - | `ENCRYPTION_KEY`, `BCRYPT_COST` |
| TLS | `TLS_CERT_FILE`, `TLS_KEY_FILE` (if enabled) | `TLS_ENABLED` |
