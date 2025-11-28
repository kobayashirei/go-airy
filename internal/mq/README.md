# Message Queue Package

This package provides message queue functionality for the Airy application using RabbitMQ.

## Overview

The message queue system enables asynchronous communication between different parts of the application. It uses RabbitMQ with a topic exchange pattern for flexible routing of events.

## Components

### MessageQueue Interface

The core interface that defines message queue operations:

```go
type MessageQueue interface {
    Publish(ctx context.Context, topic string, message interface{}) error
    Subscribe(topic string, handler MessageHandler) error
    Close() error
}
```

### RabbitMQ Implementation

The `RabbitMQ` struct implements the `MessageQueue` interface and provides:

- Automatic reconnection on connection loss
- Topic-based routing using exchanges
- Persistent message delivery
- JSON message serialization
- Error handling and logging

### Event Types

The package defines various event types for different domain events:

#### Post Events
- `PostPublishedEvent` - When a post is published
- `PostUpdatedEvent` - When a post is updated
- `PostDeletedEvent` - When a post is deleted
- `PostVotedEvent` - When a post receives a vote

#### Comment Events
- `CommentCreatedEvent` - When a comment is created
- `CommentDeletedEvent` - When a comment is deleted
- `CommentVotedEvent` - When a comment receives a vote

#### User Events
- `UserFollowedEvent` - When a user follows another user
- `UserUnfollowedEvent` - When a user unfollows another user
- `UserRegisteredEvent` - When a new user registers

#### Circle Events
- `CircleJoinedEvent` - When a user joins a circle
- `CircleLeftEvent` - When a user leaves a circle

#### Vote Events
- `VoteCreatedEvent` - When a vote is created
- `VoteUpdatedEvent` - When a vote is updated
- `VoteDeletedEvent` - When a vote is deleted

### Publisher

The `Publisher` provides convenient methods for publishing events:

```go
publisher := mq.NewPublisher(messageQueue)
err := publisher.PublishPostPublished(ctx, postID, authorID, circleID, title)
```

## Usage

### Creating a Message Queue

```go
import "github.com/kobayashirei/airy/internal/mq"

config := &mq.Config{
    URL:          "amqp://guest:guest@localhost:5672/",
    ExchangeName: "airy.events",
    Logger:       logger,
}

messageQueue, err := mq.NewRabbitMQ(config)
if err != nil {
    log.Fatal(err)
}
defer messageQueue.Close()
```

### Publishing Events

```go
// Using the Publisher helper
publisher := mq.NewPublisher(messageQueue)
err := publisher.PublishPostPublished(ctx, postID, authorID, nil, "My Post Title")

// Or publish directly
event := mq.PostPublishedEvent{
    BaseEvent: mq.BaseEvent{
        EventID:   uuid.New().String(),
        EventType: mq.TopicPostPublished,
        Timestamp: time.Now(),
    },
    PostID:   postID,
    AuthorID: authorID,
    Title:    "My Post Title",
}
err := messageQueue.Publish(ctx, mq.TopicPostPublished, event)
```

### Subscribing to Events

```go
handler := func(ctx context.Context, message []byte) error {
    var event mq.PostPublishedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return err
    }
    
    // Process the event
    log.Printf("Post published: %d by user %d", event.PostID, event.AuthorID)
    
    return nil
}

err := messageQueue.Subscribe(mq.TopicPostPublished, handler)
```

### Topic Patterns

RabbitMQ supports wildcard patterns for subscribing to multiple topics:

- `*` matches exactly one word
- `#` matches zero or more words

Examples:
- `post.*` - All post events
- `*.created` - All creation events
- `#` - All events

```go
// Subscribe to all post events
err := messageQueue.Subscribe("post.*", postEventHandler)

// Subscribe to all events
err := messageQueue.Subscribe("#", allEventsHandler)
```

## Event Flow Examples

### Post Publication Flow

1. User creates a post
2. Service publishes `PostPublishedEvent`
3. Multiple subscribers process the event:
   - Search indexer updates Elasticsearch
   - Feed generator pushes to user feeds
   - Notification service sends notifications

### Comment Creation Flow

1. User creates a comment
2. Service publishes `CommentCreatedEvent`
3. Subscribers process the event:
   - Count updater increments post comment count
   - Notification service notifies post author
   - Mention parser generates notifications for @mentions

### Vote Flow

1. User votes on content
2. Service publishes `VoteCreatedEvent`
3. Subscribers process the event:
   - Count updater updates vote counts
   - Hotness calculator recalculates post score
   - Notification service notifies content author

## Error Handling

The message queue implementation includes:

- Automatic reconnection on connection loss (up to 10 attempts)
- Message acknowledgment for successful processing
- Message requeuing on handler errors
- Panic recovery in handlers
- Structured logging of errors

## Testing

The package includes comprehensive tests:

- Unit tests for event serialization
- Integration tests (require running RabbitMQ)

Run tests:
```bash
# Run all tests (integration tests will skip if RabbitMQ is not available)
go test ./internal/mq/...

# Run with RabbitMQ available
docker run -d --name rabbitmq -p 5672:5672 rabbitmq:3-management
go test ./internal/mq/...
```

## Configuration

Environment variables:
- `MQ_HOST` - RabbitMQ host (default: localhost)
- `MQ_PORT` - RabbitMQ port (default: 5672)
- `MQ_USER` - RabbitMQ user (default: guest)
- `MQ_PASSWORD` - RabbitMQ password (default: guest)

## Best Practices

1. **Idempotency**: Design event handlers to be idempotent, as messages may be delivered more than once
2. **Error Handling**: Always return errors from handlers to trigger message requeuing
3. **Timeouts**: Use context timeouts in handlers to prevent blocking
4. **Logging**: Log all event processing for debugging and monitoring
5. **Event Versioning**: Include version information in events for backward compatibility
6. **Dead Letter Queues**: Configure DLQs for messages that fail repeatedly

## Future Enhancements

- Support for Kafka as an alternative message queue
- Message priority support
- Delayed message delivery
- Message batching for high-throughput scenarios
- Event replay functionality
- Schema validation for events
