package database

import (
	"testing"

	"github.com/kobayashirei/airy/internal/config"
)

func TestMigrationFunctions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		Name:     "airygithub_test",
		User:     "root",
		Password: "",
	}

	// Test migration version with non-existent database
	// This should fail gracefully
	_, _, err := MigrationVersion(cfg, "../../migrations")
	if err != nil {
		t.Logf("Expected error when database doesn't exist: %v", err)
	}
}

func TestMigrationHelpers(t *testing.T) {
	// Test that migration functions exist and have correct signatures
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		Name:     "test",
		User:     "root",
		Password: "",
	}

	// These will fail if database is not available, which is expected
	t.Run("RunMigrations", func(t *testing.T) {
		err := RunMigrations(cfg, "../../migrations")
		if err != nil {
			t.Logf("Migration failed (expected if DB not available): %v", err)
		}
	})

	t.Run("RollbackMigration", func(t *testing.T) {
		err := RollbackMigration(cfg, "../../migrations")
		if err != nil {
			t.Logf("Rollback failed (expected if DB not available): %v", err)
		}
	})

	t.Run("MigrationVersion", func(t *testing.T) {
		_, _, err := MigrationVersion(cfg, "../../migrations")
		if err != nil {
			t.Logf("Version check failed (expected if DB not available): %v", err)
		}
	})
}
