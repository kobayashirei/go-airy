package service

import (
	"context"
	"testing"
	"time"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVoteService_Vote_CreateNewVote(t *testing.T) {
	ctx := context.Background()
	mockVoteRepo := new(MockVoteRepository)
	mockPostRepo := new(MockPostRepository)
	mockCommentRepo := new(MockCommentRepository)
	mockPublisher := new(MockPublisher)

	service := NewVoteService(mockVoteRepo, mockPostRepo, mockCommentRepo, mockPublisher)

	userID := int64(1)
	postID := int64(100)
	voteType := "up"

	mockPostRepo.On("FindByID", ctx, postID).Return(&models.Post{
		ID:       postID,
		AuthorID: 2,
	}, nil)

	mockVoteRepo.On("FindByUserAndEntity", ctx, userID, "post", postID).Return(nil, nil)

	mockVoteRepo.On("Create", ctx, mock.MatchedBy(func(v *models.Vote) bool {
		return v.UserID == userID && v.EntityType == "post" && v.EntityID == postID && v.VoteType == voteType
	})).Return(nil)

	mockPublisher.On("PublishVoteCreated", ctx, mock.Anything, userID, "post", postID, voteType).Return(nil)

	vote, err := service.Vote(ctx, userID, "post", postID, voteType)

	assert.NoError(t, err)
	assert.NotNil(t, vote)
	assert.Equal(t, userID, vote.UserID)
	assert.Equal(t, "post", vote.EntityType)
	assert.Equal(t, postID, vote.EntityID)
	assert.Equal(t, voteType, vote.VoteType)

	mockVoteRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestVoteService_Vote_Idempotent(t *testing.T) {
	ctx := context.Background()
	mockVoteRepo := new(MockVoteRepository)
	mockPostRepo := new(MockPostRepository)
	mockCommentRepo := new(MockCommentRepository)
	mockPublisher := new(MockPublisher)

	service := NewVoteService(mockVoteRepo, mockPostRepo, mockCommentRepo, mockPublisher)

	userID := int64(1)
	postID := int64(100)
	existingVote := &models.Vote{
		ID:         1,
		UserID:     userID,
		EntityType: "post",
		EntityID:   postID,
		VoteType:   "up",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockPostRepo.On("FindByID", ctx, postID).Return(&models.Post{
		ID:       postID,
		AuthorID: 2,
	}, nil)

	mockVoteRepo.On("FindByUserAndEntity", ctx, userID, "post", postID).Return(existingVote, nil)

	vote, err := service.Vote(ctx, userID, "post", postID, "up")

	assert.NoError(t, err)
	assert.NotNil(t, vote)
	assert.Equal(t, existingVote.ID, vote.ID)
	assert.Equal(t, "up", vote.VoteType)

	mockVoteRepo.AssertNotCalled(t, "Update")
	mockPublisher.AssertNotCalled(t, "PublishVoteCreated")
	mockPublisher.AssertNotCalled(t, "PublishVoteUpdated")

	mockVoteRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

func TestVoteService_CancelVote_NoVoteExists(t *testing.T) {
	ctx := context.Background()
	mockVoteRepo := new(MockVoteRepository)
	mockPostRepo := new(MockPostRepository)
	mockCommentRepo := new(MockCommentRepository)
	mockPublisher := new(MockPublisher)

	service := NewVoteService(mockVoteRepo, mockPostRepo, mockCommentRepo, mockPublisher)

	mockVoteRepo.On("FindByUserAndEntity", ctx, int64(1), "post", int64(100)).Return(nil, nil)

	err := service.CancelVote(ctx, 1, "post", 100)

	assert.NoError(t, err)
	mockVoteRepo.AssertNotCalled(t, "DeleteByUserAndEntity")
	mockVoteRepo.AssertExpectations(t)
}
