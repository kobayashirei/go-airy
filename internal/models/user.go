package models

import "time"

// User represents a user in the system
type User struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string    `gorm:"uniqueIndex;size:100" json:"email"`
	Phone        string    `gorm:"uniqueIndex;size:20" json:"phone"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Avatar       string    `gorm:"size:255" json:"avatar"`
	Gender       string    `gorm:"size:10" json:"gender"`
	Birthday     time.Time `json:"birthday"`
	Bio          string    `gorm:"size:500" json:"bio"`
	Status       string    `gorm:"size:20;default:'inactive';index" json:"status"` // active, inactive, banned
	LastLoginAt  time.Time `json:"last_login_at"`
	LastLoginIP  string    `gorm:"size:45" json:"last_login_ip"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// UserProfile represents a user's profile information
type UserProfile struct {
	UserID         int64     `gorm:"primaryKey" json:"user_id"`
	Points         int       `gorm:"default:0" json:"points"`
	Level          int       `gorm:"default:1" json:"level"`
	FollowerCount  int       `gorm:"default:0" json:"follower_count"`
	FollowingCount int       `gorm:"default:0" json:"following_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName specifies the table name for UserProfile model
func (UserProfile) TableName() string {
	return "user_profiles"
}

// UserStats represents user statistics
type UserStats struct {
	UserID            int64     `gorm:"primaryKey" json:"user_id"`
	PostCount         int       `gorm:"default:0" json:"post_count"`
	CommentCount      int       `gorm:"default:0" json:"comment_count"`
	VoteReceivedCount int       `gorm:"default:0" json:"vote_received_count"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TableName specifies the table name for UserStats model
func (UserStats) TableName() string {
	return "user_stats"
}
