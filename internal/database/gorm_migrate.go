package database

import (
	"fmt"

	"github.com/kobayashirei/airy/internal/models"
)

// RunGormAutoMigrate runs GORM AutoMigrate on all models
func RunGormAutoMigrate() error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.AutoMigrate(
		&models.User{},
		&models.UserProfile{},
		&models.UserStats{},
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.UserRole{},
		&models.Circle{},
		&models.CircleMember{},
		&models.Post{},
		&models.Comment{},
		&models.Vote{},
		&models.Favorite{},
		&models.EntityCount{},
		&models.Notification{},
		&models.Conversation{},
		&models.Message{},
		&models.AdminLog{},
	)
}
