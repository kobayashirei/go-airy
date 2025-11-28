package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/kobayashirei/airy/internal/cache"
	"github.com/kobayashirei/airy/internal/config"
	"github.com/kobayashirei/airy/internal/database"
	"github.com/kobayashirei/airy/internal/logger"
	"github.com/kobayashirei/airy/internal/middleware"
	"github.com/kobayashirei/airy/internal/response"
	appRouter "github.com/kobayashirei/airy/internal/router"
	"github.com/kobayashirei/airy/internal/version"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(&cfg.Log); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Airy backend server",
		zap.String("project", version.Name),
		zap.String("version", version.Version),
		zap.String("author", version.Author),
		zap.String("github", version.GitHub),
		zap.String("website", version.Website),
		zap.String("mode", cfg.Server.Mode),
	)

	// Initialize database
	if err := database.Init(&cfg.Database); err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.Close()

	// Initialize Redis cache
	if err := cache.Init(&cfg.Redis, &cfg.Cache); err != nil {
		logger.Fatal("Failed to initialize cache", zap.Error(err))
	}
	defer cache.Close()

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create router
	router := setupRouter(cfg)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		if cfg.Server.TLSEnabled {
			logger.Info("Server listening with TLS",
				zap.String("address", addr),
				zap.String("cert", cfg.Server.TLSCertFile),
			)
			if err := srv.ListenAndServeTLS(cfg.Server.TLSCertFile, cfg.Server.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				logger.Fatal("Failed to start TLS server", zap.Error(err))
			}
		} else {
			logger.Info("Server listening", zap.String("address", addr))
			logger.Warn("TLS is disabled. Enable TLS in production for secure communication.")
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("Failed to start server", zap.Error(err))
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// setupRouter configures and returns the Gin router
func setupRouter(cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Apply middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorLogger())
	router.Use(middleware.PrometheusMetrics())
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		ctx := c.Request.Context()
		info := version.Info()
		info["time"] = time.Now().Format(time.RFC3339)

		// Check database health
		if err := database.HealthCheck(ctx); err != nil {
			info["status"] = "unhealthy"
			info["database"] = "error"
			response.Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service is unhealthy", nil)
			return
		}
		info["database"] = "healthy"

		// Check cache health
		if err := cache.HealthCheck(ctx); err != nil {
			info["status"] = "unhealthy"
			info["cache"] = "error"
			response.Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service is unhealthy", nil)
			return
		}
		info["cache"] = "healthy"

		info["status"] = "healthy"
		response.Success(c, info)
	})

	// Version endpoint
	router.GET("/version", func(c *gin.Context) {
		response.Success(c, version.Info())
	})

	// Metrics endpoint for Prometheus
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Ping endpoint
		v1.GET("/ping", func(c *gin.Context) {
			response.Success(c, gin.H{
				"message": "pong",
			})
		})

		// Setup auth routes
		appRouter.SetupAuthRoutes(v1, cfg)

		// Setup circle routes
		appRouter.SetupCircleRoutes(v1, cfg)

		// Setup notification routes
		appRouter.SetupNotificationRoutes(v1, cfg)

		// Setup message routes
		appRouter.SetupMessageRoutes(v1, cfg)

		// Setup user profile routes
		appRouter.SetupUserProfileRoutes(v1, cfg)

		// Setup admin routes
		appRouter.SetupAdminRoutes(v1, cfg)
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "Route not found")
	})

	return router
}
