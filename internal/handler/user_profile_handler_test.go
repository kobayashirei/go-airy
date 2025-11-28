package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/service"
)

// MockUserProfileService is a mock implementation of UserProfileService
type MockUserProfileService struct {
	mock.Mock
}

func (m *MockUserProfileService) GetProfile(ctx context.Context, userID int64) (*service.UserProfileResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.UserProfileResponse), args.Error(1)
}

func (m *MockUserProfileService) UpdateProfile(ctx context.Context, userID int64, req service.UpdateProfileRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *MockUserProfileService) GetUserPosts(ctx context.Context, userID int64, opts service.UserPostsOptions) (*service.UserPostsResponse, error) {
	args := m.Called(ctx, userID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.UserPostsResponse), args.Error(1)
}

func (m *MockUserProfileService) UpdateFollowerCount(ctx context.Context, userID int64, delta int) error {
	args := m.Called(ctx, userID, delta)
	return args.Error(0)
}

func (m *MockUserProfileService) UpdateFollowingCount(ctx context.Context, userID int64, delta int) error {
	args := m.Called(ctx, userID, delta)
	return args.Error(0)
}

func (m *MockUserProfileService) UpdatePostCount(ctx context.Context, userID int64, delta int) error {
	args := m.Called(ctx, userID, delta)
	return args.Error(0)
}

func (m *MockUserProfileService) UpdateCommentCount(ctx context.Context, userID int64, delta int) error {
	args := m.Called(ctx, userID, delta)
	return args.Error(0)
}

func (m *MockUserProfileService) UpdateVoteReceivedCount(ctx context.Context, userID int64, delta int) error {
	args := m.Called(ctx, userID, delta)
	return args.Error(0)
}

func TestUserProfileHandler_GetProfile_Success(t *testing.T) {
	mockService := new(MockUserProfileService)
	handler := NewUserProfileHandler(mockService)

	router := setupTestRouter()
	router.GET("/users/:id/profile", handler.GetProfile)

	expectedResp := &service.UserProfileResponse{
		User: &models.User{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
			Status:   "active",
		},
		Profile: &models.UserProfile{
			UserID:         1,
			Points:         100,
			Level:          5,
			FollowerCount:  50,
			FollowingCount: 30,
		},
		Stats: &models.UserStats{
			UserID:            1,
			PostCount:         10,
			CommentCount:      25,
			VoteReceivedCount: 100,
		},
	}

	mockService.On("GetProfile", mock.Anything, int64(1)).Return(expectedResp, nil)

	req, _ := http.NewRequest("GET", "/users/1/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserProfileHandler_GetProfile_InvalidUserID(t *testing.T) {
	mockService := new(MockUserProfileService)
	handler := NewUserProfileHandler(mockService)

	router := setupTestRouter()
	router.GET("/users/:id/profile", handler.GetProfile)

	req, _ := http.NewRequest("GET", "/users/invalid/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserProfileHandler_GetProfile_UserNotFound(t *testing.T) {
	mockService := new(MockUserProfileService)
	handler := NewUserProfileHandler(mockService)

	router := setupTestRouter()
	router.GET("/users/:id/profile", handler.GetProfile)

	mockService.On("GetProfile", mock.Anything, int64(999)).Return(nil, service.ErrUserNotFound)

	req, _ := http.NewRequest("GET", "/users/999/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserProfileHandler_UpdateProfile_Success(t *testing.T) {
	mockService := new(MockUserProfileService)
	handler := NewUserProfileHandler(mockService)

	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Next()
	})
	router.PUT("/users/profile", handler.UpdateProfile)

	bio := "Updated bio"
	reqBody := service.UpdateProfileRequest{
		Bio: &bio,
	}

	mockService.On("UpdateProfile", mock.Anything, int64(1), mock.MatchedBy(func(req service.UpdateProfileRequest) bool {
		return req.Bio != nil && *req.Bio == "Updated bio"
	})).Return(nil)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/users/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserProfileHandler_UpdateProfile_Unauthorized(t *testing.T) {
	mockService := new(MockUserProfileService)
	handler := NewUserProfileHandler(mockService)

	router := setupTestRouter()
	router.PUT("/users/profile", handler.UpdateProfile)

	bio := "Updated bio"
	reqBody := service.UpdateProfileRequest{
		Bio: &bio,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/users/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserProfileHandler_GetUserPosts_Success(t *testing.T) {
	mockService := new(MockUserProfileService)
	handler := NewUserProfileHandler(mockService)

	router := setupTestRouter()
	router.GET("/users/:id/posts", handler.GetUserPosts)

	expectedResp := &service.UserPostsResponse{
		Posts: []*models.Post{
			{
				ID:       1,
				Title:    "Test Post",
				AuthorID: 1,
				Status:   "published",
				CreatedAt: time.Now(),
			},
		},
		TotalCount: 1,
		Page:       1,
		Limit:      20,
	}

	mockService.On("GetUserPosts", mock.Anything, int64(1), mock.MatchedBy(func(opts service.UserPostsOptions) bool {
		return opts.Page == 1 && opts.Limit == 20
	})).Return(expectedResp, nil)

	req, _ := http.NewRequest("GET", "/users/1/posts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserProfileHandler_GetUserPosts_WithPagination(t *testing.T) {
	mockService := new(MockUserProfileService)
	handler := NewUserProfileHandler(mockService)

	router := setupTestRouter()
	router.GET("/users/:id/posts", handler.GetUserPosts)

	expectedResp := &service.UserPostsResponse{
		Posts:      []*models.Post{},
		TotalCount: 50,
		Page:       2,
		Limit:      10,
	}

	mockService.On("GetUserPosts", mock.Anything, int64(1), mock.MatchedBy(func(opts service.UserPostsOptions) bool {
		return opts.Page == 2 && opts.Limit == 10
	})).Return(expectedResp, nil)

	req, _ := http.NewRequest("GET", "/users/1/posts?page=2&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserProfileHandler_GetUserPosts_InvalidPage(t *testing.T) {
	mockService := new(MockUserProfileService)
	handler := NewUserProfileHandler(mockService)

	router := setupTestRouter()
	router.GET("/users/:id/posts", handler.GetUserPosts)

	req, _ := http.NewRequest("GET", "/users/1/posts?page=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserProfileHandler_GetUserPosts_InvalidLimit(t *testing.T) {
	mockService := new(MockUserProfileService)
	handler := NewUserProfileHandler(mockService)

	router := setupTestRouter()
	router.GET("/users/:id/posts", handler.GetUserPosts)

	req, _ := http.NewRequest("GET", "/users/1/posts?limit=200", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
