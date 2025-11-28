package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/repository"
)

// AdminService defines the interface for admin business logic
type AdminService interface {
	// Dashboard
	GetDashboard(ctx context.Context) (*DashboardResponse, error)

	// User Management
	ListUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error)
	BanUser(ctx context.Context, operatorID, userID int64, reason, ip string) error
	UnbanUser(ctx context.Context, operatorID, userID int64, ip string) error

	// Content Management
	ListPosts(ctx context.Context, req AdminListPostsRequest) (*AdminListPostsResponse, error)
	BatchReviewPosts(ctx context.Context, req BatchReviewRequest) error

	// Audit Logs
	ListLogs(ctx context.Context, req ListLogsRequest) (*ListLogsResponse, error)
	LogAction(ctx context.Context, log *models.AdminLog) error
}

// DashboardResponse represents dashboard metrics
type DashboardResponse struct {
	TotalUsers       int64   `json:"total_users"`
	ActiveUsers      int64   `json:"active_users"`
	TotalPosts       int64   `json:"total_posts"`
	PublishedPosts   int64   `json:"published_posts"`
	PendingPosts     int64   `json:"pending_posts"`
	TotalComments    int64   `json:"total_comments"`
	DailyActiveUsers int64   `json:"daily_active_users"`
	PostsToday       int64   `json:"posts_today"`
}

