package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/service"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, req service.RegisterRequest) (*service.RegisterResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.RegisterResponse), args.Error(1)
}

func (m *MockUserService) Activate(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockUserService) Login(ctx context.Context, req service.LoginRequest) (*service.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.LoginResponse), args.Error(1)
}

func (m *MockUserService) LoginWithCode(ctx context.Context, req service.LoginWithCodeRequest) (*service.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.LoginResponse), args.Error(1)
}

func (m *MockUserService) RefreshToken(ctx context.Context, token string) (*service.RefreshTokenResponse, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.RefreshTokenResponse), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Add request_id to context for response package
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		c.Next()
	})
	return router
}

func TestAuthHandler_Register_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.POST("/register", handler.Register)

	reqBody := service.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedResp := &service.RegisterResponse{
		UserID:  1,
		Message: "Registration successful",
	}

	mockService.On("Register", mock.Anything, mock.MatchedBy(func(req service.RegisterRequest) bool {
		return req.Username == "testuser" && req.Email == "test@example.com"
	})).Return(expectedResp, nil)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidEmail(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.POST("/register", handler.Register)

	reqBody := service.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockService.On("Register", mock.Anything, mock.Anything).Return(nil, service.ErrInvalidEmail)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.POST("/login", handler.Login)

	reqBody := service.LoginRequest{
		Identifier: "test@example.com",
		Password:   "password123",
	}

	expectedResp := &service.LoginResponse{
		Token:        "test-token",
		RefreshToken: "test-refresh-token",
		User: &models.User{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
		},
	}

	mockService.On("Login", mock.Anything, mock.MatchedBy(func(req service.LoginRequest) bool {
		return req.Identifier == "test@example.com" && req.Password == "password123"
	})).Return(expectedResp, nil)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.POST("/login", handler.Login)

	reqBody := service.LoginRequest{
		Identifier: "test@example.com",
		Password:   "wrongpassword",
	}

	mockService.On("Login", mock.Anything, mock.Anything).Return(nil, service.ErrInvalidCredentials)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Activate_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.POST("/activate", handler.Activate)

	reqBody := map[string]string{
		"token": "test-activation-token",
	}

	mockService.On("Activate", mock.Anything, "test-activation-token").Return(nil)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/activate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Activate_InvalidToken(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewAuthHandler(mockService)

	router := setupTestRouter()
	router.POST("/activate", handler.Activate)

	reqBody := map[string]string{
		"token": "invalid-token",
	}

	mockService.On("Activate", mock.Anything, "invalid-token").Return(service.ErrInvalidToken)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/activate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}
