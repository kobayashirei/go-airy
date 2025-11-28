package mq

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Note: These tests require a running RabbitMQ instance
// Skip them if RabbitMQ is not available

func TestRabbitMQ_Integration(t *testing.T) {
	// Skip if RabbitMQ is not available
	config := &Config{
		URL:          "amqp://guest:guest@localhost:5672/",
		ExchangeName: "test.airy.events",
		Logger:       zap.NewNop(),
	}

	mq, err := NewRabbitMQ(config)
	if err != nil {
		t.Skip("RabbitMQ not available, skipping integration tests")
		return
	}
	defer mq.Close()

	t.Run("publishes and receives message", func(t *testing.T) {
		topic := "test.message"
		testMessage := map[string]string{
			"test": "data",
		}

		var received atomic.Bool
		var receivedData []byte

		// Subscribe to topic
		handler := func(ctx context.Context, message []byte) error {
			receivedData = message
			received.Store(true)
			return nil
		}

		err := mq.Subscribe(topic, handler)
		require.NoError(t, err)

		// Give subscriber time to set up
		time.Sleep(100 * time.Millisecond)

		// Publish message
		ctx := context.Background()
		err = mq.Publish(ctx, topic, testMessage)
		require.NoError(t, err)

		// Wait for message to be received
		time.Sleep(500 * time.Millisecond)

		assert.True(t, received.Load())

		// Verify message content
		var receivedMessage map[string]string
		err = json.Unmarshal(receivedData, &receivedMessage)
		require.NoError(t, err)
		assert.Equal(t, testMessage, receivedMessage)
	})

	t.Run("handles multiple subscribers", func(t *testing.T) {
		topic := "test.multiple"
		testMessage := "test data"

		var count1 atomic.Int32
		var count2 atomic.Int32

		handler1 := func(ctx context.Context, message []byte) error {
			count1.Add(1)
			return nil
		}

		handler2 := func(ctx context.Context, message []byte) error {
			count2.Add(1)
			return nil
		}

		err := mq.Subscribe(topic, handler1)
		require.NoError(t, err)

		err = mq.Subscribe(topic, handler2)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Publish message
		ctx := context.Background()
		err = mq.Publish(ctx, topic, testMessage)
		require.NoError(t, err)

		// Wait for messages to be received
		time.Sleep(500 * time.Millisecond)

		// Both handlers should receive the message
		assert.Equal(t, int32(1), count1.Load())
		assert.Equal(t, int32(1), count2.Load())
	})

	t.Run("handles event types", func(t *testing.T) {
		topic := TopicPostPublished

		event := PostPublishedEvent{
			BaseEvent: BaseEvent{
				EventID:   "test-123",
				EventType: TopicPostPublished,
				Timestamp: time.Now(),
			},
			PostID:   1,
			AuthorID: 2,
			Title:    "Test Post",
		}

		var receivedEvent PostPublishedEvent
		var received atomic.Bool

		handler := func(ctx context.Context, message []byte) error {
			err := json.Unmarshal(message, &receivedEvent)
			if err != nil {
				return err
			}
			received.Store(true)
			return nil
		}

		err := mq.Subscribe(topic, handler)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		ctx := context.Background()
		err = mq.Publish(ctx, topic, event)
		require.NoError(t, err)

		time.Sleep(500 * time.Millisecond)

		assert.True(t, received.Load())
		assert.Equal(t, event.PostID, receivedEvent.PostID)
		assert.Equal(t, event.AuthorID, receivedEvent.AuthorID)
		assert.Equal(t, event.Title, receivedEvent.Title)
	})
}

func TestRabbitMQ_Close(t *testing.T) {
	config := &Config{
		URL:          "amqp://guest:guest@localhost:5672/",
		ExchangeName: "test.airy.events",
		Logger:       zap.NewNop(),
	}

	mq, err := NewRabbitMQ(config)
	if err != nil {
		t.Skip("RabbitMQ not available, skipping test")
		return
	}

	err = mq.Close()
	assert.NoError(t, err)

	// Closing again should not error
	err = mq.Close()
	assert.NoError(t, err)

	// Publishing after close should error
	ctx := context.Background()
	err = mq.Publish(ctx, "test", "data")
	assert.Error(t, err)
}

func TestEventTypes(t *testing.T) {
	t.Run("marshals PostPublishedEvent", func(t *testing.T) {
		event := PostPublishedEvent{
			BaseEvent: BaseEvent{
				EventID:   "test-123",
				EventType: TopicPostPublished,
				Timestamp: time.Now(),
			},
			PostID:   1,
			AuthorID: 2,
			Title:    "Test Post",
		}

		data, err := json.Marshal(event)
		require.NoError(t, err)

		var decoded PostPublishedEvent
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, event.PostID, decoded.PostID)
		assert.Equal(t, event.AuthorID, decoded.AuthorID)
		assert.Equal(t, event.Title, decoded.Title)
	})

	t.Run("marshals CommentCreatedEvent", func(t *testing.T) {
		parentID := int64(10)
		event := CommentCreatedEvent{
			BaseEvent: BaseEvent{
				EventID:   "test-456",
				EventType: TopicCommentCreated,
				Timestamp: time.Now(),
			},
			CommentID: 1,
			PostID:    2,
			AuthorID:  3,
			ParentID:  &parentID,
			Content:   "Test comment",
		}

		data, err := json.Marshal(event)
		require.NoError(t, err)

		var decoded CommentCreatedEvent
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, event.CommentID, decoded.CommentID)
		assert.Equal(t, event.PostID, decoded.PostID)
		assert.Equal(t, event.AuthorID, decoded.AuthorID)
		assert.NotNil(t, decoded.ParentID)
		assert.Equal(t, *event.ParentID, *decoded.ParentID)
		assert.Equal(t, event.Content, decoded.Content)
	})

	t.Run("marshals VoteCreatedEvent", func(t *testing.T) {
		event := VoteCreatedEvent{
			BaseEvent: BaseEvent{
				EventID:   "test-789",
				EventType: TopicVoteCreated,
				Timestamp: time.Now(),
			},
			VoteID:     1,
			UserID:     2,
			EntityType: "post",
			EntityID:   3,
			VoteType:   "up",
		}

		data, err := json.Marshal(event)
		require.NoError(t, err)

		var decoded VoteCreatedEvent
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, event.VoteID, decoded.VoteID)
		assert.Equal(t, event.UserID, decoded.UserID)
		assert.Equal(t, event.EntityType, decoded.EntityType)
		assert.Equal(t, event.EntityID, decoded.EntityID)
		assert.Equal(t, event.VoteType, decoded.VoteType)
	})
}
