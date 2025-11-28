package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"

	"github.com/kobayashirei/airy/internal/cache"
	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/mq"
	"github.com/kobayashirei/airy/internal/repository"
	"github.com/kobayashirei/airy/internal/taskpool"
)

var (
	// ErrPostNotFound is returned when post is not found
	ErrPostNotFound = errors.New("post not found")
	// ErrUnauthorized is returned when user is not authorized
	ErrUnauthorized = errors.New("unauthorized")
)

// PostService defines the interface for post business logic
type PostService interface {
	CreatePost(ctx context.Context, req CreatePostRequest) (*models.Post, error)
	GetPost(ctx context.Context, postID int64, userID *int64) (*models.Post, error)
	UpdatePost(ctx context.Context, postID int64, userID int64, req UpdatePostRequest) (*models.Post, error)
	DeletePost(ctx context.Context, postID int64, userID int64) error
	ListPosts(ctx context.Context, opts ListPostsRequest) (*ListPostsResponse, error)
}

// CreatePostRequest represents a request to create a post
type CreatePostRequest struct {
	Title        string  `json:"title" binding:"required,min=1,max=255"`
	Content      string  `json:"content" binding:"required,min=1"`
	Summary      string  `json:"summary" binding:"max=500"`
	CoverImage   string  `json:"cover_image" binding:"max=255"`
	CircleID     *int64  `json:"circle_id"`
	Category     string  `json:"category" binding:"max=50"`
	Tags         string  `json:"tags"` // JSON array
	ScheduledAt  *time.Time `json:"scheduled_at"`
	AllowComment bool    `json:"allow_comment"`
	IsAnonymous  bool    `json:"is_anonymous"`
	AuthorID     int64   `json:"-"` // Set from context, not from request body
}

// UpdatePostRequest represents a request to update a post
type UpdatePostRequest struct {
	Title        *string `json:"title" binding:"omitempty,min=1,max=255"`
	Content      *string `json:"content" binding:"omitempty,min=1"`
	Summary      *string `json:"summary" binding:"omitempty,max=500"`
	CoverImage   *string `json:"cover_image" binding:"omitempty,max=255"`
	Category     *string `json:"category" binding:"omitempty,max=50"`
	Tags         *string `json:"tags"`
	AllowComment *bool   `json:"allow_comment"`
}

// ListPostsRequest represents a request to list posts
type ListPostsRequest struct {
	AuthorID *int64  `form:"author_id"`
	CircleID *int64  `form:"circle_id"`
	Status   string  `form:"status"`
	SortBy   string  `form:"sort_by"` // "created_at", "hotness_score", "view_count"
	Order    string  `form:"order"`   // "asc", "desc"
	Page     int     `form:"page" binding:"min=1"`
	PageSize int     `form:"page_size" binding:"min=1,max=100"`
}

