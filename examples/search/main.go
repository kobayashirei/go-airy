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
	log := logger.NewLogger(cfg.Log.Level, cfg.Log.Output)

	// Initialize database
	db, err := database.NewDatabase(&cfg.Database, log)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Elasticsearch client
	esClient, err := search.NewClient(cfg, log)
	if err != nil {
		log.Fatal("Failed to create Elasticsearch client:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB)
	postRepo := repository.NewPostRepository(db.DB)
	circleRepo := repository.NewCircleRepository(db.DB)
	profileRepo := repository.NewUserProfileRepository(db.DB)
	statsRepo := repository.NewUserStatsRepository(db.DB)

	// Initialize search service
	searchService := service.NewSearchService(
		esClient,
		userRepo,
		postRepo,
		circleRepo,
		log,
	)

	// Initialize indices
	ctx := context.Background()
	if err := searchService.InitializeIndices(ctx); err != nil {
		log.Fatal("Failed to initialize search indices:", err)
	}
	fmt.Println("✓ Search indices initialized")

	// Initialize message queue
	mqConfig := &mq.Config{
		URL: cfg.MQ.GetAddr(),
	}
	messageQueue, err := mq.NewRabbitMQ(mqConfig, log)
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
		log,
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

// Example output:
//
// ✓ Search indices initialized
// ✓ Search consumer subscribed to events
//
// --- Searching for posts ---
// Found 5 posts matching 'golang'
// 1. Getting Started with Golang (score: 2.45)
// 2. Advanced Golang Patterns (score: 2.12)
// 3. Golang vs Python Performance (score: 1.89)
// 4. Building APIs with Golang (score: 1.67)
// 5. Golang Concurrency Explained (score: 1.45)
//
// --- Searching for posts in circle 1 ---
// Found 12 posts in circle 1
// 1. Welcome to the Programming Circle
// 2. Weekly Challenge: Fibonacci
// 3. Code Review: REST API Design
// ...
//
// --- Searching for posts with tag 'tutorial' ---
// Found 8 posts with tag 'tutorial'
// 1. Complete Docker Tutorial (hotness: 45.67)
// 2. React Hooks Tutorial (hotness: 38.92)
// 3. SQL Basics Tutorial (hotness: 32.15)
// ...
//
// --- Searching for users ---
// Found 3 users matching 'john'
// 1. john_doe (score: 3.21)
// 2. johnny_dev (score: 2.87)
// 3. john_smith (score: 2.45)
//
// ✓ Search system is running. Press Ctrl+C to exit.
