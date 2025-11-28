package mq

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Publisher provides helper methods for publishing events
type Publisher struct {
	mq MessageQueue
}

// NewPublisher creates a new Publisher
func NewPublisher(mq MessageQueue) *Publisher {
	return &Publisher{mq: mq}
}

// newBaseEvent creates a new BaseEvent with generated ID and timestamp
func newBaseEvent(eventType string) BaseEvent {
	return BaseEvent{
		EventID:   uuid.New().String(),
		EventType: eventType,
		Timestamp: time.Now(),
	}
}

// PublishPostPublished publishes a post published event
func (p *Publisher) PublishPostPublished(ctx context.Context, postID, authorID int64, circleID *int64, title string) error {
	event := PostPublishedEvent{
		BaseEvent: newBaseEvent(TopicPostPublished),
		PostID:    postID,
		AuthorID:  authorID,
		CircleID:  circleID,
		Title:     title,
	}
	return p.mq.Publish(ctx, TopicPostPublished, event)
}

// PublishPostUpdated publishes a post updated event
func (p *Publisher) PublishPostUpdated(ctx context.Context, postID, authorID int64) error {
	event := PostUpdatedEvent{
		BaseEvent: newBaseEvent(TopicPostUpdated),
		PostID:    postID,
		AuthorID:  authorID,
	}
	return p.mq.Publish(ctx, TopicPostUpdated, event)
}

// PublishPostDeleted publishes a post deleted event
func (p *Publisher) PublishPostDeleted(ctx context.Context, postID, authorID int64) error {
	event := PostDeletedEvent{
		BaseEvent: newBaseEvent(TopicPostDeleted),
		PostID:    postID,
		AuthorID:  authorID,
	}
	return p.mq.Publish(ctx, TopicPostDeleted, event)
}

// PublishPostVoted publishes a post voted event
func (p *Publisher) PublishPostVoted(ctx context.Context, postID, authorID, voterID int64, voteType string) error {
	event := PostVotedEvent{
		BaseEvent: newBaseEvent(TopicPostVoted),
		PostID:    postID,
		AuthorID:  authorID,
		VoterID:   voterID,
		VoteType:  voteType,
	}
	return p.mq.Publish(ctx, TopicPostVoted, event)
}

// PublishCommentCreated publishes a comment created event
func (p *Publisher) PublishCommentCreated(ctx context.Context, commentID, postID, authorID int64, parentID *int64, content string) error {
	event := CommentCreatedEvent{
		BaseEvent: newBaseEvent(TopicCommentCreated),
		CommentID: commentID,
		PostID:    postID,
		AuthorID:  authorID,
		ParentID:  parentID,
		Content:   content,
	}
	return p.mq.Publish(ctx, TopicCommentCreated, event)
}

// PublishCommentDeleted publishes a comment deleted event
func (p *Publisher) PublishCommentDeleted(ctx context.Context, commentID, postID, authorID int64) error {
	event := CommentDeletedEvent{
		BaseEvent: newBaseEvent(TopicCommentDeleted),
		CommentID: commentID,
		PostID:    postID,
		AuthorID:  authorID,
	}
	return p.mq.Publish(ctx, TopicCommentDeleted, event)
}

// PublishCommentVoted publishes a comment voted event
func (p *Publisher) PublishCommentVoted(ctx context.Context, commentID, postID, authorID, voterID int64, voteType string) error {
	event := CommentVotedEvent{
		BaseEvent: newBaseEvent(TopicCommentVoted),
		CommentID: commentID,
		PostID:    postID,
		AuthorID:  authorID,
		VoterID:   voterID,
		VoteType:  voteType,
	}
	return p.mq.Publish(ctx, TopicCommentVoted, event)
}

// PublishUserFollowed publishes a user followed event
func (p *Publisher) PublishUserFollowed(ctx context.Context, followerID, followingID int64) error {
	event := UserFollowedEvent{
		BaseEvent:   newBaseEvent(TopicUserFollowed),
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	return p.mq.Publish(ctx, TopicUserFollowed, event)
}

// PublishUserUnfollowed publishes a user unfollowed event
func (p *Publisher) PublishUserUnfollowed(ctx context.Context, followerID, followingID int64) error {
	event := UserUnfollowedEvent{
		BaseEvent:   newBaseEvent(TopicUserUnfollowed),
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	return p.mq.Publish(ctx, TopicUserUnfollowed, event)
}

// PublishUserRegistered publishes a user registered event
func (p *Publisher) PublishUserRegistered(ctx context.Context, userID int64, username, email string) error {
	event := UserRegisteredEvent{
		BaseEvent: newBaseEvent(TopicUserRegistered),
		UserID:    userID,
		Username:  username,
		Email:     email,
	}
	return p.mq.Publish(ctx, TopicUserRegistered, event)
}

// PublishCircleJoined publishes a circle joined event
func (p *Publisher) PublishCircleJoined(ctx context.Context, circleID, userID int64) error {
	event := CircleJoinedEvent{
		BaseEvent: newBaseEvent(TopicCircleJoined),
		CircleID:  circleID,
		UserID:    userID,
	}
	return p.mq.Publish(ctx, TopicCircleJoined, event)
}

// PublishCircleLeft publishes a circle left event
func (p *Publisher) PublishCircleLeft(ctx context.Context, circleID, userID int64) error {
	event := CircleLeftEvent{
		BaseEvent: newBaseEvent(TopicCircleLeft),
		CircleID:  circleID,
		UserID:    userID,
	}
	return p.mq.Publish(ctx, TopicCircleLeft, event)
}

// PublishVoteCreated publishes a vote created event
func (p *Publisher) PublishVoteCreated(ctx context.Context, voteID, userID int64, entityType string, entityID int64, voteType string) error {
	event := VoteCreatedEvent{
		BaseEvent:  newBaseEvent(TopicVoteCreated),
		VoteID:     voteID,
		UserID:     userID,
		EntityType: entityType,
		EntityID:   entityID,
		VoteType:   voteType,
	}
	return p.mq.Publish(ctx, TopicVoteCreated, event)
}

// PublishVoteUpdated publishes a vote updated event
func (p *Publisher) PublishVoteUpdated(ctx context.Context, voteID, userID int64, entityType string, entityID int64, oldVoteType, newVoteType string) error {
	event := VoteUpdatedEvent{
		BaseEvent:   newBaseEvent(TopicVoteUpdated),
		VoteID:      voteID,
		UserID:      userID,
		EntityType:  entityType,
		EntityID:    entityID,
		OldVoteType: oldVoteType,
		NewVoteType: newVoteType,
	}
	return p.mq.Publish(ctx, TopicVoteUpdated, event)
}

// PublishVoteDeleted publishes a vote deleted event
func (p *Publisher) PublishVoteDeleted(ctx context.Context, voteID, userID int64, entityType string, entityID int64, voteType string) error {
	event := VoteDeletedEvent{
		BaseEvent:  newBaseEvent(TopicVoteDeleted),
		VoteID:     voteID,
		UserID:     userID,
		EntityType: entityType,
		EntityID:   entityID,
		VoteType:   voteType,
	}
	return p.mq.Publish(ctx, TopicVoteDeleted, event)
}
