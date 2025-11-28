package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kobayashirei/airy/internal/config"
	appLogger "github.com/kobayashirei/airy/internal/logger"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance
var DB *gorm.DB

// Init initializes the database connection
func Init(cfg *config.DatabaseConfig) error {
	dsn := cfg.GetDSN()

	// Configure GORM logger
	gormLogger := logger.New(
		&gormLogWriter{},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Open database connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		// Fallback: retry with TLS skip-verify if TLS might be required
		if !strings.Contains(dsn, "tls=") && (strings.Contains(err.Error(), "invalid connection") || strings.Contains(err.Error(), "unexpected EOF")) {
			fallbackDSN := dsn + "&tls=skip-verify"
			db, err = gorm.Open(mysql.Open(fallbackDSN), &gorm.Config{
				Logger:                 gormLogger,
				SkipDefaultTransaction: true,
				PrepareStmt:            true,
			})
		}
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db

	appLogger.Info("Database connection established",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Name),
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Duration("conn_max_lifetime", cfg.ConnMaxLifetime),
	)

	return nil
}

// Close closes the database connection
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	appLogger.Info("Database connection closed")
	return nil
}

// HealthCheck checks if the database is healthy
func HealthCheck(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Ping with timeout
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check connection pool stats
	stats := sqlDB.Stats()
	appLogger.Debug("Database connection pool stats",
		zap.Int("open_connections", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int64("wait_count", stats.WaitCount),
		zap.Duration("wait_duration", stats.WaitDuration),
	)

	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// gormLogWriter implements GORM's logger.Writer interface
type gormLogWriter struct{}

// Printf implements the logger.Writer interface
func (w *gormLogWriter) Printf(format string, args ...interface{}) {
	appLogger.Info(fmt.Sprintf(format, args...))
}
