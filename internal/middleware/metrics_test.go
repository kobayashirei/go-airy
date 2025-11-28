package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPrometheusMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset metrics before test
	httpRequestsTotal.Reset()
	httpRequestDuration.Reset()
	httpErrorsTotal.Reset()

	router := gin.New()
	router.Use(PrometheusMetrics())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// Make a request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check that metrics were recorded
	count := testutil.CollectAndCount(httpRequestsTotal)
	if count == 0 {
		t.Error("Expected httpRequestsTotal to be recorded")
	}

	count = testutil.CollectAndCount(httpRequestDuration)
	if count == 0 {
		t.Error("Expected httpRequestDuration to be recorded")
	}
}

func TestPrometheusMetricsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset metrics before test
	httpRequestsTotal.Reset()
	httpErrorsTotal.Reset()

	router := gin.New()
	router.Use(PrometheusMetrics())

	router.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "internal error"})
	})

	// Make a request
	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	// Check that error metrics were recorded
	count := testutil.CollectAndCount(httpErrorsTotal)
	if count == 0 {
		t.Error("Expected httpErrorsTotal to be recorded")
	}
}

func TestPrometheusMetricsInFlight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset gauge
	httpRequestsInFlight.Set(0)

	router := gin.New()
	router.Use(PrometheusMetrics())

	router.GET("/test", func(c *gin.Context) {
		// Check that in-flight gauge is incremented
		value := testutil.ToFloat64(httpRequestsInFlight)
		if value != 1 {
			t.Errorf("Expected in-flight requests to be 1, got %f", value)
		}
		c.JSON(200, gin.H{"message": "ok"})
	})

	// Make a request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// After request completes, in-flight should be back to 0
	value := testutil.ToFloat64(httpRequestsInFlight)
	if value != 0 {
		t.Errorf("Expected in-flight requests to be 0 after completion, got %f", value)
	}
}

func TestPrometheusMetrics404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset metrics before test
	httpRequestsTotal.Reset()
	httpErrorsTotal.Reset()

	router := gin.New()
	router.Use(PrometheusMetrics())

	// No routes defined, so any request will be 404

	// Make a request to non-existent route
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	// Check that metrics were recorded for 404
	count := testutil.CollectAndCount(httpRequestsTotal)
	if count == 0 {
		t.Error("Expected httpRequestsTotal to be recorded for 404")
	}

	count = testutil.CollectAndCount(httpErrorsTotal)
	if count == 0 {
		t.Error("Expected httpErrorsTotal to be recorded for 404")
	}
}

func TestPrometheusMetricsMultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset metrics before test
	httpRequestsTotal.Reset()

	router := gin.New()
	router.Use(PrometheusMetrics())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// Make multiple requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Check that counter increased
	counter := httpRequestsTotal.WithLabelValues("GET", "/test", "200")
	value := testutil.ToFloat64(counter)
	if value != 5 {
		t.Errorf("Expected 5 requests to be counted, got %f", value)
	}
}


