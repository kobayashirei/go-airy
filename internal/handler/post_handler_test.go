package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/service"
)

// MockPostService is a mock implementation of PostService
type MockPostService struct {
	mock.Mock
}

func (m *MockPostService) CreatePost(ctx context.Context, req service.CreatePostRequest) (*models.Post, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostService) GetPost(ctx context.Context, postID int64, userID *int64) (*models.Post, error) {
	args := m.Called(ctx, postID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostService) UpdatePost(ctx context.Context, postID int64, userID int64, req service.UpdatePostRequest) (*models.Post, error) {
	args := m.Called(ctx, postID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostService) DeletePost(ctx context.Context, postID int64, userID int64) error {
	args := m.Called(ctx, postID, userID)
	return args.Error(0)
}

func (m *MockPostService) ListPosts(ctx context.Context, opts service.ListPostsRequest) (*service.ListPostsResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ListPostsResponse), args.Error(1)
}

func TestCreatePost_Success(t *testing.T) {
	mockService := new(MockPostService)
	handler := NewPostHandler(mockService)
	router := setupTestRouter()

	// Setup route
	router.POST("/api/v1/posts", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.CreatePost(c)
	})

	// Setup expectations
	mockService.On("CreatePost", mock.Anything, mock.MatchedBy(func(req service.CreatePostRequest) bool {
		return req.Title == "Test Post" && req.AuthorID == 1
	})).Return(&models.Post{
		ID:       1,
		Title:    "Test Post",
		AuthorID: 1,
		Status:   "published",
	}, nil)

	// Create request
	reqBody := map[string]interface{}{
		"title":   "Test Post",
		"content": "Test content",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreatePost_Unauthorized(t *testing.T) {
	mockService := new(MockPostService)
	handler := NewPostHandler(mockService)
	router := setupTestRouter()

	// Setup route without setting userID
	router.POST("/api/v1/posts", handler.CreatePost)

	// Create request
	reqBody := map[string]interface{}{
		"title":   "Test Post",
		"content": "Test content",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetPost_Success(t *testing.T) {
	mockService := new(MockPostService)
	handler := NewPostHandler(mockService)
	router := setupTestRouter()

	// Setup route
	router.GET("/api/v1/posts/:id", handler.GetPost)

	// Setup expectations
	mockService.On("GetPost", mock.Anything, int64(1), mock.Anything).Return(&models.Post{
		ID:       1,
		Title:    "Test Post",
		AuthorID: 1,
		Status:   "published",
	}, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/posts/1", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetPost_NotFound(t *testing.T) {
	mockService := new(MockPostService)
	handler := NewPostHandler(mockService)
	router := setupTestRouter()

	// Setup route
	router.GET("/api/v1/posts/:id", handler.GetPost)

	// Setup expectations
	mockService.On("GetPost", mock.Anything, int64(999), mock.Anything).Return(nil, service.ErrPostNotFound)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/posts/999", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeletePost_Success(t *testing.T) {
	mockService := new(MockPostService)
	handler := NewPostHandler(mockService)
	router := setupTestRouter()

	// Setup route
	router.DELETE("/api/v1/posts/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.DeletePost(c)
	})

	// Setup expectations
	mockService.On("DeletePost", mock.Anything, int64(1), int64(1)).Return(nil)

	// Create request
	req, _ := http.NewRequest("DELETE", "/api/v1/posts/1", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeletePost_Unauthorized(t *testing.T) {
	mockService := new(MockPostService)
	handler := NewPostHandler(mockService)
	router := setupTestRouter()

	// Setup route
	router.DELETE("/api/v1/posts/:id", func(c *gin.Context) {
		c.Set("userID", int64(2))
		handler.DeletePost(c)
	})

	// Setup expectations - user 2 trying to delete post owned by user 1
	mockService.On("DeletePost", mock.Anything, int64(1), int64(2)).Return(service.ErrUnauthorized)

	// Create request
	req, _ := http.NewRequest("DELETE", "/api/v1/posts/1", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusForbidden, w.Code)
	mockService.AssertExpectations(t)
}

func TestListPosts_Success(t *testing.T) {
	mockService := new(MockPostService)
	handler := NewPostHandler(mockService)
	router := setupTestRouter()

	// Setup route
	router.GET("/api/v1/posts", handler.ListPosts)

	// Setup expectations
	mockService.On("ListPosts", mock.Anything, mock.Anything).Return(&service.ListPostsResponse{
		Posts: []*models.Post{
			{ID: 1, Title: "Post 1"},
			{ID: 2, Title: "Post 2"},
		},
		Total:      2,
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
	}, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/posts?page=1&page_size=20", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdatePost_Success(t *testing.T) {
	mockService := new(MockPostService)
	handler := NewPostHandler(mockService)
	router := setupTestRouter()

	// Setup route
	router.PUT("/api/v1/posts/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.UpdatePost(c)
	})

	// Setup expectations
	title := "Updated Title"
	mockService.On("UpdatePost", mock.Anything, int64(1), int64(1), mock.MatchedBy(func(req service.UpdatePostRequest) bool {
		return req.Title != nil && *req.Title == "Updated Title"
	})).Return(&models.Post{
		ID:       1,
		Title:    "Updated Title",
		AuthorID: 1,
		Status:   "published",
	}, nil)

	// Create request
	reqBody := map[string]interface{}{
		"title": title,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/posts/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdatePost_NotFound(t *testing.T) {
	mockService := new(MockPostService)
	handler := NewPostHandler(mockService)
	router := setupTestRouter()

	// Setup route
	router.PUT("/api/v1/posts/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.UpdatePost(c)
	})

	// Setup expectations
	mockService.On("UpdatePost", mock.Anything, int64(999), int64(1), mock.Anything).Return(nil, service.ErrPostNotFound)

	// Create request
	reqBody := map[string]interface{}{
		"title": "Updated Title",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/posts/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreatePost_ModerationFailed(t *testing.T) {
	mockService := new(MockPostService)
	handler := NewPostHandler(mockService)
	router := setupTestRouter()

	// Setup route
	router.POST("/api/v1/posts", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.CreatePost(c)
	})

	// Setup expectations
	mockService.On("CreatePost", mock.Anything, mock.Anything).Return(nil, service.ErrModerationFailed)

	// Create request
	reqBody := map[string]interface{}{
		"title":   "Test Post",
		"content": "Test content",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetPost_InternalError(t *testing.T) {
	mockService := new(MockPostService)
	handler := NewPostHandler(mockService)
	router := setupTestRouter()

	// Setup route
	router.GET("/api/v1/posts/:id", handler.GetPost)

	// Setup expectations
	mockService.On("GetPost", mock.Anything, int64(1), mock.Anything).Return(nil, errors.New("database error"))

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/posts/1", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