// ListUsersRequest represents a request to list users
type ListUsersRequest struct {
	Status   string `json:"status"`
	Keyword  string `json:"keyword"`
	SortBy   string `json:"sort_by"`
	Order    string `json:"order"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// ListUsersResponse represents a response with user list
type ListUsersResponse struct {
	Users      []*models.User `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// AdminListPostsRequest represents a request to list posts in admin panel
type AdminListPostsRequest struct {
	Status   string `json:"status"`
	AuthorID *int64 `json:"author_id"`
	CircleID *int64 `json:"circle_id"`
	Keyword  string `json:"keyword"`
	SortBy   string `json:"sort_by"`
	Order    string `json:"order"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// AdminListPostsResponse represents a response with post list in admin panel
type AdminListPostsResponse struct {
	Posts      []*models.Post `json:"posts"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// BatchReviewRequest represents a batch review request
type BatchReviewRequest struct {
	OperatorID int64   `json:"operator_id"`
	PostIDs    []int64 `json:"post_ids"`
	Action     string  `json:"action"` // "approve", "reject"
	Reason     string  `json:"reason"`
	IP         string  `json:"ip"`
}

// ListLogsRequest represents a request to list admin logs
type ListLogsRequest struct {
	OperatorID *int64  `json:"operator_id"`
	Action     string  `json:"action"`
	EntityType string  `json:"entity_type"`
	StartDate  *string `json:"start_date"`
	EndDate    *string `json:"end_date"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
}

// ListLogsResponse represents a response with admin log list
type ListLogsResponse struct {
	Logs       []*models.AdminLog `json:"logs"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// adminService implements AdminService interface
type adminService struct {
	userRepo     repository.UserRepository
	postRepo     repository.PostRepository
	commentRepo  repository.CommentRepository
	adminLogRepo repository.AdminLogRepository
}

// NewAdminService creates a new admin service
func NewAdminService(
	userRepo repository.UserRepository,
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
	adminLogRepo repository.AdminLogRepository,
) AdminService {
	return &adminService{
		userRepo:     userRepo,
		postRepo:     postRepo,
		commentRepo:  commentRepo,
		adminLogRepo: adminLogRepo,
	}
}

// GetDashboard retrieves dashboard metrics
func (s *adminService) GetDashboard(ctx context.Context) (*DashboardResponse, error) {
	// Get total users
	totalUsers, err := s.userRepo.Count(ctx, repository.UserListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	// Get active users
	activeStatus := "active"
	activeUsers, err := s.userRepo.Count(ctx, repository.UserListOptions{
		Status: activeStatus,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}

	// Get total posts
	totalPosts, err := s.postRepo.Count(ctx, repository.PostListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to count total posts: %w", err)
	}

	// Get published posts
	publishedPosts, err := s.postRepo.Count(ctx, repository.PostListOptions{
		Status: "published",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count published posts: %w", err)
	}

	// Get pending posts
	pendingPosts, err := s.postRepo.Count(ctx, repository.PostListOptions{
		Status: "pending",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count pending posts: %w", err)
	}

	// Get total comments
	totalComments, err := s.commentRepo.Count(ctx, repository.CommentListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to count total comments: %w", err)
	}

	// Get daily active users (users who logged in today)
	today := time.Now().Format("2006-01-02")
	dailyActiveUsers, err := s.userRepo.CountByLastLogin(ctx, today)
	if err != nil {
		// If method doesn't exist, set to 0
		dailyActiveUsers = 0
	}

	// Get posts created today
	postsToday, err := s.postRepo.CountByDate(ctx, today)
	if err != nil {
		// If method doesn't exist, set to 0
		postsToday = 0
	}

	return &DashboardResponse{
		TotalUsers:       totalUsers,
		ActiveUsers:      activeUsers,
		TotalPosts:       totalPosts,
		PublishedPosts:   publishedPosts,
		PendingPosts:     pendingPosts,
		TotalComments:    totalComments,
		DailyActiveUsers: dailyActiveUsers,
		PostsToday:       postsToday,
	}, nil
}

// ListUsers retrieves a list of users with filtering and pagination
func (s *adminService) ListUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
	// Set default pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	offset := (req.Page - 1) * req.PageSize

	// Build options
	opts := repository.UserListOptions{
		Status:  req.Status,
		Keyword: req.Keyword,
		SortBy:  req.SortBy,
		Order:   req.Order,
		Limit:   req.PageSize,
		Offset:  offset,
	}

	// Get users
	users, err := s.userRepo.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Get total count
	total, err := s.userRepo.Count(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Remove password hashes from response
	for _, user := range users {
		user.PasswordHash = ""
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &ListUsersResponse{
		Users:      users,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// BanUser bans a user
func (s *adminService) BanUser(ctx context.Context, operatorID, userID int64, reason, ip string) error {
	// Find user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Update user status
	user.Status = "banned"
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}

	// Log action
	details := map[string]interface{}{
		"reason": reason,
	}
	detailsJSON, _ := json.Marshal(details)
	log := &models.AdminLog{
		OperatorID: operatorID,
		Action:     "ban_user",
		EntityType: "user",
		EntityID:   &userID,
		IP:         ip,
		Details:    string(detailsJSON),
		CreatedAt:  time.Now(),
	}
	if err := s.adminLogRepo.Create(ctx, log); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("failed to create admin log: %v\n", err)
	}

	return nil
}

// UnbanUser unbans a user
func (s *adminService) UnbanUser(ctx context.Context, operatorID, userID int64, ip string) error {
	// Find user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Update user status
	user.Status = "active"
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to unban user: %w", err)
	}

	// Log action
	log := &models.AdminLog{
		OperatorID: operatorID,
		Action:     "unban_user",
		EntityType: "user",
		EntityID:   &userID,
		IP:         ip,
		Details:    "{}",
		CreatedAt:  time.Now(),
	}
	if err := s.adminLogRepo.Create(ctx, log); err != nil {
		fmt.Printf("failed to create admin log: %v\n", err)
	}

	return nil
}

// ListPosts retrieves a list of posts with filtering and pagination
func (s *adminService) ListPosts(ctx context.Context, req AdminListPostsRequest) (*AdminListPostsResponse, error) {
	// Set default pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	offset := (req.Page - 1) * req.PageSize

	// Build options
	opts := repository.PostListOptions{
		AuthorID: req.AuthorID,
		CircleID: req.CircleID,
		Status:   req.Status,
		SortBy:   req.SortBy,
		Order:    req.Order,
		Limit:    req.PageSize,
		Offset:   offset,
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

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &AdminListPostsResponse{
		Posts:      posts,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// BatchReviewPosts reviews multiple posts in batch
func (s *adminService) BatchReviewPosts(ctx context.Context, req BatchReviewRequest) error {
	if len(req.PostIDs) == 0 {
		return fmt.Errorf("no post IDs provided")
	}

	// Determine new status based on action
	var newStatus string
	switch req.Action {
	case "approve":
		newStatus = "published"
	case "reject":
		newStatus = "hidden"
	default:
		return fmt.Errorf("invalid action: %s", req.Action)
	}

	// Update each post
	for _, postID := range req.PostIDs {
		if err := s.postRepo.UpdateStatus(ctx, postID, newStatus); err != nil {
			return fmt.Errorf("failed to update post %d: %w", postID, err)
		}

		// Log action for each post
		details := map[string]interface{}{
			"action": req.Action,
			"reason": req.Reason,
		}
		detailsJSON, _ := json.Marshal(details)
		log := &models.AdminLog{
			OperatorID: req.OperatorID,
			Action:     "batch_review_post",
			EntityType: "post",
			EntityID:   &postID,
			IP:         req.IP,
			Details:    string(detailsJSON),
			CreatedAt:  time.Now(),
		}
		if err := s.adminLogRepo.Create(ctx, log); err != nil {
			fmt.Printf("failed to create admin log: %v\n", err)
		}
	}

	return nil
}

// ListLogs retrieves a list of admin logs with filtering and pagination
func (s *adminService) ListLogs(ctx context.Context, req ListLogsRequest) (*ListLogsResponse, error) {
	// Set default pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	offset := (req.Page - 1) * req.PageSize

	// Build options
	opts := repository.AdminLogListOptions{
		OperatorID: req.OperatorID,
		Action:     req.Action,
		EntityType: req.EntityType,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		SortBy:     "created_at",
		Order:      "DESC",
		Limit:      req.PageSize,
		Offset:     offset,
	}

	// Get logs
	logs, err := s.adminLogRepo.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list logs: %w", err)
	}

	// Get total count
	total, err := s.adminLogRepo.Count(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to count logs: %w", err)
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &ListLogsResponse{
		Logs:       logs,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// LogAction logs an administrative action
func (s *adminService) LogAction(ctx context.Context, log *models.AdminLog) error {
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	return s.adminLogRepo.Create(ctx, log)
}
