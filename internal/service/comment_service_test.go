package service

import (
	"context"
	"testing"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateComment_RootComment(t *testing.T) {
	// Setup
	mockCommentRepo := new(MockCommentRepository)
	mockPostRepo := new(MockPostRepository)
	mockMQ := new(MockMessageQueue)

	service := NewCommentService(mockCommentRepo, mockPostRepo, mockMQ)

	ctx := context.Background()
	postID := int64(1)
	authorID := int64(100)

	// Mock post exists and allows comments
	mockPostRepo.On("FindByID", ctx, postID).Return(&models.Post{
		ID:           postID,
		AllowComment: true,
	}, nil)

	// Mock comment creation
	mockCommentRepo.On("Create", ctx, mock.AnythingOfType("*models.Comment")).Return(nil)
	mockCommentRepo.On("Update", ctx, mock.AnythingOfType("*models.Comment")).Return(nil)

	// Mock message queue
	mockMQ.On("Publish", ctx, mock.Anything, mock.Anything).Return(nil)

	// Execute
	req := CreateCommentRequest{
		PostID:   postID,
		ParentID: nil,
		Content:  "This is a root comment",
		AuthorID: authorID,
	}

	comment, err := service.CreateComment(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, postID, comment.PostID)
	assert.Equal(t, authorID, comment.AuthorID)
	assert.Equal(t, 0, comment.Level)
	assert.Equal(t, comment.ID, comment.RootID)
	assert.Equal(t, "1", comment.Path)

	mockPostRepo.AssertExpectations(t)
	mockCommentRepo.AssertExpectations(t)
	mockMQ.AssertExpectations(t)
}

func TestCreateComment_ReplyComment(t *testing.T) {
	// Setup
	mockCommentRepo := new(MockCommentRepository)
	mockPostRepo := new(MockPostRepository)
	mockMQ := new(MockMessageQueue)

	service := NewCommentService(mockCommentRepo, mockPostRepo, mockMQ)

	ctx := context.Background()
	postID := int64(1)
	parentID := int64(10)
	authorID := int64(100)

	// Mock post exists and allows comments
	mockPostRepo.On("FindByID", ctx, postID).Return(&models.Post{
		ID:           postID,
		AllowComment: true,
	}, nil)

	// Mock parent comment exists
	parentComment := &models.Comment{
		ID:       parentID,
		PostID:   postID,
		Level:    0,
		RootID:   parentID,
		Path:     "10",
		AuthorID: 200,
	}
	mockCommentRepo.On("FindByID", ctx, parentID).Return(parentComment, nil)

	// Mock comment creation
	mockCommentRepo.On("Create", ctx, mock.AnythingOfType("*models.Comment")).Return(nil)
	mockCommentRepo.On("Update", ctx, mock.AnythingOfType("*models.Comment")).Return(nil)

	// Mock message queue
	mockMQ.On("Publish", ctx, mock.Anything, mock.Anything).Return(nil)

	// Execute
	req := CreateCommentRequest{
		PostID:   postID,
		ParentID: &parentID,
		Content:  "This is a reply",
		AuthorID: authorID,
	}

	comment, err := service.CreateComment(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, postID, comment.PostID)
	assert.Equal(t, authorID, comment.AuthorID)
	assert.Equal(t, 1, comment.Level)
	assert.Equal(t, parentID, comment.RootID)
	assert.Equal(t, "10.1", comment.Path)

	mockPostRepo.AssertExpectations(t)
	mockCommentRepo.AssertExpectations(t)
	mockMQ.AssertExpectations(t)
}

func TestCreateComment_PostNotFound(t *testing.T) {
	// Setup
	mockCommentRepo := new(MockCommentRepository)
	mockPostRepo := new(MockPostRepository)
	mockMQ := new(MockMessageQueue)

	service := NewCommentService(mockCommentRepo, mockPostRepo, mockMQ)

	ctx := context.Background()
	postID := int64(999)

	// Mock post not found
	mockPostRepo.On("FindByID", ctx, postID).Return(nil, nil)

	// Execute
	req := CreateCommentRequest{
		PostID:   postID,
		Content:  "Comment on non-existent post",
		AuthorID: 100,
	}

	comment, err := service.CreateComment(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, ErrPostNotFound, err)

	mockPostRepo.AssertExpectations(t)
}

func TestCreateComment_CommentsNotAllowed(t *testing.T) {
	// Setup
	mockCommentRepo := new(MockCommentRepository)
	mockPostRepo := new(MockPostRepository)
	mockMQ := new(MockMessageQueue)

	service := NewCommentService(mockCommentRepo, mockPostRepo, mockMQ)

	ctx := context.Background()
	postID := int64(1)

	// Mock post exists but doesn't allow comments
	mockPostRepo.On("FindByID", ctx, postID).Return(&models.Post{
		ID:           postID,
		AllowComment: false,
	}, nil)

	// Execute
	req := CreateCommentRequest{
		PostID:   postID,
		Content:  "Comment on locked post",
		AuthorID: 100,
	}

	comment, err := service.CreateComment(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Contains(t, err.Error(), "comments are not allowed")

	mockPostRepo.AssertExpectations(t)
}

func TestGetCommentTree(t *testing.T) {
	// Setup
	mockCommentRepo := new(MockCommentRepository)
	mockPostRepo := new(MockPostRepository)
	mockMQ := new(MockMessageQueue)

	service := NewCommentService(mockCommentRepo, mockPostRepo, mockMQ)

	ctx := context.Background()
	postID := int64(1)

	// Create test comments
	comment1 := &models.Comment{
		ID:       1,
		PostID:   postID,
		ParentID: nil,
		Level:    0,
		RootID:   1,
		Path:     "1",
		Content:  "Root comment 1",
	}
	comment2 := &models.Comment{
		ID:       2,
		PostID:   postID,
		ParentID: func() *int64 { id := int64(1); return &id }(),
		Level:    1,
		RootID:   1,
		Path:     "1.2",
		Content:  "Reply to comment 1",
	}
	comment3 := &models.Comment{
		ID:       3,
		PostID:   postID,
		ParentID: nil,
		Level:    0,
		RootID:   3,
		Path:     "3",
		Content:  "Root comment 2",
	}

	comments := []*models.Comment{comment1, comment2, comment3}

	// Mock repository
	mockCommentRepo.On("FindByPostID", ctx, postID).Return(comments, nil)

	// Execute
	tree, err := service.GetCommentTree(ctx, postID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, tree)
	assert.Len(t, tree, 2) // Two root comments
	assert.Len(t, tree[0].Children, 1) // First root has one child
	assert.Len(t, tree[1].Children, 0) // Second root has no children

	mockCommentRepo.AssertExpectations(t)
}

func TestDeleteComment(t *testing.T) {
	// Setup
	mockCommentRepo := new(MockCommentRepository)
	mockPostRepo := new(MockPostRepository)
	mockMQ := new(MockMessageQueue)

	service := NewCommentService(mockCommentRepo, mockPostRepo, mockMQ)

	ctx := context.Background()
	commentID := int64(1)
	userID := int64(100)

	// Mock comment exists and belongs to user
	comment := &models.Comment{
		ID:       commentID,
		AuthorID: userID,
		PostID:   1,
	}
	mockCommentRepo.On("FindByID", ctx, commentID).Return(comment, nil)
	mockCommentRepo.On("Delete", ctx, commentID).Return(nil)

	// Mock message queue
	mockMQ.On("Publish", ctx, mock.Anything, mock.Anything).Return(nil)

	// Execute
	err := service.DeleteComment(ctx, commentID, userID)

	// Assert
	assert.NoError(t, err)

	mockCommentRepo.AssertExpectations(t)
	mockMQ.AssertExpectations(t)
}

func TestDeleteComment_Unauthorized(t *testing.T) {
	// Setup
	mockCommentRepo := new(MockCommentRepository)
	mockPostRepo := new(MockPostRepository)
	mockMQ := new(MockMessageQueue)

	service := NewCommentService(mockCommentRepo, mockPostRepo, mockMQ)

	ctx := context.Background()
	commentID := int64(1)
	userID := int64(100)
	differentUserID := int64(200)

	// Mock comment exists but belongs to different user
	comment := &models.Comment{
		ID:       commentID,
		AuthorID: differentUserID,
		PostID:   1,
	}
	mockCommentRepo.On("FindByID", ctx, commentID).Return(comment, nil)

	// Execute
	err := service.DeleteComment(ctx, commentID, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)

	mockCommentRepo.AssertExpectations(t)
}

func TestExtractMentions(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "Single mention",
			content:  "Hello @john, how are you?",
			expected: []string{"john"},
		},
		{
			name:     "Multiple mentions",
			content:  "@alice and @bob are invited",
			expected: []string{"alice", "bob"},
		},
		{
			name:     "Mention with punctuation",
			content:  "Thanks @user123!",
			expected: []string{"user123"},
		},
		{
			name:     "No mentions",
			content:  "This is a regular comment",
			expected: nil,
		},
		{
			name:     "Mention at end",
			content:  "Replying to @someone.",
			expected: []string{"someone"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractMentions(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}
