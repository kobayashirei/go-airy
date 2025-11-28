package models

import "gorm.io/gorm"

// AllModels returns a slice of all model types for migration
func AllModels() []interface{} {
	return []interface{}{
		// User models
		&User{},
		&UserProfile{},
		&UserStats{},

		// Permission models
		&Role{},
		&Permission{},
		&RolePermission{},
		&UserRole{},

		// Content models
		&Post{},
		&Comment{},
		&Vote{},
		&Favorite{},
		&EntityCount{},

		// Circle models
		&Circle{},
		&CircleMember{},

		// Notification models
		&Notification{},
		&Conversation{},
		&Message{},

		// Admin models
		&AdminLog{},
	}
}

// AutoMigrate runs auto migration for all models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(AllModels()...)
}
