package models

import "time"

// Circle represents a community circle/group
type Circle struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Avatar      string    `gorm:"size:255" json:"avatar"`
	Background  string    `gorm:"size:255" json:"background"`
	CreatorID   int64     `gorm:"index;not null" json:"creator_id"`
	Status      string    `gorm:"size:20;default:'public'" json:"status"` // public, semi_public, private
	JoinRule    string    `gorm:"size:20;default:'free'" json:"join_rule"` // free, approval
	MemberCount int       `gorm:"default:0" json:"member_count"`
	PostCount   int       `gorm:"default:0" json:"post_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for Circle model
func (Circle) TableName() string {
	return "circles"
}

// CircleMember represents a member of a circle
type CircleMember struct {
	ID       int64     `gorm:"primaryKey" json:"id"`
	CircleID int64     `gorm:"uniqueIndex:idx_circle_user;index;not null" json:"circle_id"`
	UserID   int64     `gorm:"uniqueIndex:idx_circle_user;index;not null" json:"user_id"`
	Role     string    `gorm:"size:20;default:'member'" json:"role"` // member, moderator, pending
	JoinedAt time.Time `json:"joined_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for CircleMember model
func (CircleMember) TableName() string {
	return "circle_members"
}
