package models

import (
	"testing"
	"time"
)

func TestUserTableName(t *testing.T) {
	user := User{}
	if user.TableName() != "users" {
		t.Errorf("Expected table name 'users', got '%s'", user.TableName())
	}
}

func TestUserProfileTableName(t *testing.T) {
	profile := UserProfile{}
	if profile.TableName() != "user_profiles" {
		t.Errorf("Expected table name 'user_profiles', got '%s'", profile.TableName())
	}
}

func TestUserStatsTableName(t *testing.T) {
	stats := UserStats{}
	if stats.TableName() != "user_stats" {
		t.Errorf("Expected table name 'user_stats', got '%s'", stats.TableName())
	}
}

func TestRoleTableName(t *testing.T) {
	role := Role{}
	if role.TableName() != "roles" {
		t.Errorf("Expected table name 'roles', got '%s'", role.TableName())
	}
}

func TestPermissionTableName(t *testing.T) {
	permission := Permission{}
	if permission.TableName() != "permissions" {
		t.Errorf("Expected table name 'permissions', got '%s'", permission.TableName())
	}
}

func TestPostTableName(t *testing.T) {
	post := Post{}
	if post.TableName() != "posts" {
		t.Errorf("Expected table name 'posts', got '%s'", post.TableName())
	}
}

func TestCommentTableName(t *testing.T) {
	comment := Comment{}
	if comment.TableName() != "comments" {
		t.Errorf("Expected table name 'comments', got '%s'", comment.TableName())
	}
}

func TestVoteTableName(t *testing.T) {
	vote := Vote{}
	if vote.TableName() != "votes" {
		t.Errorf("Expected table name 'votes', got '%s'", vote.TableName())
	}
}

func TestCircleTableName(t *testing.T) {
	circle := Circle{}
	if circle.TableName() != "circles" {
		t.Errorf("Expected table name 'circles', got '%s'", circle.TableName())
	}
}

func TestNotificationTableName(t *testing.T) {
	notification := Notification{}
	if notification.TableName() != "notifications" {
		t.Errorf("Expected table name 'notifications', got '%s'", notification.TableName())
	}
}

func TestAdminLogTableName(t *testing.T) {
	log := AdminLog{}
	if log.TableName() != "admin_logs" {
		t.Errorf("Expected table name 'admin_logs', got '%s'", log.TableName())
	}
}

func TestAllModels(t *testing.T) {
	models := AllModels()
	
	// Check that we have all expected models
	// User: 3, Permission: 4, Content: 5, Circle: 2, Notification: 3, Admin: 1 = 18 total
	expectedCount := 18
	if len(models) != expectedCount {
		t.Errorf("Expected %d models, got %d", expectedCount, len(models))
	}

	// Verify each model is a pointer
	for i, model := range models {
		if model == nil {
			t.Errorf("Model at index %d is nil", i)
		}
	}
}

func TestUserModel(t *testing.T) {
	user := User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		Phone:        "1234567890",
		PasswordHash: "hashedpassword",
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if user.ID != 1 {
		t.Errorf("Expected ID 1, got %d", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	if user.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", user.Status)
	}
}

func TestPostModel(t *testing.T) {
	post := Post{
		ID:              1,
		Title:           "Test Post",
		ContentMarkdown: "# Test",
		ContentHTML:     "<h1>Test</h1>",
		AuthorID:        1,
		Status:          "published",
		ViewCount:       0,
		HotnessScore:    0.0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if post.ID != 1 {
		t.Errorf("Expected ID 1, got %d", post.ID)
	}
	if post.Title != "Test Post" {
		t.Errorf("Expected title 'Test Post', got '%s'", post.Title)
	}
	if post.Status != "published" {
		t.Errorf("Expected status 'published', got '%s'", post.Status)
	}
}

func TestCommentModel(t *testing.T) {
	comment := Comment{
		ID:       1,
		Content:  "Test comment",
		AuthorID: 1,
		PostID:   1,
		RootID:   1,
		Level:    0,
		Path:     "1",
		Status:   "published",
	}

	if comment.ID != 1 {
		t.Errorf("Expected ID 1, got %d", comment.ID)
	}
	if comment.Level != 0 {
		t.Errorf("Expected level 0, got %d", comment.Level)
	}
	if comment.Path != "1" {
		t.Errorf("Expected path '1', got '%s'", comment.Path)
	}
}

func TestVoteModel(t *testing.T) {
	vote := Vote{
		ID:         1,
		UserID:     1,
		EntityType: "post",
		EntityID:   1,
		VoteType:   "up",
	}

	if vote.EntityType != "post" {
		t.Errorf("Expected entity type 'post', got '%s'", vote.EntityType)
	}
	if vote.VoteType != "up" {
		t.Errorf("Expected vote type 'up', got '%s'", vote.VoteType)
	}
}

func TestCircleModel(t *testing.T) {
	circle := Circle{
		ID:          1,
		Name:        "Test Circle",
		CreatorID:   1,
		Status:      "public",
		JoinRule:    "free",
		MemberCount: 0,
		PostCount:   0,
	}

	if circle.Status != "public" {
		t.Errorf("Expected status 'public', got '%s'", circle.Status)
	}
	if circle.JoinRule != "free" {
		t.Errorf("Expected join rule 'free', got '%s'", circle.JoinRule)
	}
}

func TestNotificationModel(t *testing.T) {
	notification := Notification{
		ID:         1,
		ReceiverID: 1,
		Type:       "comment",
		EntityType: "post",
		Content:    "Test notification",
		IsRead:     false,
	}

	if notification.Type != "comment" {
		t.Errorf("Expected type 'comment', got '%s'", notification.Type)
	}
	if notification.IsRead {
		t.Error("Expected IsRead to be false")
	}
}
