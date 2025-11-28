package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kobayashirei/airy/internal/mq"
	"github.com/kobayashirei/airy/internal/taskpool"
	"go.uber.org/zap"
)

// This example demonstrates how to use the task pool and message queue together
// for asynchronous processing in the Airy application

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Example 1: Using Task Pool
	fmt.Println("=== Task Pool Example ===")
	taskPoolExample(logger)

	// Example 2: Using Message Queue (requires RabbitMQ)
	fmt.Println("\n=== Message Queue Example ===")
	messageQueueExample(logger)

	// Example 3: Combining Task Pool and Message Queue
	fmt.Println("\n=== Combined Example ===")
	combinedExample(logger)
}

func taskPoolExample(logger *zap.Logger) {
	// Create task pool
	config := &taskpool.Config{
		Size:   10,
		Logger: logger,
	}

	pool, err := taskpool.NewPool(config)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Release()

	// Submit some tasks
	for i := 0; i < 5; i++ {
		taskNum := i
		err := pool.SubmitFunc(func(ctx context.Context) error {
			logger.Info("executing task",
				zap.Int("task_num", taskNum))
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		if err != nil {
			logger.Error("failed to submit task", zap.Error(err))
		}
	}

	// Wait for all tasks to complete
	pool.Wait()
	logger.Info("all tasks completed")
}

func messageQueueExample(logger *zap.Logger) {
	// Create message queue connection
	mqConfig := &mq.Config{
		URL:          "amqp://guest:guest@localhost:5672/",
		ExchangeName: "example.events",
		Logger:       logger,
	}

	messageQueue, err := mq.NewRabbitMQ(mqConfig)
	if err != nil {
		logger.Warn("RabbitMQ not available, skipping example", zap.Error(err))
		return
	}
	defer messageQueue.Close()

	// Create publisher
	publisher := mq.NewPublisher(messageQueue)

	// Subscribe to post published events
	handler := func(ctx context.Context, message []byte) error {
		var event mq.PostPublishedEvent
		if err := json.Unmarshal(message, &event); err != nil {
			return err
		}

		logger.Info("received post published event",
			zap.Int64("post_id", event.PostID),
			zap.Int64("author_id", event.AuthorID),
			zap.String("title", event.Title))

		return nil
	}

	err = messageQueue.Subscribe(mq.TopicPostPublished, handler)
	if err != nil {
		logger.Error("failed to subscribe", zap.Error(err))
		return
	}

	// Give subscriber time to set up
	time.Sleep(100 * time.Millisecond)

	// Publish some events
	ctx := context.Background()
	for i := 1; i <= 3; i++ {
		err := publisher.PublishPostPublished(
			ctx,
			int64(i),
			int64(100),
			nil,
			fmt.Sprintf("Test Post %d", i),
		)
		if err != nil {
			logger.Error("failed to publish event", zap.Error(err))
		}
	}

	// Wait for events to be processed
	time.Sleep(500 * time.Millisecond)
	logger.Info("all events processed")
}

func combinedExample(logger *zap.Logger) {
	// This example shows how to use task pool and message queue together
	// for a typical use case: processing events asynchronously

	// Create task pool
	poolConfig := &taskpool.Config{
		Size:   10,
		Logger: logger,
	}

	pool, err := taskpool.NewPool(poolConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Release()

	// Create message queue
	mqConfig := &mq.Config{
		URL:          "amqp://guest:guest@localhost:5672/",
		ExchangeName: "example.events",
		Logger:       logger,
	}

	messageQueue, err := mq.NewRabbitMQ(mqConfig)
	if err != nil {
		logger.Warn("RabbitMQ not available, skipping example", zap.Error(err))
		return
	}
	defer messageQueue.Close()

	// Subscribe to events and process them using the task pool
	handler := func(ctx context.Context, message []byte) error {
		var event mq.PostPublishedEvent
		if err := json.Unmarshal(message, &event); err != nil {
			return err
		}

		// Submit event processing to task pool
		return pool.SubmitFunc(func(ctx context.Context) error {
			logger.Info("processing post published event in task pool",
				zap.Int64("post_id", event.PostID),
				zap.String("title", event.Title))

			// Simulate some processing work
			time.Sleep(200 * time.Millisecond)

			// Here you would:
			// 1. Update search index
			// 2. Push to user feeds
			// 3. Send notifications
			// etc.

			logger.Info("finished processing event",
				zap.Int64("post_id", event.PostID))

			return nil
		})
	}

	err = messageQueue.Subscribe(mq.TopicPostPublished, handler)
	if err != nil {
		logger.Error("failed to subscribe", zap.Error(err))
		return
	}

	// Give subscriber time to set up
	time.Sleep(100 * time.Millisecond)

	// Publish events
	publisher := mq.NewPublisher(messageQueue)
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		err := publisher.PublishPostPublished(
			ctx,
			int64(i),
			int64(100),
			nil,
			fmt.Sprintf("Test Post %d", i),
		)
		if err != nil {
			logger.Error("failed to publish event", zap.Error(err))
		}
	}

	// Wait for all tasks to complete
	pool.Wait()
	logger.Info("all events processed through task pool")
}
