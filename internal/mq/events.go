package mq

import "time"

// Event topics (routing keys)
const (
	// Post events
	TopicPostPublished = "post.published"
	TopicPostUpdated   = "post.updated"
	TopicPostDeleted   = "post.deleted"
	TopicPostVoted     = "post.voted"

	// Comment events
	TopicCommentCreated = "comment.created"
	TopicCommentDeleted = "comment.deleted"
	TopicCommentVoted   = "comment.voted"

	// User events
	TopicUserFollowed   = "user.followed"
	TopicUserUnfollowed = "user.unfollowed"
	TopicUserRegistered = "user.registered"

	// Circle events
	TopicCircleJoined = "circle.joined"
	TopicCircleLeft   = "circle.left"

	// Vote events
	TopicVoteCreated = "vote.created"
	TopicVoteUpdated = "vote.updated"
	TopicVoteDeleted = "vote.deleted"
)

// BaseEvent contains common fields for all events
type BaseEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}

// PostPublishedEvent is published when a post is published
type PostPublishedEvent struct {
	BaseEvent
	PostID   int64  `json:"post_id"`
	AuthorID int64  `json:"author_id"`
	CircleID *int64 `json:"circle_id,omitempty"`
	Title    string `json:"title"`
}

// PostUpdatedEvent is published when a post is updated
type PostUpdatedEvent struct {
	BaseEvent
	PostID   int64 `json:"post_id"`
	AuthorID int64 `json:"author_id"`
}

// PostDeletedEvent is published when a post is deleted
type PostDeletedEvent struct {
	BaseEvent
	PostID   int64 `json:"post_id"`
	AuthorID int64 `json:"author_id"`
}

// PostVotedEvent is published when a post receives a vote
type PostVotedEvent struct {
	BaseEvent
	PostID   int64  `json:"post_id"`
	AuthorID int64  `json:"author_id"`
	VoterID  int64  `json:"voter_id"`
	VoteType string `json:"vote_type"` // "up" or "down"
}

// CommentCreatedEvent is published when a comment is created
type CommentCreatedEvent struct {
	BaseEvent
	CommentID int64  `json:"comment_id"`
	PostID    int64  `json:"post_id"`
	AuthorID  int64  `json:"author_id"`
	ParentID  *int64 `json:"parent_id,omitempty"`
	Content   string `json:"content"`
}

// CommentDeletedEvent is published when a comment is deleted
type CommentDeletedEvent struct {
	BaseEvent
	CommentID int64 `json:"comment_id"`
	PostID    int64 `json:"post_id"`
	AuthorID  int64 `json:"author_id"`
}

// CommentVotedEvent is published when a comment receives a vote
type CommentVotedEvent struct {
	BaseEvent
	CommentID int64  `json:"comment_id"`
	PostID    int64  `json:"post_id"`
	AuthorID  int64  `json:"author_id"`
	VoterID   int64  `json:"voter_id"`
	VoteType  string `json:"vote_type"` // "up" or "down"
}

// UserFollowedEvent is published when a user follows another user
type UserFollowedEvent struct {
	BaseEvent
	FollowerID  int64 `json:"follower_id"`
	FollowingID int64 `json:"following_id"`
}

// UserUnfollowedEvent is published when a user unfollows another user
type UserUnfollowedEvent struct {
	BaseEvent
	FollowerID  int64 `json:"follower_id"`
	FollowingID int64 `json:"following_id"`
}

// UserRegisteredEvent is published when a new user registers
type UserRegisteredEvent struct {
	BaseEvent
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// CircleJoinedEvent is published when a user joins a circle
type CircleJoinedEvent struct {
	BaseEvent
	CircleID int64 `json:"circle_id"`
	UserID   int64 `json:"user_id"`
}

// CircleLeftEvent is published when a user leaves a circle
type CircleLeftEvent struct {
	BaseEvent
	CircleID int64 `json:"circle_id"`
	UserID   int64 `json:"user_id"`
}

// VoteCreatedEvent is published when a vote is created
type VoteCreatedEvent struct {
	BaseEvent
	VoteID     int64  `json:"vote_id"`
	UserID     int64  `json:"user_id"`
	EntityType string `json:"entity_type"` // "post" or "comment"
	EntityID   int64  `json:"entity_id"`
	VoteType   string `json:"vote_type"` // "up" or "down"
}

// VoteUpdatedEvent is published when a vote is updated
type VoteUpdatedEvent struct {
	BaseEvent
	VoteID      int64  `json:"vote_id"`
	UserID      int64  `json:"user_id"`
	EntityType  string `json:"entity_type"` // "post" or "comment"
	EntityID    int64  `json:"entity_id"`
	OldVoteType string `json:"old_vote_type"` // "up" or "down"
	NewVoteType string `json:"new_vote_type"` // "up" or "down"
}

// VoteDeletedEvent is published when a vote is deleted
type VoteDeletedEvent struct {
	BaseEvent
	VoteID     int64  `json:"vote_id"`
	UserID     int64  `json:"user_id"`
	EntityType string `json:"entity_type"` // "post" or "comment"
	EntityID   int64  `json:"entity_id"`
	VoteType   string `json:"vote_type"` // "up" or "down"
}
