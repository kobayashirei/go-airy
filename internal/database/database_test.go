package database

import (
	"context"
	"testing"
	"time"

	"github.com/kobayashirei/airy/internal/config"
	appLogger "github.com/kobayashirei/airy/internal/logger"
)

func TestHealthCheck(t *testing.T) {
	// Test when DB is not initialized
	DB = nil
	ctx := context.Background()
	err := HealthCheck(ctx)
	if err == nil {
		t.Error("Expected error when DB is not initialized, got nil")
	}
}

func TestGetDB(t *testing.T) {
	// Test GetDB returns the global DB instance
	DB = nil
	db := GetDB()
	if db != nil {
		t.Error("Expected nil when DB is not initialized")
	}
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		Name:     "testdb",
		User:     "testuser",
		Password: "testpass",
	}

	expected := "testuser:testpass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	actual := cfg.GetDSN()

	if actual != expected {
		t.Errorf("Expected DSN %s, got %s", expected, actual)
	}
}

func TestGormLogWriter(t *testing.T) {
	// Initialize logger for test
	cfg := &config.LogConfig{
		Level:  "info",
		Output: "stdout",
	}
	if err := appLogger.Init(cfg); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	writer := &gormLogWriter{}
	// Test that Printf doesn't panic
	writer.Printf("test log: %s", "message")
}

// Integration test - only runs if database is available
func TestInit_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            3306,
		Name:            "airygithub_test",
		User:            "root",
		Password:        "",
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: 1 * time.Hour,
	}

	// This will fail if database is not available, which is expected
	err := Init(cfg)
	if err != nil {
		t.Logf("Database not available (expected in CI): %v", err)
		return
	}

	// If database is available, test health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := HealthCheck(ctx); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// Clean up
	if err := Close(); err != nil {
		t.Errorf("Failed to close database: %v", err)
	}
}
