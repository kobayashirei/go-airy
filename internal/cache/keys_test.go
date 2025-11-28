package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyGenerator_UserKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.UserKey(123)
	assert.Equal(t, "user:123", key)
}

func TestKeyGenerator_UserProfileKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.UserProfileKey(456)
	assert.Equal(t, "user_profile:456", key)
}

func TestKeyGenerator_UserStatsKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.UserStatsKey(789)
	assert.Equal(t, "user_stats:789", key)
}

func TestKeyGenerator_PostKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.PostKey(100)
	assert.Equal(t, "post:100", key)
}

func TestKeyGenerator_CommentKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.CommentKey(200)
	assert.Equal(t, "comment:200", key)
}

func TestKeyGenerator_CircleKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.CircleKey(300)
	assert.Equal(t, "circle:300", key)
}

func TestKeyGenerator_PostCountKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.PostCountKey(100)
	assert.Equal(t, "count:post:100", key)
}

func TestKeyGenerator_CommentCountKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.CommentCountKey(200)
	assert.Equal(t, "count:comment:200", key)
}

func TestKeyGenerator_UserFeedKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.UserFeedKey(123)
	assert.Equal(t, "timeline:user:123", key)
}

func TestKeyGenerator_CircleFeedKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.CircleFeedKey(300)
	assert.Equal(t, "timeline:circle:300", key)
}

func TestKeyGenerator_SessionKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.SessionKey("abc123token")
	assert.Equal(t, "session:abc123token", key)
}

func TestKeyGenerator_VerificationCodeKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.VerificationCodeKey("user@example.com")
	assert.Equal(t, "code:user@example.com", key)
}

func TestKeyGenerator_ActivationTokenKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.ActivationTokenKey("activation123")
	assert.Equal(t, "token:activation:activation123", key)
}

func TestKeyGenerator_NotificationListKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.NotificationListKey(123)
	assert.Equal(t, "notification:user:123", key)
}

func TestKeyGenerator_ConversationKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.ConversationKey(500)
	assert.Equal(t, "conversation:500", key)
}

func TestKeyGenerator_ConversationListKey(t *testing.T) {
	kg := NewKeyGenerator()
	key := kg.ConversationListKey(123)
	assert.Equal(t, "conversation:user:123", key)
}
