package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/repository"
	"github.com/kobayashirei/airy/internal/search"
	"go.uber.org/zap"
)

// SearchService defines the interface for search operations
type SearchService interface {
	// Post search operations
	IndexPost(ctx context.Context, post *models.Post) error
	UpdatePost(ctx context.Context, postID int64, post *models.Post) error
	DeletePost(ctx context.Context, postID int64) error
	SearchPosts(ctx context.Context, query SearchQuery) (*SearchResult, error)

	// User search operations
	IndexUser(ctx context.Context, user *models.User, profile *models.UserProfile, stats *models.UserStats) error
	UpdateUser(ctx context.Context, userID int64, user *models.User, profile *models.UserProfile, stats *models.UserStats) error
	DeleteUser(ctx context.Context, userID int64) error
	SearchUsers(ctx context.Context, query SearchQuery) (*SearchResult, error)

	// Initialize indices
	InitializeIndices(ctx context.Context) error
}

// SearchQuery represents a search query
type SearchQuery struct {
	Keyword  string
	CircleID *int64
	Tags     []string
	SortBy   string // "time", "hotness", "relevance"
	Page     int
	PageSize int
}

// SearchResult represents search results
type SearchResult struct {
	Total   int64
	Results []SearchResultItem
}

// SearchResultItem represents a single search result
type SearchResultItem struct {
	ID    int64
	Score float64
	Data  map[string]interface{}
}

type searchService struct {
	esClient   *search.Client
	userRepo   repository.UserRepository
	postRepo   repository.PostRepository
	circleRepo repository.CircleRepository
	log        *zap.Logger
}

// NewSearchService creates a new search service
func NewSearchService(
	esClient *search.Client,
	userRepo repository.UserRepository,
	postRepo repository.PostRepository,
	circleRepo repository.CircleRepository,
	log *zap.Logger,
) SearchService {
	return &searchService{
		esClient:   esClient,
		userRepo:   userRepo,
		postRepo:   postRepo,
		circleRepo: circleRepo,
		log:        log,
	}
}

// InitializeIndices creates the necessary Elasticsearch indices
func (s *searchService) InitializeIndices(ctx context.Context) error {
	// Create posts index
	if err := s.esClient.CreateIndex(ctx, search.PostIndex, search.PostIndexMapping); err != nil {
		return fmt.Errorf("failed to create posts index: %w", err)
	}

	// Create users index
	if err := s.esClient.CreateIndex(ctx, search.UserIndex, search.UserIndexMapping); err != nil {
		return fmt.Errorf("failed to create users index: %w", err)
	}

	s.log.Info("Search indices initialized successfully")
	return nil
}

// IndexPost indexes a post in Elasticsearch
func (s *searchService) IndexPost(ctx context.Context, post *models.Post) error {
	// Get author information
	author, err := s.userRepo.FindByID(ctx, post.AuthorID)
	if err != nil {
		return fmt.Errorf("failed to get author: %w", err)
	}

	// Get circle information if applicable
	var circleName string
	if post.CircleID != nil {
		circle, err := s.circleRepo.FindByID(ctx, *post.CircleID)
		if err == nil && circle != nil {
			circleName = circle.Name
		}
	}

	// Parse tags
	var tags []string
	if post.Tags != "" {
		if err := json.Unmarshal([]byte(post.Tags), &tags); err != nil {
			s.log.Warn(fmt.Sprintf("Failed to parse tags for post %d: %v", post.ID, err))
			tags = []string{}
		}
	}

	// Create document
	doc := map[string]interface{}{
		"id":              post.ID,
		"title":           post.Title,
		"content":         post.ContentMarkdown,
		"summary":         post.Summary,
		"author_id":       post.AuthorID,
		"author_username": author.Username,
		"status":          post.Status,
		"category":        post.Category,
		"tags":            tags,
		"view_count":      post.ViewCount,
		"hotness_score":   post.HotnessScore,
		"created_at":      post.CreatedAt,
		"updated_at":      post.UpdatedAt,
	}

	if post.CircleID != nil {
		doc["circle_id"] = *post.CircleID
		doc["circle_name"] = circleName
	}

	if post.PublishedAt != nil {
		doc["published_at"] = *post.PublishedAt
	}

	// Index the document
	if err := s.esClient.Index(ctx, search.PostIndex, strconv.FormatInt(post.ID, 10), doc); err != nil {
		return fmt.Errorf("failed to index post: %w", err)
	}

	s.log.Info("Indexed post in Elasticsearch", zap.Int64("post_id", post.ID))
	return nil
}

