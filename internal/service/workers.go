package service

import (
	"fmt"

	"github.com/kobayashirei/airy/internal/mq"
	"go.uber.org/zap"
)

// WorkerManager manages all background workers
type WorkerManager struct {
	messageQueue   mq.MessageQueue
	hotnessWorker  *HotnessWorker
	logger         *zap.Logger
}

// NewWorkerManager creates a new worker manager
func NewWorkerManager(
	messageQueue mq.MessageQueue,
	hotnessWorker *HotnessWorker,
	logger *zap.Logger,
) *WorkerManager {
	return &WorkerManager{
		messageQueue:  messageQueue,
		hotnessWorker: hotnessWorker,
		logger:        logger,
	}
}

// Start starts all workers by subscribing to message queue topics
func (wm *WorkerManager) Start() error {
	if wm.messageQueue == nil {
		wm.logger.Warn("Message queue not configured, workers will not start")
		return nil
	}

	// Subscribe hotness worker to vote events
	if wm.hotnessWorker != nil {
		if err := wm.messageQueue.Subscribe(mq.TopicVoteCreated, wm.hotnessWorker.HandleVoteCreated); err != nil {
			return fmt.Errorf("failed to subscribe to vote.created: %w", err)
		}
		wm.logger.Info("Subscribed hotness worker to vote.created")

		if err := wm.messageQueue.Subscribe(mq.TopicVoteUpdated, wm.hotnessWorker.HandleVoteUpdated); err != nil {
			return fmt.Errorf("failed to subscribe to vote.updated: %w", err)
		}
		wm.logger.Info("Subscribed hotness worker to vote.updated")

		if err := wm.messageQueue.Subscribe(mq.TopicVoteDeleted, wm.hotnessWorker.HandleVoteDeleted); err != nil {
			return fmt.Errorf("failed to subscribe to vote.deleted: %w", err)
		}
		wm.logger.Info("Subscribed hotness worker to vote.deleted")

		// Subscribe to comment events
		if err := wm.messageQueue.Subscribe(mq.TopicCommentCreated, wm.hotnessWorker.HandleCommentCreated); err != nil {
			return fmt.Errorf("failed to subscribe to comment.created: %w", err)
		}
		wm.logger.Info("Subscribed hotness worker to comment.created")

		if err := wm.messageQueue.Subscribe(mq.TopicCommentDeleted, wm.hotnessWorker.HandleCommentDeleted); err != nil {
			return fmt.Errorf("failed to subscribe to comment.deleted: %w", err)
		}
		wm.logger.Info("Subscribed hotness worker to comment.deleted")
	}

	wm.logger.Info("All workers started successfully")
	return nil
}

// Stop stops all workers
func (wm *WorkerManager) Stop() error {
	// Currently, RabbitMQ subscriptions are closed when the connection is closed
	// No additional cleanup needed here
	wm.logger.Info("All workers stopped")
	return nil
}
