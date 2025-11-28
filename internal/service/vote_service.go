package service

import (
	"context"
	"fmt"
	"time"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/mq"
	"github.com/kobayashirei/airy/internal/repository"
)

// VotePublisher defines the interface for publishing vote events
type VotePublisher interface {
	PublishVoteCreated(ctx context.Context, voteID, userID int64, entityType string, entityID int64, voteType string) error
	PublishVoteUpdated(ctx context.Context, voteID, userID int64, entityType string, entityID int64, oldVoteType, newVoteType string) error
	PublishVoteDeleted(ctx context.Context, voteID, userID int64, entityType string, entityID int64, voteType string) error
}

// VoteService defines the interface for vote business logic
type VoteService interface {
	// Vote creates or updates a vote (idempotent operation)
	Vote(ctx context.Context, userID int64, entityType string, entityID int64, voteType string) (*models.Vote, error)
	// CancelVote removes a user's vote
	CancelVote(ctx context.Context, userID int64, entityType string, entityID int64) error
	// GetVote retrieves a user's vote for an entity
	GetVote(ctx context.Context, userID int64, entityType string, entityID int64) (*models.Vote, error)
}

// voteService implements VoteService interface
type voteService struct {
	voteRepo    repository.VoteRepository
	postRepo    repository.PostRepository
	commentRepo repository.CommentRepository
	publisher   VotePublisher
}

// NewVoteService creates a new vote service
func NewVoteService(
	voteRepo repository.VoteRepository,
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
	publisher VotePublisher,
) VoteService {
	return &voteService{
		voteRepo:    voteRepo,
		postRepo:    postRepo,
		commentRepo: commentRepo,
		publisher:   publisher,
	}
}

// Vote creates or updates a vote (idempotent operation)
func (s *voteService) Vote(ctx context.Context, userID int64, entityType string, entityID int64, voteType string) (*models.Vote, error) {
	// Validate entity type
	if entityType != "post" && entityType != "comment" {
		return nil, fmt.Errorf("invalid entity type: %s", entityType)
	}

	// Validate vote type
	if voteType != "up" && voteType != "down" {
		return nil, fmt.Errorf("invalid vote type: %s", voteType)
	}

	// Verify entity exists and get author ID
	var authorID int64
	var err error
	
	if entityType == "post" {
		post, err := s.postRepo.FindByID(ctx, entityID)
		if err != nil {
			return nil, fmt.Errorf("failed to find post: %w", err)
		}
		if post == nil {
			return nil, fmt.Errorf("post not found: %d", entityID)
		}
		authorID = post.AuthorID
	} else if entityType == "comment" {
		comment, err := s.commentRepo.FindByID(ctx, entityID)
		if err != nil {
			return nil, fmt.Errorf("failed to find comment: %w", err)
		}
		if comment == nil {
			return nil, fmt.Errorf("comment not found: %d", entityID)
		}
		authorID = comment.AuthorID
	}

	// Check if vote already exists
	existingVote, err := s.voteRepo.FindByUserAndEntity(ctx, userID, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing vote: %w", err)
	}

	var vote *models.Vote
	var eventType string
	var oldVoteType string

	if existingVote != nil {
		// Vote exists - update it (idempotent)
		if existingVote.VoteType == voteType {
			// Same vote type - return existing vote (idempotent)
			return existingVote, nil
		}

		// Different vote type - update
		oldVoteType = existingVote.VoteType
		existingVote.VoteType = voteType
		existingVote.UpdatedAt = time.Now()

		if err := s.voteRepo.Update(ctx, existingVote); err != nil {
			return nil, fmt.Errorf("failed to update vote: %w", err)
		}

		vote = existingVote
		eventType = mq.TopicVoteUpdated
	} else {
		// Create new vote
		vote = &models.Vote{
			UserID:     userID,
			EntityType: entityType,
			EntityID:   entityID,
			VoteType:   voteType,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := s.voteRepo.Create(ctx, vote); err != nil {
			return nil, fmt.Errorf("failed to create vote: %w", err)
		}

		eventType = mq.TopicVoteCreated
	}

	// Publish vote event to message queue for async processing
	if err := s.publishVoteEvent(ctx, vote, eventType, oldVoteType, authorID); err != nil {
		// Log error but don't fail the request
		// The vote is already saved, event publishing is best-effort
		fmt.Printf("failed to publish vote event: %v\n", err)
	}

	return vote, nil
}

// CancelVote removes a user's vote
func (s *voteService) CancelVote(ctx context.Context, userID int64, entityType string, entityID int64) error {
	// Validate entity type
	if entityType != "post" && entityType != "comment" {
		return fmt.Errorf("invalid entity type: %s", entityType)
	}

	// Check if vote exists
	existingVote, err := s.voteRepo.FindByUserAndEntity(ctx, userID, entityType, entityID)
	if err != nil {
		return fmt.Errorf("failed to check existing vote: %w", err)
	}

	if existingVote == nil {
		// No vote to cancel - idempotent
		return nil
	}

	// Get author ID for notification
	var authorID int64
	if entityType == "post" {
		post, err := s.postRepo.FindByID(ctx, entityID)
		if err != nil {
			return fmt.Errorf("failed to find post: %w", err)
		}
		if post != nil {
			authorID = post.AuthorID
		}
	} else if entityType == "comment" {
		comment, err := s.commentRepo.FindByID(ctx, entityID)
		if err != nil {
			return fmt.Errorf("failed to find comment: %w", err)
		}
		if comment != nil {
			authorID = comment.AuthorID
		}
	}

	// Delete the vote
	if err := s.voteRepo.DeleteByUserAndEntity(ctx, userID, entityType, entityID); err != nil {
		return fmt.Errorf("failed to delete vote: %w", err)
	}

	// Publish vote deleted event
	if err := s.publishVoteDeletedEvent(ctx, existingVote, authorID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("failed to publish vote deleted event: %v\n", err)
	}

	return nil
}

// GetVote retrieves a user's vote for an entity
func (s *voteService) GetVote(ctx context.Context, userID int64, entityType string, entityID int64) (*models.Vote, error) {
	vote, err := s.voteRepo.FindByUserAndEntity(ctx, userID, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote: %w", err)
	}
	return vote, nil
}

// publishVoteEvent publishes a vote event to the message queue
func (s *voteService) publishVoteEvent(ctx context.Context, vote *models.Vote, eventType string, oldVoteType string, authorID int64) error {
	switch eventType {
	case mq.TopicVoteCreated:
		return s.publisher.PublishVoteCreated(ctx, vote.ID, vote.UserID, vote.EntityType, vote.EntityID, vote.VoteType)
	case mq.TopicVoteUpdated:
		return s.publisher.PublishVoteUpdated(ctx, vote.ID, vote.UserID, vote.EntityType, vote.EntityID, oldVoteType, vote.VoteType)
	default:
		return fmt.Errorf("unknown event type: %s", eventType)
	}
}

// publishVoteDeletedEvent publishes a vote deleted event
func (s *voteService) publishVoteDeletedEvent(ctx context.Context, vote *models.Vote, authorID int64) error {
	return s.publisher.PublishVoteDeleted(ctx, vote.ID, vote.UserID, vote.EntityType, vote.EntityID, vote.VoteType)
}