// UpdatePost updates a post in Elasticsearch
func (s *searchService) UpdatePost(ctx context.Context, postID int64, post *models.Post) error {
	// Get author information
	author, err := s.userRepo.FindByID(ctx, post.AuthorID)
	if err != nil {
		return fmt.Errorf("failed to get author: %w", err)
	}

	// Get circle information if applicable
	var circleName string
	if post.CircleID != nil {
		circle, err := s.circleRepo.FindByID(ctx, *post.CircleID)
		if err == nil && circle != nil {
			circleName = circle.Name
		}
	}

	// Parse tags
	var tags []string
	if post.Tags != "" {
		if err := json.Unmarshal([]byte(post.Tags), &tags); err != nil {
			s.log.Warn(fmt.Sprintf("Failed to parse tags for post %d: %v", post.ID, err))
			tags = []string{}
		}
	}

	// Create document
	doc := map[string]interface{}{
		"title":           post.Title,
		"content":         post.ContentMarkdown,
		"summary":         post.Summary,
		"author_username": author.Username,
		"status":          post.Status,
		"category":        post.Category,
		"tags":            tags,
		"view_count":      post.ViewCount,
		"hotness_score":   post.HotnessScore,
		"updated_at":      post.UpdatedAt,
	}

	if post.CircleID != nil {
		doc["circle_name"] = circleName
	}

	if post.PublishedAt != nil {
		doc["published_at"] = *post.PublishedAt
	}

	// Update the document
	if err := s.esClient.Update(ctx, search.PostIndex, strconv.FormatInt(postID, 10), doc); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	s.log.Info("Updated post in Elasticsearch", zap.Int64("post_id", postID))
	return nil
}

// DeletePost deletes a post from Elasticsearch
func (s *searchService) DeletePost(ctx context.Context, postID int64) error {
	if err := s.esClient.Delete(ctx, search.PostIndex, strconv.FormatInt(postID, 10)); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	s.log.Info("Deleted post from Elasticsearch", zap.Int64("post_id", postID))
	return nil
}

// SearchPosts searches for posts
func (s *searchService) SearchPosts(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	// Build Elasticsearch query
	esQuery := s.buildPostSearchQuery(query)

	// Execute search
	res, err := s.esClient.Search(ctx, search.PostIndex, esQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts: %w", err)
	}

	// Convert results
	results := make([]SearchResultItem, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		id, _ := strconv.ParseInt(hit.ID, 10, 64)
		results = append(results, SearchResultItem{
			ID:    id,
			Score: hit.Score,
			Data:  hit.Source,
		})
	}

	return &SearchResult{
		Total:   res.Hits.Total.Value,
		Results: results,
	}, nil
}

// buildPostSearchQuery builds an Elasticsearch query for post search
func (s *searchService) buildPostSearchQuery(query SearchQuery) map[string]interface{} {
	// Set defaults
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	if query.SortBy == "" {
		query.SortBy = "relevance"
	}

	// Build must clauses
	mustClauses := []map[string]interface{}{
		{
			"term": map[string]interface{}{
				"status": "published",
			},
		},
	}

	// Add keyword search if provided
	if query.Keyword != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query.Keyword,
				"fields": []string{"title^3", "content", "summary^2"},
				"type":   "best_fields",
			},
		})
	}

	// Add circle filter if provided
	if query.CircleID != nil {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"circle_id": *query.CircleID,
			},
		})
	}

	// Add tags filter if provided
	if len(query.Tags) > 0 {
		mustClauses = append(mustClauses, map[string]interface{}{
			"terms": map[string]interface{}{
				"tags": query.Tags,
			},
		})
	}

	// Build sort
	var sort []map[string]interface{}
	switch query.SortBy {
	case "time":
		sort = []map[string]interface{}{
			{"published_at": map[string]interface{}{"order": "desc"}},
		}
	case "hotness":
		sort = []map[string]interface{}{
			{"hotness_score": map[string]interface{}{"order": "desc"}},
		}
	default: // relevance
		sort = []map[string]interface{}{
			{"_score": map[string]interface{}{"order": "desc"}},
		}
	}

	// Build final query
	esQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		},
		"sort": sort,
		"from": (query.Page - 1) * query.PageSize,
		"size": query.PageSize,
	}

	return esQuery
}

