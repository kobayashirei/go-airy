package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kobayashirei/airy/internal/config"
	"github.com/kobayashirei/airy/internal/database"
	"github.com/kobayashirei/airy/internal/logger"
	"github.com/kobayashirei/airy/internal/mq"
	"github.com/kobayashirei/airy/internal/repository"
	"github.com/kobayashirei/airy/internal/search"
	"github.com/kobayashirei/airy/internal/service"
)

// This example demonstrates how to set up and use the search system
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize logger
	if err := logger.Init(&cfg.Log); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Initialize database
	if err := database.Init(&cfg.Database); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	db := database.GetDB()

	// Initialize Elasticsearch client
	esClient, err := search.NewClient(cfg, logger.Logger)
	if err != nil {
		log.Fatal("Failed to create Elasticsearch client:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	circleRepo := repository.NewCircleRepository(db)
	profileRepo := repository.NewUserProfileRepository(db)
	statsRepo := repository.NewUserStatsRepository(db)

	// Initialize search service
	searchService := service.NewSearchService(
		esClient,
		userRepo,
		postRepo,
		circleRepo,
		logger.Logger,
	)

	// Initialize indices
	ctx := context.Background()
	if err := searchService.InitializeIndices(ctx); err != nil {
		log.Fatal("Failed to initialize search indices:", err)
	}
	fmt.Println("✓ Search indices initialized")

	// Initialize message queue
	mqConfig := &mq.Config{
		URL:    cfg.MQ.GetAddr(),
		Logger: logger.Logger,
	}
	messageQueue, err := mq.NewRabbitMQ(mqConfig)
	if err != nil {
		log.Fatal("Failed to connect to message queue:", err)
	}
	defer messageQueue.Close()

	// Initialize search consumer
	searchConsumer := service.NewSearchConsumer(
		searchService,
		postRepo,
		userRepo,
		profileRepo,
		statsRepo,
		logger.Logger,
	)

	// Subscribe to events
	if err := searchConsumer.Subscribe(messageQueue); err != nil {
		log.Fatal("Failed to subscribe to search events:", err)
	}
	fmt.Println("✓ Search consumer subscribed to events")

	// Example: Search for posts
	fmt.Println("\n--- Searching for posts ---")
	searchQuery := service.SearchQuery{
		Keyword:  "golang",
		SortBy:   "relevance",
		Page:     1,
		PageSize: 10,
	}

	results, err := searchService.SearchPosts(ctx, searchQuery)
	if err != nil {
		log.Fatal("Failed to search posts:", err)
	}

	fmt.Printf("Found %d posts matching 'golang'\n", results.Total)
	for i, result := range results.Results {
		title := result.Data["title"].(string)
		score := result.Score
		fmt.Printf("%d. %s (score: %.2f)\n", i+1, title, score)
	}

	// Example: Search for posts in a specific circle
	fmt.Println("\n--- Searching for posts in circle 1 ---")
	circleID := int64(1)
	circleSearchQuery := service.SearchQuery{
		Keyword:  "",
		CircleID: &circleID,
		SortBy:   "time",
		Page:     1,
		PageSize: 10,
	}

	circleResults, err := searchService.SearchPosts(ctx, circleSearchQuery)
	if err != nil {
		log.Fatal("Failed to search posts in circle:", err)
	}

	fmt.Printf("Found %d posts in circle 1\n", circleResults.Total)
	for i, result := range circleResults.Results {
		title := result.Data["title"].(string)
		fmt.Printf("%d. %s\n", i+1, title)
	}

	// Example: Search for posts with tags
	fmt.Println("\n--- Searching for posts with tag 'tutorial' ---")
	tagSearchQuery := service.SearchQuery{
		Tags:     []string{"tutorial"},
		SortBy:   "hotness",
		Page:     1,
		PageSize: 10,
	}

	tagResults, err := searchService.SearchPosts(ctx, tagSearchQuery)
	if err != nil {
		log.Fatal("Failed to search posts by tag:", err)
	}

	fmt.Printf("Found %d posts with tag 'tutorial'\n", tagResults.Total)
	for i, result := range tagResults.Results {
		title := result.Data["title"].(string)
		hotness := result.Data["hotness_score"].(float64)
		fmt.Printf("%d. %s (hotness: %.2f)\n", i+1, title, hotness)
	}

	// Example: Search for users
	fmt.Println("\n--- Searching for users ---")
	userSearchQuery := service.SearchQuery{
		Keyword:  "john",
		Page:     1,
		PageSize: 10,
	}

	userResults, err := searchService.SearchUsers(ctx, userSearchQuery)
	if err != nil {
		log.Fatal("Failed to search users:", err)
	}

	fmt.Printf("Found %d users matching 'john'\n", userResults.Total)
	for i, result := range userResults.Results {
		username := result.Data["username"].(string)
		score := result.Score
		fmt.Printf("%d. %s (score: %.2f)\n", i+1, username, score)
	}

	// Keep the consumer running
	fmt.Println("\n✓ Search system is running. Press Ctrl+C to exit.")
	select {}
}
