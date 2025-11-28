package middleware

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/kobayashirei/airy/internal/config"
	"github.com/kobayashirei/airy/internal/logger"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	cfg := &config.LogConfig{
		Level:  "info",
		Output: "stdout",
	}
	_ = logger.Init(cfg)
	
	os.Exit(m.Run())
}

func TestRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.Use(RequestLogger())
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})
	
	req := httptest.NewRequest("GET", "/test?foo=bar", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequestLoggerWithUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.Use(RequestLogger())
	
	router.GET("/test", func(c *gin.Context) {
		c.Set("userID", int64(123))
		c.JSON(200, gin.H{"message": "ok"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