// IndexUser indexes a user in Elasticsearch
func (s *searchService) IndexUser(ctx context.Context, user *models.User, profile *models.UserProfile, stats *models.UserStats) error {
	doc := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"bio":        user.Bio,
		"status":     user.Status,
		"created_at": user.CreatedAt,
	}

	if profile != nil {
		doc["follower_count"] = profile.FollowerCount
		doc["following_count"] = profile.FollowingCount
	}

	if stats != nil {
		doc["post_count"] = stats.PostCount
	}

	if err := s.esClient.Index(ctx, search.UserIndex, strconv.FormatInt(user.ID, 10), doc); err != nil {
		return fmt.Errorf("failed to index user: %w", err)
	}

	s.log.Info("Indexed user in Elasticsearch", zap.Int64("user_id", user.ID))
	return nil
}

// UpdateUser updates a user in Elasticsearch
func (s *searchService) UpdateUser(ctx context.Context, userID int64, user *models.User, profile *models.UserProfile, stats *models.UserStats) error {
	doc := map[string]interface{}{
		"username": user.Username,
		"bio":      user.Bio,
		"status":   user.Status,
	}

	if profile != nil {
		doc["follower_count"] = profile.FollowerCount
		doc["following_count"] = profile.FollowingCount
	}

	if stats != nil {
		doc["post_count"] = stats.PostCount
	}

	if err := s.esClient.Update(ctx, search.UserIndex, strconv.FormatInt(userID, 10), doc); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	s.log.Info("Updated user in Elasticsearch", zap.Int64("user_id", userID))
	return nil
}

// DeleteUser deletes a user from Elasticsearch
func (s *searchService) DeleteUser(ctx context.Context, userID int64) error {
	if err := s.esClient.Delete(ctx, search.UserIndex, strconv.FormatInt(userID, 10)); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.log.Info("Deleted user from Elasticsearch", zap.Int64("user_id", userID))
	return nil
}

// SearchUsers searches for users
func (s *searchService) SearchUsers(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	// Build Elasticsearch query
	esQuery := s.buildUserSearchQuery(query)

	// Execute search
	res, err := s.esClient.Search(ctx, search.UserIndex, esQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Convert results
	results := make([]SearchResultItem, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		id, _ := strconv.ParseInt(hit.ID, 10, 64)
		results = append(results, SearchResultItem{
			ID:    id,
			Score: hit.Score,
			Data:  hit.Source,
		})
	}

	return &SearchResult{
		Total:   res.Hits.Total.Value,
		Results: results,
	}, nil
}

// buildUserSearchQuery builds an Elasticsearch query for user search
func (s *searchService) buildUserSearchQuery(query SearchQuery) map[string]interface{} {
	// Set defaults
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// Build must clauses
	mustClauses := []map[string]interface{}{
		{
			"term": map[string]interface{}{
				"status": "active",
			},
		},
	}

	// Add keyword search if provided
	if query.Keyword != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query.Keyword,
				"fields": []string{"username^3", "bio"},
				"type":   "best_fields",
			},
		})
	}

	// Build sort
	sort := []map[string]interface{}{
		{"_score": map[string]interface{}{"order": "desc"}},
		{"follower_count": map[string]interface{}{"order": "desc"}},
	}

	// Build final query
	esQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		},
		"sort": sort,
		"from": (query.Page - 1) * query.PageSize,
		"size": query.PageSize,
	}

	return esQuery
}
