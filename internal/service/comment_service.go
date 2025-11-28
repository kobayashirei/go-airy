package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/mq"
	"github.com/kobayashirei/airy/internal/repository"
)

var (
	// ErrCommentNotFound is returned when comment is not found
	ErrCommentNotFound = errors.New("comment not found")
	// ErrInvalidParentComment is returned when parent comment is invalid
	ErrInvalidParentComment = errors.New("invalid parent comment")
)

// CommentService defines the interface for comment business logic
type CommentService interface {
	CreateComment(ctx context.Context, req CreateCommentRequest) (*models.Comment, error)
	GetCommentTree(ctx context.Context, postID int64) ([]*CommentNode, error)
	DeleteComment(ctx context.Context, commentID int64, userID int64) error
}

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	PostID   int64  `json:"post_id" binding:"required"`
	ParentID *int64 `json:"parent_id"`
	Content  string `json:"content" binding:"required,min=1"`
	AuthorID int64  `json:"-"` // Set from context, not from request body
}

// CommentNode represents a comment with its children in a tree structure
type CommentNode struct {
	*models.Comment
	Children []*CommentNode `json:"children"`
}

// commentService implements CommentService interface
type commentService struct {
	commentRepo  repository.CommentRepository
	postRepo     repository.PostRepository
	messageQueue mq.MessageQueue
}

// NewCommentService creates a new comment service
func NewCommentService(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
	messageQueue mq.MessageQueue,
) CommentService {
	return &commentService{
		commentRepo:  commentRepo,
		postRepo:     postRepo,
		messageQueue: messageQueue,
	}
}

// CreateComment creates a new comment
// Implements Requirements 5.1, 5.2, 5.3, 5.4
func (s *commentService) CreateComment(ctx context.Context, req CreateCommentRequest) (*models.Comment, error) {
	// Verify post exists
	post, err := s.postRepo.FindByID(ctx, req.PostID)
	if err != nil {
		return nil, fmt.Errorf("failed to find post: %w", err)
	}
	if post == nil {
		return nil, ErrPostNotFound
	}

	// Check if comments are allowed
	if !post.AllowComment {
		return nil, errors.New("comments are not allowed on this post")
	}

	now := time.Now()
	comment := &models.Comment{
		Content:   req.Content,
		AuthorID:  req.AuthorID,
		PostID:    req.PostID,
		ParentID:  req.ParentID,
		Status:    "published",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Calculate level, root ID, and path based on parent
	if req.ParentID != nil {
		// This is a reply to another comment
		parentComment, err := s.commentRepo.FindByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to find parent comment: %w", err)
		}
		if parentComment == nil {
			return nil, ErrInvalidParentComment
		}

		// Verify parent comment belongs to the same post
		if parentComment.PostID != req.PostID {
			return nil, errors.New("parent comment does not belong to the same post")
		}

		// Calculate level (parent level + 1)
		comment.Level = parentComment.Level + 1

		// Set root ID (same as parent's root ID)
		comment.RootID = parentComment.RootID

		// Path will be set after we get the comment ID
	} else {
		// This is a root-level comment
		comment.Level = 0
		// RootID will be set to the comment's own ID after creation
		comment.RootID = 0 // Temporary, will be updated
	}

	// Create comment in database
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Update root ID and path after getting the comment ID
	if req.ParentID != nil {
		// For replies, construct path from parent path + comment ID
		parentComment, _ := s.commentRepo.FindByID(ctx, *req.ParentID)
		if parentComment.Path != "" {
			comment.Path = fmt.Sprintf("%s.%d", parentComment.Path, comment.ID)
		} else {
			comment.Path = fmt.Sprintf("%d.%d", parentComment.ID, comment.ID)
		}
	} else {
		// For root comments, set root ID to own ID and path to own ID
		comment.RootID = comment.ID
		comment.Path = strconv.FormatInt(comment.ID, 10)
	}

	// Update the comment with correct root ID and path
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to update comment path: %w", err)
	}

	// Send comment event to message queue for async processing
	// Implements Requirement 5.4
	s.publishCommentCreatedEvent(ctx, comment)

	return comment, nil
}

