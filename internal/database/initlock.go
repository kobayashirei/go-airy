package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/kobayashirei/airy/internal/logger"
)

type initLockInfo struct {
	CreatedAt string `json:"created_at"`
	Hostname  string `json:"hostname"`
	PID       int    `json:"pid"`
}

// HasInitLock checks whether the init lock file exists
func HasInitLock(lockPath string) (bool, error) {
	if lockPath == "" {
		return false, nil
	}
	_, err := os.Stat(lockPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to stat init lock: %w", err)
}

// CreateInitLock writes the init lock file with metadata
func CreateInitLock(lockPath string) error {
	if lockPath == "" {
		return nil
	}

	dir := filepath.Dir(lockPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create lock dir: %w", err)
	}

	hostname, _ := os.Hostname()
	info := initLockInfo{
		CreatedAt: time.Now().Format(time.RFC3339),
		Hostname:  hostname,
		PID:       os.Getpid(),
	}
	data, _ := json.MarshalIndent(info, "", "  ")

	// Write atomically
	tmp := lockPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp lock file: %w", err)
	}
	if err := os.Rename(tmp, lockPath); err != nil {
		return fmt.Errorf("failed to rename temp lock file: %w", err)
	}

	logger.Info("Init lock created", zap.String("path", lockPath))
	return nil
}
