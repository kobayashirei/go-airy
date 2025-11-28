package models

import "time"

// AdminLog represents an administrative action log
type AdminLog struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	OperatorID int64     `gorm:"index;not null" json:"operator_id"`
	Action     string    `gorm:"size:50;index;not null" json:"action"`
	EntityType string    `gorm:"size:20" json:"entity_type"`
	EntityID   *int64    `json:"entity_id"`
	IP         string    `gorm:"size:45" json:"ip"`
	Details    string    `gorm:"type:json" json:"details"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// TableName specifies the table name for AdminLog model
func (AdminLog) TableName() string {
	return "admin_logs"
}
