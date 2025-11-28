package config

import (
    "fmt"
    "time"

    "github.com/joho/godotenv"
    "github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	ES        ESConfig
	JWT       JWTConfig
	MQ        MQConfig
	Log       LogConfig
	Pool      PoolConfig
	Cache     CacheConfig
	Feed      FeedConfig
	Hotness   HotnessConfig
	RateLimit RateLimitConfig
	Security  SecurityConfig
	Features  FeaturesConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host        string
	Port        int
	Mode        string
	TLSEnabled  bool
	TLSCertFile string
	TLSKeyFile  string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	InitLockFile    string
	TLSMode         string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// ESConfig holds Elasticsearch configuration
type ESConfig struct {
	Host string
	Port int
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// MQConfig holds message queue configuration
type MQConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level    string
	Output   string
	FilePath string
}

// PoolConfig holds goroutine pool configuration
type PoolConfig struct {
	Size int
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	DefaultExpiration time.Duration
	CleanupInterval   time.Duration
	// Warmup configuration
	WarmupEnabled         bool
	WarmupHotPosts        int
	WarmupHotUsers        int
	WarmupHotCircles      int
	WarmupRefreshInterval time.Duration
	WarmupConcurrency     int
}

// FeedConfig holds feed configuration
type FeedConfig struct {
	FanoutThreshold int
}

// HotnessConfig holds hotness calculation configuration
type HotnessConfig struct {
	Algorithm string // "reddit" or "hackernews"
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool
	RequestsPerSecond int
	BurstSize         int
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	// EncryptionKey is the base64-encoded AES encryption key for sensitive data
	// Must be 16, 24, or 32 bytes when decoded (AES-128, AES-192, or AES-256)
	EncryptionKey string
	// BcryptCost is the cost factor for bcrypt password hashing (4-31, default 10)
	BcryptCost int
}

// FeaturesConfig holds feature toggles for local development
type FeaturesConfig struct {
    EnableDatabase      bool
    EnableRedis         bool
    AutoCreateDB        bool
    AutoMigrate         bool
    AllowStartWithoutDB bool
    UseGormAutoMigrate  bool
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
    // Load .env file if it exists
    _ = godotenv.Load(".env.local")
    _ = godotenv.Load()

	// Set up viper
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	config := &Config{
		Server: ServerConfig{
			Host:        viper.GetString("SERVER_HOST"),
			Port:        viper.GetInt("SERVER_PORT"),
			Mode:        viper.GetString("GIN_MODE"),
			TLSEnabled:  viper.GetBool("TLS_ENABLED"),
			TLSCertFile: viper.GetString("TLS_CERT_FILE"),
			TLSKeyFile:  viper.GetString("TLS_KEY_FILE"),
		},
		Database: DatabaseConfig{
			Host:            viper.GetString("DB_HOST"),
			Port:            viper.GetInt("DB_PORT"),
			Name:            viper.GetString("DB_NAME"),
			User:            viper.GetString("DB_USER"),
			Password:        viper.GetString("DB_PASSWORD"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			ConnMaxLifetime: viper.GetDuration("DB_CONN_MAX_LIFETIME") * time.Second,
			InitLockFile:    viper.GetString("DB_INIT_LOCK_FILE"),
			TLSMode:         viper.GetString("DB_TLS_MODE"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetInt("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		ES: ESConfig{
			Host: viper.GetString("ES_HOST"),
			Port: viper.GetInt("ES_PORT"),
		},
		JWT: JWTConfig{
			Secret:     viper.GetString("JWT_SECRET"),
			Expiration: viper.GetDuration("JWT_EXPIRATION") * time.Second,
		},
		MQ: MQConfig{
			Host:     viper.GetString("MQ_HOST"),
			Port:     viper.GetInt("MQ_PORT"),
			User:     viper.GetString("MQ_USER"),
			Password: viper.GetString("MQ_PASSWORD"),
		},
		Log: LogConfig{
			Level:    viper.GetString("LOG_LEVEL"),
			Output:   viper.GetString("LOG_OUTPUT"),
			FilePath: viper.GetString("LOG_FILE_PATH"),
		},
		Pool: PoolConfig{
			Size: viper.GetInt("GOROUTINE_POOL_SIZE"),
		},
		Cache: CacheConfig{
			DefaultExpiration:     viper.GetDuration("CACHE_DEFAULT_EXPIRATION") * time.Second,
			CleanupInterval:       viper.GetDuration("CACHE_CLEANUP_INTERVAL") * time.Second,
			WarmupEnabled:         viper.GetBool("CACHE_WARMUP_ENABLED"),
			WarmupHotPosts:        viper.GetInt("CACHE_WARMUP_HOT_POSTS"),
			WarmupHotUsers:        viper.GetInt("CACHE_WARMUP_HOT_USERS"),
			WarmupHotCircles:      viper.GetInt("CACHE_WARMUP_HOT_CIRCLES"),
			WarmupRefreshInterval: viper.GetDuration("CACHE_WARMUP_REFRESH_INTERVAL") * time.Minute,
			WarmupConcurrency:     viper.GetInt("CACHE_WARMUP_CONCURRENCY"),
		},
		Feed: FeedConfig{
			FanoutThreshold: viper.GetInt("FEED_FANOUT_THRESHOLD"),
		},
		Hotness: HotnessConfig{
			Algorithm: viper.GetString("HOTNESS_ALGORITHM"),
		},
		RateLimit: RateLimitConfig{
			Enabled:           viper.GetBool("RATE_LIMIT_ENABLED"),
			RequestsPerSecond: viper.GetInt("RATE_LIMIT_REQUESTS_PER_SECOND"),
			BurstSize:         viper.GetInt("RATE_LIMIT_BURST_SIZE"),
		},
		Security: SecurityConfig{
			EncryptionKey: viper.GetString("ENCRYPTION_KEY"),
			BcryptCost:    viper.GetInt("BCRYPT_COST"),
		},
		Features: FeaturesConfig{
			EnableDatabase:      viper.GetBool("ENABLE_DATABASE"),
			EnableRedis:         viper.GetBool("ENABLE_REDIS"),
			AutoCreateDB:        viper.GetBool("ENABLE_DB_AUTO_CREATE"),
			AutoMigrate:         viper.GetBool("ENABLE_DB_AUTO_MIGRATE"),
			AllowStartWithoutDB: viper.GetBool("ALLOW_START_WITHOUT_DB"),
			UseGormAutoMigrate:  viper.GetBool("USE_GORM_AUTOMIGRATE"),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("SERVER_PORT", 8080)
	viper.SetDefault("GIN_MODE", "debug")

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 3306)
	viper.SetDefault("DB_NAME", "airygithub")
	viper.SetDefault("DB_USER", "root")
	viper.SetDefault("DB_PASSWORD", "")
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_MAX_OPEN_CONNS", 100)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", 3600)
	viper.SetDefault("DB_INIT_LOCK_FILE", "./data/db_init.lock")
	viper.SetDefault("DB_TLS_MODE", "preferred")

	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("ES_HOST", "localhost")
	viper.SetDefault("ES_PORT", 9200)

	viper.SetDefault("JWT_SECRET", "change-me-in-production")
	viper.SetDefault("JWT_EXPIRATION", 86400)

	viper.SetDefault("MQ_HOST", "localhost")
	viper.SetDefault("MQ_PORT", 5672)
	viper.SetDefault("MQ_USER", "guest")
	viper.SetDefault("MQ_PASSWORD", "guest")

	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_OUTPUT", "stdout")
	viper.SetDefault("LOG_FILE_PATH", "logs/app.log")

	viper.SetDefault("GOROUTINE_POOL_SIZE", 10000)

	viper.SetDefault("CACHE_DEFAULT_EXPIRATION", 3600)
	viper.SetDefault("CACHE_CLEANUP_INTERVAL", 600)
	viper.SetDefault("CACHE_WARMUP_ENABLED", true)
	viper.SetDefault("CACHE_WARMUP_HOT_POSTS", 100)
	viper.SetDefault("CACHE_WARMUP_HOT_USERS", 50)
	viper.SetDefault("CACHE_WARMUP_HOT_CIRCLES", 20)
	viper.SetDefault("CACHE_WARMUP_REFRESH_INTERVAL", 30)
	viper.SetDefault("CACHE_WARMUP_CONCURRENCY", 10)

	viper.SetDefault("FEED_FANOUT_THRESHOLD", 1000)

	viper.SetDefault("HOTNESS_ALGORITHM", "reddit")

	viper.SetDefault("RATE_LIMIT_ENABLED", true)
	viper.SetDefault("RATE_LIMIT_REQUESTS_PER_SECOND", 100)
	viper.SetDefault("RATE_LIMIT_BURST_SIZE", 200)

	// TLS/HTTPS defaults
	viper.SetDefault("TLS_ENABLED", false)
	viper.SetDefault("TLS_CERT_FILE", "")
	viper.SetDefault("TLS_KEY_FILE", "")

	// Security defaults
	viper.SetDefault("ENCRYPTION_KEY", "")
	viper.SetDefault("BCRYPT_COST", 10)

	// Feature toggles (useful for local development)
	viper.SetDefault("ENABLE_DATABASE", true)
	viper.SetDefault("ENABLE_REDIS", true)
	viper.SetDefault("ENABLE_DB_AUTO_CREATE", true)
	viper.SetDefault("ENABLE_DB_AUTO_MIGRATE", true)
	viper.SetDefault("ALLOW_START_WITHOUT_DB", true)
	viper.SetDefault("USE_GORM_AUTOMIGRATE", true)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.JWT.Secret == "" || c.JWT.Secret == "change-me-in-production" {
		return fmt.Errorf("JWT secret must be set and changed from default")
	}

	if c.Features.EnableDatabase {
		if c.Database.Name == "" {
			return fmt.Errorf("database name is required")
		}
	}

	if c.Pool.Size <= 0 {
		return fmt.Errorf("goroutine pool size must be positive")
	}

	// Validate TLS configuration
	if c.Server.TLSEnabled {
		if c.Server.TLSCertFile == "" {
			return fmt.Errorf("TLS certificate file is required when TLS is enabled")
		}
		if c.Server.TLSKeyFile == "" {
			return fmt.Errorf("TLS key file is required when TLS is enabled")
		}
	}

	// Validate bcrypt cost
	if c.Security.BcryptCost < 4 || c.Security.BcryptCost > 31 {
		return fmt.Errorf("bcrypt cost must be between 4 and 31")
	}

	return nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	params := "charset=utf8mb4&parseTime=True&loc=Local&timeout=5s&readTimeout=5s&writeTimeout=5s"
	if c.TLSMode != "" {
		params = params + "&tls=" + c.TLSMode
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		c.User, c.Password, c.Host, c.Port, c.Name, params)
}

// GetRedisAddr returns the Redis address
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetESAddr returns the Elasticsearch address
func (c *ESConfig) GetAddr() string {
	return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
}

// GetMQAddr returns the message queue address
func (c *MQConfig) GetAddr() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", c.User, c.Password, c.Host, c.Port)
}
