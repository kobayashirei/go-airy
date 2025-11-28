package cache

import "fmt"

// Key prefixes for different entity types
const (
	PrefixUser         = "user"
	PrefixUserProfile  = "user_profile"
	PrefixUserStats    = "user_stats"
	PrefixPost         = "post"
	PrefixComment      = "comment"
	PrefixCircle       = "circle"
	PrefixCount        = "count"
	PrefixFeed         = "timeline"
	PrefixSession      = "session"
	PrefixCode         = "code"
	PrefixToken        = "token"
	PrefixNotification = "notification"
	PrefixConversation = "conversation"
)

// KeyGenerator provides methods to generate cache keys
type KeyGenerator struct{}

// NewKeyGenerator creates a new key generator instance
func NewKeyGenerator() *KeyGenerator {
	return &KeyGenerator{}
}

// UserKey generates a cache key for user data
// Format: user:{id}
func (kg *KeyGenerator) UserKey(userID int64) string {
	return fmt.Sprintf("%s:%d", PrefixUser, userID)
}

// UserProfileKey generates a cache key for user profile data
// Format: user_profile:{id}
func (kg *KeyGenerator) UserProfileKey(userID int64) string {
	return fmt.Sprintf("%s:%d", PrefixUserProfile, userID)
}

// UserStatsKey generates a cache key for user statistics
// Format: user_stats:{id}
func (kg *KeyGenerator) UserStatsKey(userID int64) string {
	return fmt.Sprintf("%s:%d", PrefixUserStats, userID)
}

// PostKey generates a cache key for post data
// Format: post:{id}
func (kg *KeyGenerator) PostKey(postID int64) string {
	return fmt.Sprintf("%s:%d", PrefixPost, postID)
}

// CommentKey generates a cache key for comment data
// Format: comment:{id}
func (kg *KeyGenerator) CommentKey(commentID int64) string {
	return fmt.Sprintf("%s:%d", PrefixComment, commentID)
}

// CircleKey generates a cache key for circle data
// Format: circle:{id}
func (kg *KeyGenerator) CircleKey(circleID int64) string {
	return fmt.Sprintf("%s:%d", PrefixCircle, circleID)
}

// PostCountKey generates a cache key for post count data
// Format: count:post:{id}
func (kg *KeyGenerator) PostCountKey(postID int64) string {
	return fmt.Sprintf("%s:%s:%d", PrefixCount, PrefixPost, postID)
}

// CommentCountKey generates a cache key for comment count data
// Format: count:comment:{id}
func (kg *KeyGenerator) CommentCountKey(commentID int64) string {
	return fmt.Sprintf("%s:%s:%d", PrefixCount, PrefixComment, commentID)
}

// UserFeedKey generates a cache key for user feed/timeline
// Format: timeline:user:{id}
func (kg *KeyGenerator) UserFeedKey(userID int64) string {
	return fmt.Sprintf("%s:%s:%d", PrefixFeed, PrefixUser, userID)
}

// CircleFeedKey generates a cache key for circle feed
// Format: timeline:circle:{id}
func (kg *KeyGenerator) CircleFeedKey(circleID int64) string {
	return fmt.Sprintf("%s:%s:%d", PrefixFeed, PrefixCircle, circleID)
}

// SessionKey generates a cache key for session token
// Format: session:{token}
func (kg *KeyGenerator) SessionKey(token string) string {
	return fmt.Sprintf("%s:%s", PrefixSession, token)
}

// VerificationCodeKey generates a cache key for verification code
// Format: code:{phone/email}
func (kg *KeyGenerator) VerificationCodeKey(identifier string) string {
	return fmt.Sprintf("%s:%s", PrefixCode, identifier)
}

// ActivationTokenKey generates a cache key for activation token
// Format: token:activation:{token}
func (kg *KeyGenerator) ActivationTokenKey(token string) string {
	return fmt.Sprintf("%s:activation:%s", PrefixToken, token)
}

// NotificationListKey generates a cache key for user notification list
// Format: notification:user:{id}
func (kg *KeyGenerator) NotificationListKey(userID int64) string {
	return fmt.Sprintf("%s:%s:%d", PrefixNotification, PrefixUser, userID)
}

// ConversationKey generates a cache key for conversation data
// Format: conversation:{id}
func (kg *KeyGenerator) ConversationKey(conversationID int64) string {
	return fmt.Sprintf("%s:%d", PrefixConversation, conversationID)
}

// ConversationListKey generates a cache key for user conversation list
// Format: conversation:user:{id}
func (kg *KeyGenerator) ConversationListKey(userID int64) string {
	return fmt.Sprintf("%s:%s:%d", PrefixConversation, PrefixUser, userID)
}

// Standalone key generation functions for convenience

// UserKey generates a cache key for user data
func UserKey(userID int64) string {
	return fmt.Sprintf("%s:%d", PrefixUser, userID)
}

// ActivationTokenKey generates a cache key for activation token
func ActivationTokenKey(token string) string {
	return fmt.Sprintf("%s:activation:%s", PrefixToken, token)
}

// VerificationCodeKey generates a cache key for verification code
func VerificationCodeKey(identifier string) string {
	return fmt.Sprintf("%s:%s", PrefixCode, identifier)
}

// PostKey generates a cache key for post data
func PostKey(postID int64) string {
	return fmt.Sprintf("%s:%d", PrefixPost, postID)
}

// GetUserFeedKey generates a cache key for user feed/timeline
func GetUserFeedKey(userID int64) string {
	return fmt.Sprintf("%s:%s:%d", PrefixFeed, PrefixUser, userID)
}

// GetCircleFeedKey generates a cache key for circle feed
func GetCircleFeedKey(circleID int64) string {
	return fmt.Sprintf("%s:%s:%d", PrefixFeed, PrefixCircle, circleID)
}

// GetPostKey generates a cache key for post data
func GetPostKey(postID int64) string {
	return fmt.Sprintf("%s:%d", PrefixPost, postID)
}
