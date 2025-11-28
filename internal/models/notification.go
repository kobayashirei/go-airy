package models

import "time"

// Notification represents a notification to a user
type Notification struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	ReceiverID    int64     `gorm:"index;not null" json:"receiver_id"`
	TriggerUserID *int64    `json:"trigger_user_id"`
	Type          string    `gorm:"size:20;not null" json:"type"` // comment, vote, mention, system
	EntityType    string    `gorm:"size:20" json:"entity_type"` // post, comment
	EntityID      *int64    `json:"entity_id"`
	Content       string    `gorm:"size:500" json:"content"`
	IsRead        bool      `gorm:"default:false;index" json:"is_read"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}

// TableName specifies the table name for Notification model
func (Notification) TableName() string {
	return "notifications"
}

// Conversation represents a private conversation between two users
type Conversation struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	User1ID       int64     `gorm:"uniqueIndex:idx_users;not null" json:"user1_id"`
	User2ID       int64     `gorm:"uniqueIndex:idx_users;not null" json:"user2_id"`
	LastMessageAt time.Time `gorm:"index" json:"last_message_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// TableName specifies the table name for Conversation model
func (Conversation) TableName() string {
	return "conversations"
}

// Message represents a message in a conversation
type Message struct {
	ID             int64     `gorm:"primaryKey" json:"id"`
	ConversationID int64     `gorm:"index;not null" json:"conversation_id"`
	SenderID       int64     `gorm:"index;not null" json:"sender_id"`
	ContentType    string    `gorm:"size:20;default:'text'" json:"content_type"` // text, image
	Content        string    `gorm:"type:text;not null" json:"content"`
	IsRead         bool      `gorm:"default:false" json:"is_read"`
	CreatedAt      time.Time `gorm:"index" json:"created_at"`
}

// TableName specifies the table name for Message model
func (Message) TableName() string {
	return "messages"
}