// ListPostsResponse represents a response for listing posts
type ListPostsResponse struct {
	Posts      []*models.Post `json:"posts"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// postService implements PostService interface
type postService struct {
	postRepo       repository.PostRepository
	cacheService   cache.Service
	moderationService ContentModerationService
	messageQueue   mq.MessageQueue
	taskPool       *taskpool.Pool
	sanitizer      *bluemonday.Policy
}

// NewPostService creates a new post service
func NewPostService(
	postRepo repository.PostRepository,
	cacheService cache.Service,
	moderationService ContentModerationService,
	messageQueue mq.MessageQueue,
	taskPool *taskpool.Pool,
) PostService {
	// Create HTML sanitizer policy
	sanitizer := bluemonday.UGCPolicy()
	
	return &postService{
		postRepo:          postRepo,
		cacheService:      cacheService,
		moderationService: moderationService,
		messageQueue:      messageQueue,
		taskPool:          taskPool,
		sanitizer:         sanitizer,
	}
}

// CreatePost creates a new post
// Implements Requirements 4.1, 4.2, 4.3, 4.4
func (s *postService) CreatePost(ctx context.Context, req CreatePostRequest) (*models.Post, error) {
	// Convert Markdown to HTML
	htmlContent := s.markdownToHTML(req.Content)
	
	// Sanitize HTML to prevent XSS
	htmlContent = s.sanitizer.Sanitize(htmlContent)
	
	// Create post object
	now := time.Now()
	post := &models.Post{
		Title:           req.Title,
		ContentMarkdown: req.Content,
		ContentHTML:     htmlContent,
		Summary:         req.Summary,
		CoverImage:      req.CoverImage,
		AuthorID:        req.AuthorID,
		CircleID:        req.CircleID,
		Status:          "draft", // Initial status
		Category:        req.Category,
		Tags:            req.Tags,
		ScheduledAt:     req.ScheduledAt,
		AllowComment:    req.AllowComment,
		IsAnonymous:     req.IsAnonymous,
		ViewCount:       0,
		HotnessScore:    0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	
	// Call content moderation service
	moderationResult, err := s.moderationService.CheckContent(ctx, req.Content)
	if err != nil {
		return nil, fmt.Errorf("content moderation failed: %w", err)
	}
	
	// Map moderation result to post status
	post.Status = MapModerationStatusToPostStatus(moderationResult.Status)
	
	// If status is published, set published timestamp
	if post.Status == "published" {
		publishedAt := now
		post.PublishedAt = &publishedAt
	}
	
	// Save post to database
	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	
	// If post is published, trigger async tasks
	if post.Status == "published" {
		s.triggerAsyncTasks(ctx, post)
	}
	
	return post, nil
}

// GetPost retrieves a post by ID
func (s *postService) GetPost(ctx context.Context, postID int64, userID *int64) (*models.Post, error) {
	// Try to get from cache first
	cacheKey := cache.PostKey(postID)
	var post models.Post
	err := s.cacheService.Get(ctx, cacheKey, &post)
	if err == nil {
		// Found in cache, increment view count asynchronously
		s.incrementViewCountAsync(ctx, postID)
		return &post, nil
	}
	
	// Cache miss, get from database
	post2, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to find post: %w", err)
	}
	if post2 == nil {
		return nil, ErrPostNotFound
	}
	
	// Cache the post
	if err := s.cacheService.Set(ctx, cacheKey, post2, 1*time.Hour); err != nil {
		// Log error but don't fail the request
		fmt.Printf("failed to cache post: %v\n", err)
	}
	
	// Increment view count asynchronously
	s.incrementViewCountAsync(ctx, postID)
	
	return post2, nil
}

// UpdatePost updates an existing post
func (s *postService) UpdatePost(ctx context.Context, postID int64, userID int64, req UpdatePostRequest) (*models.Post, error) {
	// Get existing post
	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to find post: %w", err)
	}
	if post == nil {
		return nil, ErrPostNotFound
	}
	
	// Check authorization
	if post.AuthorID != userID {
		return nil, ErrUnauthorized
	}
	
	// Update fields if provided
	needsModeration := false
	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Content != nil {
		post.ContentMarkdown = *req.Content
		post.ContentHTML = s.sanitizer.Sanitize(s.markdownToHTML(*req.Content))
		needsModeration = true
	}
	if req.Summary != nil {
		post.Summary = *req.Summary
	}
	if req.CoverImage != nil {
		post.CoverImage = *req.CoverImage
	}
	if req.Category != nil {
		post.Category = *req.Category
	}
	if req.Tags != nil {
		post.Tags = *req.Tags
	}
	if req.AllowComment != nil {
		post.AllowComment = *req.AllowComment
	}
	
	// If content changed, re-moderate
	if needsModeration {
		moderationResult, err := s.moderationService.CheckContent(ctx, post.ContentMarkdown)
		if err != nil {
			return nil, fmt.Errorf("content moderation failed: %w", err)
		}
		post.Status = MapModerationStatusToPostStatus(moderationResult.Status)
		
		// Update published timestamp if newly published
		if post.Status == "published" && post.PublishedAt == nil {
			now := time.Now()
			post.PublishedAt = &now
		}
	}
	
	post.UpdatedAt = time.Now()
	
	// Save to database
	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}
	
	// Invalidate cache
	cacheKey := cache.PostKey(postID)
	if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("failed to invalidate cache: %v\n", err)
	}
	
	// Publish update event
	s.publishPostUpdatedEvent(ctx, post)
	
	return post, nil
}

// DeletePost deletes a post (soft delete)
func (s *postService) DeletePost(ctx context.Context, postID int64, userID int64) error {
	// Get existing post
	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to find post: %w", err)
	}
	if post == nil {
		return ErrPostNotFound
	}
	
	// Check authorization
	if post.AuthorID != userID {
		return ErrUnauthorized
	}
	
	// Soft delete
	if err := s.postRepo.Delete(ctx, postID); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	
	// Invalidate cache
	cacheKey := cache.PostKey(postID)
	if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("failed to invalidate cache: %v\n", err)
	}
	
	// Publish delete event
	s.publishPostDeletedEvent(ctx, post)
	
	return nil
}

// ListPosts lists posts with pagination and filtering
func (s *postService) ListPosts(ctx context.Context, req ListPostsRequest) (*ListPostsResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	
	// Build repository options
	opts := repository.PostListOptions{
		AuthorID: req.AuthorID,
		CircleID: req.CircleID,
		Status:   req.Status,
		SortBy:   req.SortBy,
		Order:    req.Order,
		Limit:    req.PageSize,
		Offset:   (req.Page - 1) * req.PageSize,
	}
	
	// Get posts
	posts, err := s.postRepo.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}
	
	// Get total count
	total, err := s.postRepo.Count(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to count posts: %w", err)
	}
	
	// Calculate total pages
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}
	
	return &ListPostsResponse{
		Posts:      posts,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// markdownToHTML converts markdown to HTML
func (s *postService) markdownToHTML(markdown string) string {
	// Use blackfriday to convert markdown to HTML
	html := blackfriday.Run([]byte(markdown))
	return string(html)
}

// incrementViewCountAsync increments view count asynchronously
func (s *postService) incrementViewCountAsync(ctx context.Context, postID int64) {
	if s.taskPool != nil {
		s.taskPool.SubmitFunc(func(ctx context.Context) error {
			return s.postRepo.IncrementViewCount(ctx, postID)
		})
	}
}

// triggerAsyncTasks triggers async tasks for post publication
// Implements Requirement 4.4
func (s *postService) triggerAsyncTasks(ctx context.Context, post *models.Post) {
	// Publish post published event to message queue
	// This will trigger:
	// - Search index update
	// - Feed push
	// - Notifications
	s.publishPostPublishedEvent(ctx, post)
}

// publishPostPublishedEvent publishes a post published event
func (s *postService) publishPostPublishedEvent(ctx context.Context, post *models.Post) {
	if s.messageQueue == nil {
		return
	}
	
	event := mq.PostPublishedEvent{
		BaseEvent: mq.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: mq.TopicPostPublished,
			Timestamp: time.Now(),
		},
		PostID:   post.ID,
		AuthorID: post.AuthorID,
		CircleID: post.CircleID,
		Title:    post.Title,
	}
	
	if err := s.messageQueue.Publish(ctx, mq.TopicPostPublished, event); err != nil {
		fmt.Printf("failed to publish post published event: %v\n", err)
	}
}

// publishPostUpdatedEvent publishes a post updated event
func (s *postService) publishPostUpdatedEvent(ctx context.Context, post *models.Post) {
	if s.messageQueue == nil {
		return
	}
	
	event := mq.PostUpdatedEvent{
		BaseEvent: mq.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: mq.TopicPostUpdated,
			Timestamp: time.Now(),
		},
		PostID:   post.ID,
		AuthorID: post.AuthorID,
	}
	
	if err := s.messageQueue.Publish(ctx, mq.TopicPostUpdated, event); err != nil {
		fmt.Printf("failed to publish post updated event: %v\n", err)
	}
}

// publishPostDeletedEvent publishes a post deleted event
func (s *postService) publishPostDeletedEvent(ctx context.Context, post *models.Post) {
	if s.messageQueue == nil {
		return
	}
	
	event := mq.PostDeletedEvent{
		BaseEvent: mq.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: mq.TopicPostDeleted,
			Timestamp: time.Now(),
		},
		PostID:   post.ID,
		AuthorID: post.AuthorID,
	}
	
	if err := s.messageQueue.Publish(ctx, mq.TopicPostDeleted, event); err != nil {
		fmt.Printf("failed to publish post deleted event: %v\n", err)
	}
}
