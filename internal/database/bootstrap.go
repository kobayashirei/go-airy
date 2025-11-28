package database

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/kobayashirei/airy/internal/config"
	appLogger "github.com/kobayashirei/airy/internal/logger"
	"github.com/kobayashirei/airy/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func BootstrapAdmin(cfg *config.Config) error {
	db := GetDB()
	if db == nil {
		return gorm.ErrInvalidDB
	}

	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}
	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		email = "admin@example.com"
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = uuid.NewString()
	}

	var existing models.User
	if err := db.Where("email = ? OR username = ?", email, username).First(&existing).Error; err == nil {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.Security.BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	user := models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	profile := models.UserProfile{UserID: user.ID, CreatedAt: now, UpdatedAt: now}
	_ = db.FirstOrCreate(&profile, models.UserProfile{UserID: user.ID}).Error
	stats := models.UserStats{UserID: user.ID, CreatedAt: now, UpdatedAt: now}
	_ = db.FirstOrCreate(&stats, models.UserStats{UserID: user.ID}).Error

	var role models.Role
	if err := db.Where("name = ?", "super_admin").First(&role).Error; err != nil {
		role = models.Role{Name: "super_admin", CreatedAt: now, UpdatedAt: now}
		if err := db.Create(&role).Error; err != nil {
			return fmt.Errorf("failed to create role: %w", err)
		}
	}

	var ur models.UserRole
	if err := db.Where("user_id = ? AND role_id = ? AND circle_id IS NULL", user.ID, role.ID).First(&ur).Error; err != nil {
		ur = models.UserRole{UserID: user.ID, RoleID: role.ID, CreatedAt: now}
		if err := db.Create(&ur).Error; err != nil {
			return fmt.Errorf("failed to assign admin role: %w", err)
		}
	}

	appLogger.Info("Admin bootstrap completed")
	return nil
}
