package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kobayashirei/airy/internal/cache"
	"github.com/kobayashirei/airy/internal/config"
	"github.com/kobayashirei/airy/internal/database"
	"github.com/kobayashirei/airy/internal/logger"
	"github.com/kobayashirei/airy/internal/middleware"
	"github.com/kobayashirei/airy/internal/response"
	appRouter "github.com/kobayashirei/airy/internal/router"
	"github.com/kobayashirei/airy/internal/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
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

	// Initialize database (optional)
	if cfg.Features.EnableDatabase {
		// Resolve init lock path from env to avoid dependency on struct field
		lockPath := os.Getenv("DB_INIT_LOCK_FILE")
		if lockPath == "" {
			lockPath = "./data/db_init.lock"
		}
		// Check init lock
		hasLock, err := database.HasInitLock(lockPath)
		if err != nil {
			logger.Warn("Failed to check DB init lock", zap.Error(err))
		}
		if !hasLock && cfg.Features.AutoCreateDB {
			if err := database.EnsureDatabaseExists(&cfg.Database); err != nil {
				logger.Warn("Failed to ensure database", zap.Error(err))
			} else {
				if err := database.CreateInitLock(lockPath); err != nil {
					logger.Warn("Failed to create DB init lock", zap.Error(err))
				}
			}
		} else if hasLock {
			logger.Info("DB init lock present, skipping ensure", zap.String("lock", lockPath))
		}

		// Read optional flag from env to avoid module cache inconsistencies
		allowStartWithoutDB := strings.EqualFold(os.Getenv("ALLOW_START_WITHOUT_DB"), "true") || os.Getenv("ALLOW_START_WITHOUT_DB") == ""
		if err := database.Init(&cfg.Database); err != nil {
			if allowStartWithoutDB {
				logger.Warn("Failed to initialize database, continuing without DB", zap.Error(err))
			} else {
				logger.Fatal("Failed to initialize database", zap.Error(err))
			}
		}
		// Always run migrations when enabled to apply new changes
		if cfg.Features.AutoMigrate && database.GetDB() != nil {
			sqlLockPath := os.Getenv("DB_SQL_INIT_LOCK_FILE")
			if sqlLockPath == "" {
				sqlLockPath = "./data/db_sql_init.lock"
			}
			hasSqlLock, err := database.HasInitLock(sqlLockPath)
			if err != nil {
				logger.Warn("Failed to check DB SQL init lock", zap.Error(err))
			}
			if hasSqlLock {
				logger.Info("DB SQL init lock present, skipping migrations", zap.String("lock", sqlLockPath))
			} else {
				if cfg.Features.UseGormAutoMigrate {
					if err := database.RunGormAutoMigrate(); err != nil {
						if allowStartWithoutDB {
							logger.Warn("Failed to run GORM auto-migrate", zap.Error(err))
						} else {
							logger.Fatal("Failed to run GORM auto-migrate", zap.Error(err))
						}
					} else {
						if err := database.BootstrapAdmin(cfg); err != nil {
							logger.Warn("Failed to bootstrap admin", zap.Error(err))
						}
						if err := database.CreateInitLock(sqlLockPath); err != nil {
							logger.Warn("Failed to create DB SQL init lock", zap.Error(err))
						} else {
							logger.Info("DB SQL init lock created", zap.String("path", sqlLockPath))
						}
					}
				} else {
					// If current migration state is dirty, force to current version then retry
					if v, dirty, verr := database.MigrationVersion(&cfg.Database, "migrations"); verr == nil && dirty {
						if ferr := database.ForceMigrationVersion(&cfg.Database, "migrations", v); ferr != nil {
							logger.Warn("Failed to force dirty migration version", zap.Error(ferr))
						}
					}
					if err := database.RunMigrations(&cfg.Database, "migrations"); err != nil {
						if allowStartWithoutDB {
							logger.Warn("Failed to run migrations", zap.Error(err))
						} else {
							logger.Fatal("Failed to run migrations", zap.Error(err))
						}
					} else {
						if err := database.BootstrapAdmin(cfg); err != nil {
							logger.Warn("Failed to bootstrap admin", zap.Error(err))
						}
						if err := database.CreateInitLock(sqlLockPath); err != nil {
							logger.Warn("Failed to create DB SQL init lock", zap.Error(err))
						} else {
							logger.Info("DB SQL init lock created", zap.String("path", sqlLockPath))
						}
					}
				}
			}
		}
		if database.GetDB() != nil {
			defer database.Close()
		}
	} else {
		logger.Warn("Database initialization skipped (ENABLE_DATABASE=false)")
	}

	// Initialize Redis cache (optional)
	if cfg.Features.EnableRedis {
		if err := cache.Init(&cfg.Redis, &cfg.Cache); err != nil {
			logger.Fatal("Failed to initialize cache", zap.Error(err))
		}
		defer cache.Close()
	} else {
		logger.Warn("Redis initialization skipped (ENABLE_REDIS=false)")
	}

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
		base := version.Info()
		payload := gin.H{}
		for k, v := range base {
			payload[k] = v
		}
		payload["time"] = time.Now().Format(time.RFC3339)

		if cfg.Features.EnableDatabase {
			allowStartWithoutDB := strings.EqualFold(os.Getenv("ALLOW_START_WITHOUT_DB"), "true") || os.Getenv("ALLOW_START_WITHOUT_DB") == ""
			if database.GetDB() == nil {
				payload["database"] = "unavailable"
				if allowStartWithoutDB {
					payload["status"] = "degraded"
				} else {
					payload["status"] = "unhealthy"
					response.Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service is unhealthy", nil)
					return
				}
			} else {
				if err := database.HealthCheck(ctx); err != nil {
					payload["database"] = "error"
					payload["status"] = "unhealthy"
					response.Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service is unhealthy", nil)
					return
				}
				payload["database"] = "healthy"
				if schema, err := database.CheckSchema(); err == nil {
					if schema.OK {
						payload["schema"] = gin.H{"status": "ok"}
					} else {
						payload["schema"] = gin.H{
							"status":          "partial",
							"missing_tables":  schema.MissingTables,
							"missing_columns": schema.MissingColumns,
						}
						payload["status"] = "degraded"
					}
				}
			}
		} else {
			payload["database"] = "disabled"
		}

		if cfg.Features.EnableRedis {
			if err := cache.HealthCheck(ctx); err != nil {
				payload["status"] = "unhealthy"
				payload["cache"] = "error"
				response.Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service is unhealthy", nil)
				return
			}
			payload["cache"] = "healthy"
		} else {
			payload["cache"] = "disabled"
		}

		if _, ok := payload["status"]; !ok {
			payload["status"] = "healthy"
		}
		response.Success(c, payload)
	})

	router.GET("/health/view", func(c *gin.Context) {
		ctx := c.Request.Context()
		status := "healthy"
		dbStatus := "disabled"
		cacheStatus := "disabled"
		schemaObj := gin.H{"status": "n/a"}
		if cfg.Features.EnableDatabase {
			allow := strings.EqualFold(os.Getenv("ALLOW_START_WITHOUT_DB"), "true") || os.Getenv("ALLOW_START_WITHOUT_DB") == ""
			if database.GetDB() == nil {
				dbStatus = "unavailable"
				if allow {
					status = "degraded"
				} else {
					status = "unhealthy"
				}
			} else {
				if err := database.HealthCheck(ctx); err != nil {
					dbStatus = "error"
					status = "unhealthy"
				} else {
					dbStatus = "healthy"
					if s, err := database.CheckSchema(); err == nil {
						if s.OK {
							schemaObj = gin.H{"status": "ok"}
						} else {
							schemaObj = gin.H{"status": "partial", "missing_tables": s.MissingTables, "missing_columns": s.MissingColumns}
							status = "degraded"
						}
					}
				}
			}
		}
		if cfg.Features.EnableRedis {
			if err := cache.HealthCheck(ctx); err != nil {
				cacheStatus = "error"
				status = "unhealthy"
			} else {
				cacheStatus = "healthy"
			}
		}
		schemaJSON, _ := json.Marshal(schemaObj)
		class := map[string]string{"healthy": "ok", "degraded": "warn", "unhealthy": "bad"}[status]
		html := fmt.Sprintf("<!doctype html><html><head><meta charset='utf-8'><title>Health</title><style>body{font-family:system-ui,Segoe UI,Arial;padding:20px}h1{margin-top:0}.ok{color:#16a34a}.bad{color:#dc2626}.warn{color:#d97706}code,pre{background:#f8fafc;border:1px solid #e5e7eb;border-radius:6px;padding:10px}</style></head><body><h1>Airy Health</h1><p>Status: <strong class='%s'>%s</strong></p><p>Database: <strong>%s</strong></p><p>Cache: <strong>%s</strong></p><h2>Schema</h2><pre>%s</pre></body></html>", class, status, dbStatus, cacheStatus, string(schemaJSON))
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
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
