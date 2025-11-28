package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kobayashirei/airy/internal/mq"
	"github.com/kobayashirei/airy/internal/search"
)

// HotnessWorker handles hotness recalculation events
type HotnessWorker struct {
	hotnessService HotnessService
	searchClient   *search.Client
}

// NewHotnessWorker creates a new hotness worker
// Implements Requirements 12.2, 12.4
func NewHotnessWorker(
	hotnessService HotnessService,
	searchClient *search.Client,
) *HotnessWorker {
	return &HotnessWorker{
		hotnessService: hotnessService,
		searchClient:   searchClient,
	}
}

// HandleVoteCreated handles vote created events and recalculates hotness
// Implements Requirement 12.2
func (w *HotnessWorker) HandleVoteCreated(ctx context.Context, message []byte) error {
	var event mq.VoteCreatedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal vote created event: %w", err)
	}

	// Only recalculate for post votes
	if event.EntityType != "post" {
		return nil
	}

	return w.recalculateAndSync(ctx, event.EntityID)
}

// HandleVoteUpdated handles vote updated events and recalculates hotness
// Implements Requirement 12.2
func (w *HotnessWorker) HandleVoteUpdated(ctx context.Context, message []byte) error {
	var event mq.VoteUpdatedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal vote updated event: %w", err)
	}

	// Only recalculate for post votes
	if event.EntityType != "post" {
		return nil
	}

	return w.recalculateAndSync(ctx, event.EntityID)
}

// HandleVoteDeleted handles vote deleted events and recalculates hotness
// Implements Requirement 12.2
func (w *HotnessWorker) HandleVoteDeleted(ctx context.Context, message []byte) error {
	var event mq.VoteDeletedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal vote deleted event: %w", err)
	}

	// Only recalculate for post votes
	if event.EntityType != "post" {
		return nil
	}

	return w.recalculateAndSync(ctx, event.EntityID)
}

// HandleCommentCreated handles comment created events and recalculates hotness
// Implements Requirement 12.2
func (w *HotnessWorker) HandleCommentCreated(ctx context.Context, message []byte) error {
	var event mq.CommentCreatedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal comment created event: %w", err)
	}

	return w.recalculateAndSync(ctx, event.PostID)
}

// HandleCommentDeleted handles comment deleted events and recalculates hotness
// Implements Requirement 12.2
func (w *HotnessWorker) HandleCommentDeleted(ctx context.Context, message []byte) error {
	var event mq.CommentDeletedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal comment deleted event: %w", err)
	}

	return w.recalculateAndSync(ctx, event.PostID)
}

// recalculateAndSync recalculates hotness and syncs to Elasticsearch
// Implements Requirements 12.2, 12.4
func (w *HotnessWorker) recalculateAndSync(ctx context.Context, postID int64) error {
	// Recalculate hotness score in database
	newScore, err := w.hotnessService.RecalculatePostHotness(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to recalculate hotness: %w", err)
	}

	// Update Elasticsearch index with new hotness score
	if w.searchClient != nil {
		if err := w.searchClient.UpdatePostHotnessScore(ctx, postID, newScore); err != nil {
			// Log error but don't fail the operation
			// ES sync can be eventually consistent
			fmt.Printf("failed to update ES hotness for post %d: %v\n", postID, err)
		}
	}

	return nil
}
