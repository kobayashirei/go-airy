package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"

	"github.com/kobayashirei/airy/internal/config"
	appLogger "github.com/kobayashirei/airy/internal/logger"
)

// RunMigrations runs database migrations
func RunMigrations(cfg *config.DatabaseConfig, migrationsPath string) error {
	// Open database connection
	dsn := cfg.GetDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create MySQL driver instance
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	appLogger.Info("Database migrations completed",
		zap.Uint("version", version),
		zap.Bool("dirty", dirty),
	)

	return nil
}

// RollbackMigration rolls back the last migration
func RollbackMigration(cfg *config.DatabaseConfig, migrationsPath string) error {
	// Open database connection
	dsn := cfg.GetDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create MySQL driver instance
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Rollback one step
	if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	appLogger.Info("Migration rolled back",
		zap.Uint("version", version),
		zap.Bool("dirty", dirty),
	)

	return nil
}

// MigrationVersion returns the current migration version
func MigrationVersion(cfg *config.DatabaseConfig, migrationsPath string) (uint, bool, error) {
	// Open database connection
	dsn := cfg.GetDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return 0, false, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create MySQL driver instance
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"mysql",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}
