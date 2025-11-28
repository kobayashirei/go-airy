package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kobayashirei/airy/internal/auth"
	"github.com/kobayashirei/airy/internal/cache"
	"github.com/kobayashirei/airy/internal/config"
	"github.com/kobayashirei/airy/internal/database"
	"github.com/kobayashirei/airy/internal/handler"
	"github.com/kobayashirei/airy/internal/repository"
	"github.com/kobayashirei/airy/internal/service"
)

// SetupAuthRoutes sets up authentication routes
func SetupAuthRoutes(router *gin.RouterGroup, cfg *config.Config) {
	// Initialize dependencies
	db := database.GetDB()
	cacheService := cache.NewCacheService(cache.GetClient(), cfg.Cache.DefaultExpiration)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	emailService := service.NewEmailService()
	authJWTService := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiration)
	jwtService := service.NewJWTService(authJWTService, 7*24*time.Hour) // 7 days for refresh token
	userService := service.NewUserService(userRepo, cacheService, emailService, jwtService)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService)

	// Auth routes
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/activate", authHandler.Activate)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/login/code", authHandler.LoginWithCode)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/resend-activation", authHandler.ResendActivation)
	}
}

// SetupCircleRoutes sets up circle management routes
func SetupCircleRoutes(router *gin.RouterGroup, cfg *config.Config) {
	// Initialize dependencies
	db := database.GetDB()

	// Initialize repositories
	circleRepo := repository.NewCircleRepository(db)
	circleMemberRepo := repository.NewCircleMemberRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	// Initialize services
	circleService := service.NewCircleService(
		circleRepo,
		circleMemberRepo,
		userRepo,
		userRoleRepo,
		roleRepo,
	)

	// Initialize handlers
	circleHandler := handler.NewCircleHandler(circleService)

	// Circle routes
	circleGroup := router.Group("/circles")
	{
		// Public routes
		circleGroup.GET("/:id", circleHandler.GetCircle)
		circleGroup.GET("/:id/members", circleHandler.GetCircleMembers)

		// Protected routes (require authentication)
		// Note: In production, these should use the auth middleware
		circleGroup.POST("", circleHandler.CreateCircle)
		circleGroup.POST("/:id/join", circleHandler.JoinCircle)
		circleGroup.POST("/:id/members/:userId/approve", circleHandler.ApproveMember)
		circleGroup.POST("/:id/moderators", circleHandler.AssignModerator)
	}
}

// SetupNotificationRoutes sets up notification routes
func SetupNotificationRoutes(router *gin.RouterGroup, cfg *config.Config) {
	// Initialize dependencies
	db := database.GetDB()

	// Initialize repositories
	notificationRepo := repository.NewNotificationRepository(db)
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	// Initialize services
	notificationService := service.NewNotificationService(
		notificationRepo,
		userRepo,
		postRepo,
		commentRepo,
	)

	// Initialize handlers
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// Notification routes (all require authentication)
	notificationGroup := router.Group("/notifications")
	{
		// Note: In production, these should use the auth middleware
		notificationGroup.GET("", notificationHandler.GetNotifications)
		notificationGroup.GET("/unread-count", notificationHandler.GetUnreadCount)
		notificationGroup.PUT("/:id/read", notificationHandler.MarkAsRead)
		notificationGroup.PUT("/read-all", notificationHandler.MarkAllAsRead)
	}
}

// SetupMessageRoutes sets up private messaging routes
func SetupMessageRoutes(router *gin.RouterGroup, cfg *config.Config) {
	// Initialize dependencies
	db := database.GetDB()

	// Initialize repositories
	conversationRepo := repository.NewConversationRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	messageService := service.NewMessageService(
		conversationRepo,
		messageRepo,
		userRepo,
	)

	// Initialize handlers
	messageHandler := handler.NewMessageHandler(messageService)

	// Conversation routes (all require authentication)
	conversationGroup := router.Group("/conversations")
	{
		// Note: In production, these should use the auth middleware
		conversationGroup.GET("", messageHandler.GetConversations)
		conversationGroup.POST("", messageHandler.CreateConversation)
		conversationGroup.GET("/:id/messages", messageHandler.GetMessages)
		conversationGroup.POST("/:id/messages", messageHandler.SendMessage)
	}
}

// SetupUserProfileRoutes sets up user profile routes
func SetupUserProfileRoutes(router *gin.RouterGroup, cfg *config.Config) {
	// Initialize dependencies
	db := database.GetDB()
	cacheService := cache.NewCacheService(cache.GetClient(), cfg.Cache.DefaultExpiration)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	userProfileRepo := repository.NewUserProfileRepository(db)
	userStatsRepo := repository.NewUserStatsRepository(db)
	postRepo := repository.NewPostRepository(db)

	// Initialize services
	userProfileService := service.NewUserProfileService(
		userRepo,
		userProfileRepo,
		userStatsRepo,
		postRepo,
		cacheService,
	)

	// Initialize handlers
	userProfileHandler := handler.NewUserProfileHandler(userProfileService)

	// User profile routes
	userGroup := router.Group("/users")
	{
		// Public routes
		userGroup.GET("/:id/profile", userProfileHandler.GetProfile)
		userGroup.GET("/:id/posts", userProfileHandler.GetUserPosts)

		// Protected routes (require authentication)
		// Note: In production, these should use the auth middleware
		userGroup.PUT("/profile", userProfileHandler.UpdateProfile)
	}
}

// SetupAdminRoutes sets up admin management routes
func SetupAdminRoutes(router *gin.RouterGroup, cfg *config.Config) {
	// Initialize dependencies
	db := database.GetDB()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	adminLogRepo := repository.NewAdminLogRepository(db)

	// Initialize services
	adminService := service.NewAdminService(
		userRepo,
		postRepo,
		commentRepo,
		adminLogRepo,
	)

	// Initialize handlers
	adminHandler := handler.NewAdminHandler(adminService)

	// Admin routes (all require authentication and admin permissions)
	adminGroup := router.Group("/admin")
	{
		// Note: In production, these should use the auth middleware and RBAC middleware
		// to ensure only administrators can access these endpoints
		adminGroup.GET("/dashboard", adminHandler.GetDashboard)
		adminGroup.GET("/users", adminHandler.ListUsers)
		adminGroup.POST("/users/:id/ban", adminHandler.BanUser)
		adminGroup.POST("/users/:id/unban", adminHandler.UnbanUser)
		adminGroup.GET("/posts", adminHandler.ListPosts)
		adminGroup.POST("/posts/batch-review", adminHandler.BatchReviewPosts)
		adminGroup.GET("/logs", adminHandler.ListLogs)
	}
}
