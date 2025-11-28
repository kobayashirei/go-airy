package models

import "time"

// Role represents a user role in the system
type Role struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:50;not null" json:"name"` // super_admin, moderator, user
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for Role model
func (Role) TableName() string {
	return "roles"
}

// Permission represents a permission in the system
type Permission struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:100;not null" json:"name"` // post:create, user:ban
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for Permission model
func (Permission) TableName() string {
	return "permissions"
}

// RolePermission represents the association between roles and permissions
type RolePermission struct {
	RoleID       int64     `gorm:"primaryKey" json:"role_id"`
	PermissionID int64     `gorm:"primaryKey" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName specifies the table name for RolePermission model
func (RolePermission) TableName() string {
	return "role_permissions"
}

// UserRole represents the association between users and roles
type UserRole struct {
	UserID    int64     `gorm:"primaryKey" json:"user_id"`
	RoleID    int64     `gorm:"primaryKey" json:"role_id"`
	CircleID  *int64    `gorm:"primaryKey" json:"circle_id"` // Optional, for circle-specific roles
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for UserRole model
func (UserRole) TableName() string {
	return "user_roles"
}