// GetCommentTree retrieves all comments for a post and builds a tree structure
// Implements Requirement 5.2
func (s *commentService) GetCommentTree(ctx context.Context, postID int64) ([]*CommentNode, error) {
	// Get all comments for the post, ordered by path
	comments, err := s.commentRepo.FindByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to find comments: %w", err)
	}

	// Build tree structure
	tree := s.buildCommentTree(comments)
	return tree, nil
}

// DeleteComment deletes a comment (soft delete)
// Implements Requirement 5.1
func (s *commentService) DeleteComment(ctx context.Context, commentID int64, userID int64) error {
	// Get existing comment
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("failed to find comment: %w", err)
	}
	if comment == nil {
		return ErrCommentNotFound
	}

	// Check authorization
	if comment.AuthorID != userID {
		return ErrUnauthorized
	}

	// Soft delete
	if err := s.commentRepo.Delete(ctx, commentID); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	// Publish delete event
	s.publishCommentDeletedEvent(ctx, comment)

	return nil
}

// buildCommentTree builds a tree structure from a flat list of comments
func (s *commentService) buildCommentTree(comments []*models.Comment) []*CommentNode {
	// Create a map for quick lookup
	nodeMap := make(map[int64]*CommentNode)
	var rootNodes []*CommentNode

	// First pass: create all nodes
	for _, comment := range comments {
		node := &CommentNode{
			Comment:  comment,
			Children: []*CommentNode{},
		}
		nodeMap[comment.ID] = node
	}

	// Second pass: build tree structure
	for _, comment := range comments {
		node := nodeMap[comment.ID]
		if comment.ParentID == nil {
			// Root comment
			rootNodes = append(rootNodes, node)
		} else {
			// Child comment - add to parent's children
			if parentNode, exists := nodeMap[*comment.ParentID]; exists {
				parentNode.Children = append(parentNode.Children, node)
			}
		}
	}

	return rootNodes
}

// publishCommentCreatedEvent publishes a comment created event
func (s *commentService) publishCommentCreatedEvent(ctx context.Context, comment *models.Comment) {
	if s.messageQueue == nil {
		return
	}

	event := mq.CommentCreatedEvent{
		BaseEvent: mq.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: mq.TopicCommentCreated,
			Timestamp: time.Now(),
		},
		CommentID: comment.ID,
		PostID:    comment.PostID,
		AuthorID:  comment.AuthorID,
		ParentID:  comment.ParentID,
		Content:   comment.Content,
	}

	if err := s.messageQueue.Publish(ctx, mq.TopicCommentCreated, event); err != nil {
		fmt.Printf("failed to publish comment created event: %v\n", err)
	}
}

// publishCommentDeletedEvent publishes a comment deleted event
func (s *commentService) publishCommentDeletedEvent(ctx context.Context, comment *models.Comment) {
	if s.messageQueue == nil {
		return
	}

	event := mq.CommentDeletedEvent{
		BaseEvent: mq.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: mq.TopicCommentDeleted,
			Timestamp: time.Now(),
		},
		CommentID: comment.ID,
		PostID:    comment.PostID,
		AuthorID:  comment.AuthorID,
	}

	if err := s.messageQueue.Publish(ctx, mq.TopicCommentDeleted, event); err != nil {
		fmt.Printf("failed to publish comment deleted event: %v\n", err)
	}
}

// ExtractMentions extracts @mentions from comment content
// Returns a list of mentioned usernames
func ExtractMentions(content string) []string {
	var mentions []string
	words := strings.Fields(content)
	
	for _, word := range words {
		if strings.HasPrefix(word, "@") && len(word) > 1 {
			// Remove @ prefix and any trailing punctuation
			username := strings.TrimPrefix(word, "@")
			username = strings.TrimRight(username, ".,!?;:")
			if username != "" {
				mentions = append(mentions, username)
			}
		}
	}
	
	return mentions
}
