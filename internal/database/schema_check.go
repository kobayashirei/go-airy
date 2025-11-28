package database

import (
	"github.com/kobayashirei/airy/internal/models"
	"gorm.io/gorm"
)

type SchemaStatus struct {
	OK             bool
	MissingTables  []string
	MissingColumns map[string][]string
}

func CheckSchema() (*SchemaStatus, error) {
	db := GetDB()
	if db == nil {
		return nil, gorm.ErrInvalidDB
	}

	migrator := db.Migrator()
	status := &SchemaStatus{OK: true, MissingColumns: map[string][]string{}}

	tables := map[string][]string{
		models.User{}.TableName():           {"id", "username", "password_hash"},
		models.UserProfile{}.TableName():    {"user_id"},
		models.UserStats{}.TableName():      {"user_id"},
		models.Role{}.TableName():           {"id", "name"},
		models.Permission{}.TableName():     {"id", "name"},
		models.RolePermission{}.TableName(): {"role_id", "permission_id"},
		models.UserRole{}.TableName():       {"user_id", "role_id"},
		models.Circle{}.TableName():         {"id", "name", "creator_id"},
		models.CircleMember{}.TableName():   {"id", "circle_id", "user_id"},
		models.Post{}.TableName():           {"id", "author_id"},
		models.Comment{}.TableName():        {"id", "author_id", "post_id"},
		models.Vote{}.TableName():           {"id", "user_id", "entity_type", "entity_id"},
		models.Favorite{}.TableName():       {"id", "user_id", "post_id"},
		models.EntityCount{}.TableName():    {"entity_type", "entity_id"},
		models.Notification{}.TableName():   {"id", "receiver_id"},
		models.Conversation{}.TableName():   {"id", "user1_id", "user2_id"},
		models.Message{}.TableName():        {"id", "conversation_id", "sender_id"},
		models.AdminLog{}.TableName():       {"id", "operator_id"},
	}

	for table, cols := range tables {
		if !migrator.HasTable(table) {
			status.MissingTables = append(status.MissingTables, table)
			status.OK = false
			continue
		}
		for _, c := range cols {
			if !migrator.HasColumn(table, c) {
				status.MissingColumns[table] = append(status.MissingColumns[table], c)
				status.OK = false
			}
		}
	}

	return status, nil
}
