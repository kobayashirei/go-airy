package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kobayashirei/airy/internal/response"
	"github.com/kobayashirei/airy/internal/service"
)

// SearchHandler handles search-related HTTP requests
type SearchHandler struct {
	searchService service.SearchService
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(searchService service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// SearchPostsRequest represents the request for searching posts
type SearchPostsRequest struct {
	Keyword  string   `form:"keyword"`
	CircleID *int64   `form:"circle_id"`
	Tags     []string `form:"tags"`
	SortBy   string   `form:"sort_by"` // "time", "hotness", "relevance"
	Page     int      `form:"page"`
	PageSize int      `form:"page_size"`
}

// SearchUsersRequest represents the request for searching users
type SearchUsersRequest struct {
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// SearchPostsResponse represents the response for post search
type SearchPostsResponse struct {
	Total   int64                  `json:"total"`
	Page    int                    `json:"page"`
	Results []PostSearchResultItem `json:"results"`
}

// PostSearchResultItem represents a single post search result
type PostSearchResultItem struct {
	ID             int64    `json:"id"`
	Title          string   `json:"title"`
	Summary        string   `json:"summary"`
	AuthorID       int64    `json:"author_id"`
	AuthorUsername string   `json:"author_username"`
	CircleID       *int64   `json:"circle_id,omitempty"`
	CircleName     string   `json:"circle_name,omitempty"`
	Category       string   `json:"category"`
	Tags           []string `json:"tags"`
	ViewCount      int      `json:"view_count"`
	HotnessScore   float64  `json:"hotness_score"`
	PublishedAt    string   `json:"published_at"`
	Score          float64  `json:"score"`
}

// SearchUsersResponse represents the response for user search
type SearchUsersResponse struct {
	Total   int64                  `json:"total"`
	Page    int                    `json:"page"`
	Results []UserSearchResultItem `json:"results"`
}

// UserSearchResultItem represents a single user search result
type UserSearchResultItem struct {
	ID             int64   `json:"id"`
	Username       string  `json:"username"`
	Bio            string  `json:"bio"`
	FollowerCount  int     `json:"follower_count"`
	FollowingCount int     `json:"following_count"`
	PostCount      int     `json:"post_count"`
	Score          float64 `json:"score"`
}

// SearchPosts handles POST /api/v1/search/posts
// Implements Requirements 9.2, 9.3, 9.4
func (h *SearchHandler) SearchPosts(c *gin.Context) {
	var req SearchPostsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	// Parse tags if provided as comma-separated string
	tagsParam := c.Query("tags")
	if tagsParam != "" {
		req.Tags = strings.Split(tagsParam, ",")
	}

	// Build search query
	query := service.SearchQuery{
		Keyword:  req.Keyword,
		CircleID: req.CircleID,
		Tags:     req.Tags,
		SortBy:   req.SortBy,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// Execute search
	result, err := h.searchService.SearchPosts(c.Request.Context(), query)
	if err != nil {
		response.InternalError(c, "Failed to search posts")
		return
	}

	// Convert results to response format
	results := make([]PostSearchResultItem, 0, len(result.Results))
	for _, item := range result.Results {
		resultItem := PostSearchResultItem{
			ID:    item.ID,
			Score: item.Score,
		}

		// Extract fields from source data
		if title, ok := item.Data["title"].(string); ok {
			resultItem.Title = title
		}
		if summary, ok := item.Data["summary"].(string); ok {
			resultItem.Summary = summary
		}
		if authorID, ok := item.Data["author_id"].(float64); ok {
			resultItem.AuthorID = int64(authorID)
		}
		if authorUsername, ok := item.Data["author_username"].(string); ok {
			resultItem.AuthorUsername = authorUsername
		}
		if circleID, ok := item.Data["circle_id"].(float64); ok {
			cid := int64(circleID)
			resultItem.CircleID = &cid
		}
		if circleName, ok := item.Data["circle_name"].(string); ok {
			resultItem.CircleName = circleName
		}
		if category, ok := item.Data["category"].(string); ok {
			resultItem.Category = category
		}
		if tags, ok := item.Data["tags"].([]interface{}); ok {
			resultItem.Tags = make([]string, 0, len(tags))
			for _, tag := range tags {
				if tagStr, ok := tag.(string); ok {
					resultItem.Tags = append(resultItem.Tags, tagStr)
				}
			}
		}
		if viewCount, ok := item.Data["view_count"].(float64); ok {
			resultItem.ViewCount = int(viewCount)
		}
		if hotnessScore, ok := item.Data["hotness_score"].(float64); ok {
			resultItem.HotnessScore = hotnessScore
		}
		if publishedAt, ok := item.Data["published_at"].(string); ok {
			resultItem.PublishedAt = publishedAt
		}

		results = append(results, resultItem)
	}

	resp := SearchPostsResponse{
		Total:   result.Total,
		Page:    query.Page,
		Results: results,
	}

	response.Success(c, resp)
}

// SearchUsers handles GET /api/v1/search/users
// Implements Requirements 9.2, 9.3, 9.4
func (h *SearchHandler) SearchUsers(c *gin.Context) {
	var req SearchUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	// Build search query
	query := service.SearchQuery{
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// Execute search
	result, err := h.searchService.SearchUsers(c.Request.Context(), query)
	if err != nil {
		response.InternalError(c, "Failed to search users")
		return
	}

	// Convert results to response format
	results := make([]UserSearchResultItem, 0, len(result.Results))
	for _, item := range result.Results {
		resultItem := UserSearchResultItem{
			ID:    item.ID,
			Score: item.Score,
		}

		// Extract fields from source data
		if username, ok := item.Data["username"].(string); ok {
			resultItem.Username = username
		}
		if bio, ok := item.Data["bio"].(string); ok {
			resultItem.Bio = bio
		}
		if followerCount, ok := item.Data["follower_count"].(float64); ok {
			resultItem.FollowerCount = int(followerCount)
		}
		if followingCount, ok := item.Data["following_count"].(float64); ok {
			resultItem.FollowingCount = int(followingCount)
		}
		if postCount, ok := item.Data["post_count"].(float64); ok {
			resultItem.PostCount = int(postCount)
		}

		results = append(results, resultItem)
	}

	resp := SearchUsersResponse{
		Total:   result.Total,
		Page:    query.Page,
		Results: results,
	}

	response.Success(c, resp)
}

// RegisterRoutes registers search routes
func (h *SearchHandler) RegisterRoutes(r *gin.RouterGroup) {
	search := r.Group("/search")
	{
		search.GET("/posts", h.SearchPosts)
		search.GET("/users", h.SearchUsers)
	}
}
