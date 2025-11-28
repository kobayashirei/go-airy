package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kobayashirei/airy/internal/config"
	appLogger "github.com/kobayashirei/airy/internal/logger"
	"go.uber.org/zap"
)

// EnsureDatabaseExists creates the target database if it does not exist
func EnsureDatabaseExists(cfg *config.DatabaseConfig) error {
	// Connect without specifying database name
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	tlsMode := os.Getenv("DB_TLS_MODE")
	if tlsMode == "" {
		tlsMode = "preferred"
	}
	if tlsMode != "" {
		dsn = dsn + fmt.Sprintf("?tls=%s", tlsMode)
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open server connection: %w", err)
	}
	defer db.Close()

	// Create database if not exists
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.Name)
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create database %s: %w", cfg.Name, err)
	}

	appLogger.Info("Database ensured", zap.String("database", cfg.Name), zap.String("host", cfg.Host), zap.Int("port", cfg.Port))
	return nil
}
