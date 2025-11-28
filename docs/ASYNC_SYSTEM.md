# Asynchronous Task System

This document describes the asynchronous task processing system in the Airy application, which consists of two main components: the Task Pool and the Message Queue.

## Overview

The async system enables the application to:
- Process compute-intensive tasks without blocking HTTP requests
- Handle IO-intensive operations asynchronously
- Decouple components through event-driven architecture
- Scale horizontally by distributing work across multiple workers
- Improve response times and system throughput

## Architecture

```
┌─────────────┐
│   Handler   │
└──────┬──────┘
       │
       ├──────────────┐
       │              │
       ▼              ▼
┌─────────────┐  ┌──────────────┐
│  Task Pool  │  │ Message Queue│
└──────┬──────┘  └──────┬───────┘
       │                │
       ▼                ▼
┌─────────────┐  ┌──────────────┐
│   Worker    │  │  Subscriber  │
└─────────────┘  └──────────────┘
```

## Components

### 1. Task Pool (ants)

The task pool manages a pool of goroutines for executing tasks efficiently.

**Location**: `internal/taskpool/`

**Key Features**:
- Fixed-size goroutine pool to prevent resource exhaustion
- Task submission with error handling
- Panic recovery
- Metrics (running, free, waiting goroutines)
- Graceful shutdown with timeout
- Context cancellation support

**When to Use**:
- CPU-intensive computations
- Short-lived tasks (< 5 minutes)
- Tasks that don't require persistence
- Tasks that should complete before server shutdown

**Example**:
```go
pool, _ := taskpool.NewPool(&taskpool.Config{
    Size: 1000,
    Logger: logger,
})
defer pool.Release()

// Submit a task
pool.SubmitFunc(func(ctx context.Context) error {
    // Do some work
    return nil
})
```

### 2. Message Queue (RabbitMQ)

The message queue provides reliable, persistent message delivery between components.

**Location**: `internal/mq/`

**Key Features**:
- Topic-based routing
- Persistent message delivery
- Automatic reconnection
- Multiple subscribers per topic
- Message acknowledgment and requeuing
- JSON serialization

**When to Use**:
- Long-running tasks
- Tasks that must survive server restarts
- Tasks that need to be distributed across multiple servers
- Event-driven workflows
- Tasks that require guaranteed delivery

**Example**:
```go
mq, _ := mq.NewRabbitMQ(&mq.Config{
    URL: "amqp://guest:guest@localhost:5672/",
    ExchangeName: "airy.events",
    Logger: logger,
})
defer mq.Close()

// Publish an event
publisher := mq.NewPublisher(mq)
publisher.PublishPostPublished(ctx, postID, authorID, nil, title)

// Subscribe to events
mq.Subscribe(mq.TopicPostPublished, func(ctx context.Context, msg []byte) error {
    // Process event
    return nil
})
```

## Use Cases

### Post Publication Flow

When a user publishes a post, multiple async tasks are triggered:

```go
// In PostService.Create()
func (s *PostService) Create(ctx context.Context, post *Post) error {
    // 1. Save post to database
    if err := s.repo.Create(ctx, post); err != nil {
        return err
    }

    // 2. Publish event to message queue
    s.publisher.PublishPostPublished(ctx, post.ID, post.AuthorID, post.CircleID, post.Title)

    return nil
}

// Subscribers process the event asynchronously:

// Subscriber 1: Update search index
mq.Subscribe(mq.TopicPostPublished, func(ctx context.Context, msg []byte) error {
    var event mq.PostPublishedEvent
    json.Unmarshal(msg, &event)
    
    // Submit to task pool for processing
    return pool.SubmitFunc(func(ctx context.Context) error {
        return searchService.IndexPost(ctx, event.PostID)
    })
})

// Subscriber 2: Update user feeds
mq.Subscribe(mq.TopicPostPublished, func(ctx context.Context, msg []byte) error {
    var event mq.PostPublishedEvent
    json.Unmarshal(msg, &event)
    
    return pool.SubmitFunc(func(ctx context.Context) error {
        return feedService.PushToFollowers(ctx, event.PostID, event.AuthorID)
    })
})

// Subscriber 3: Send notifications
mq.Subscribe(mq.TopicPostPublished, func(ctx context.Context, msg []byte) error {
    var event mq.PostPublishedEvent
    json.Unmarshal(msg, &event)
    
    return pool.SubmitFunc(func(ctx context.Context) error {
        return notificationService.NotifyFollowers(ctx, event.PostID, event.AuthorID)
    })
})
```

### Comment Creation Flow

```go
// In CommentService.Create()
func (s *CommentService) Create(ctx context.Context, comment *Comment) error {
    // 1. Save comment to database
    if err := s.repo.Create(ctx, comment); err != nil {
        return err
    }

    // 2. Publish event
    s.publisher.PublishCommentCreated(
        ctx,
        comment.ID,
        comment.PostID,
        comment.AuthorID,
        comment.ParentID,
        comment.Content,
    )

    return nil
}

// Subscribers:
// - Update post comment count
// - Notify post author
// - Notify parent comment author
// - Parse @mentions and notify mentioned users
```

### Vote Processing Flow

```go
// In VoteService.Vote()
func (s *VoteService) Vote(ctx context.Context, vote *Vote) error {
    // 1. Create or update vote (idempotent)
    if err := s.repo.Upsert(ctx, vote); err != nil {
        return err
    }

    // 2. Publish event
    s.publisher.PublishVoteCreated(
        ctx,
        vote.ID,
        vote.UserID,
        vote.EntityType,
        vote.EntityID,
        vote.VoteType,
    )

    return nil
}

// Subscribers:
// - Update vote counts in entity_counts table
// - Recalculate hotness score
// - Notify content author (if not self-vote)
```

## Event Types

### Post Events
- `post.published` - Post is published
- `post.updated` - Post is updated
- `post.deleted` - Post is deleted
- `post.voted` - Post receives a vote

### Comment Events
- `comment.created` - Comment is created
- `comment.deleted` - Comment is deleted
- `comment.voted` - Comment receives a vote

### User Events
- `user.followed` - User follows another user
- `user.unfollowed` - User unfollows another user
- `user.registered` - New user registers

### Circle Events
- `circle.joined` - User joins a circle
- `circle.left` - User leaves a circle

### Vote Events
- `vote.created` - Vote is created
- `vote.updated` - Vote is updated
- `vote.deleted` - Vote is deleted

## Configuration

### Task Pool Configuration

Environment variables:
- `GOROUTINE_POOL_SIZE` - Maximum number of goroutines (default: 10000)

Code configuration:
```go
config := &taskpool.Config{
    Size:             10000,              // Pool size
    ExpiryDuration:   10 * time.Second,   // Goroutine expiry time
    PreAlloc:         false,              // Pre-allocate goroutines
    MaxBlockingTasks: 0,                  // Max blocking tasks (0 = unlimited)
    Nonblocking:      false,              // Nonblocking mode
    Logger:           logger,             // Logger instance
}
```

### Message Queue Configuration

Environment variables:
- `MQ_HOST` - RabbitMQ host (default: localhost)
- `MQ_PORT` - RabbitMQ port (default: 5672)
- `MQ_USER` - RabbitMQ user (default: guest)
- `MQ_PASSWORD` - RabbitMQ password (default: guest)

Code configuration:
```go
config := &mq.Config{
    URL:          "amqp://guest:guest@localhost:5672/",
    ExchangeName: "airy.events",
    Logger:       logger,
}
```

## Best Practices

### Task Pool

1. **Task Size**: Keep tasks small and focused (< 5 minutes)
2. **Error Handling**: Always return errors from tasks for logging
3. **Context**: Respect context cancellation in long-running tasks
4. **Resource Cleanup**: Use defer for cleanup in tasks
5. **Pool Size**: Size the pool based on CPU cores and workload

### Message Queue

1. **Idempotency**: Design handlers to be idempotent (messages may be delivered multiple times)
2. **Error Handling**: Return errors to trigger message requeuing
3. **Timeouts**: Use context timeouts to prevent blocking
4. **Event Schema**: Keep events backward compatible
5. **Logging**: Log all event processing for debugging
6. **Dead Letter Queues**: Configure DLQs for failed messages

### Combined Usage

1. **Separation of Concerns**: Use MQ for distribution, task pool for execution
2. **Backpressure**: Task pool provides natural backpressure for MQ consumers
3. **Monitoring**: Monitor both pool metrics and queue depths
4. **Graceful Shutdown**: Wait for task pool before closing MQ connection

## Monitoring

### Task Pool Metrics

```go
pool.Running()  // Number of running goroutines
pool.Free()     // Number of available goroutines
pool.Waiting()  // Number of waiting tasks
pool.Cap()      // Pool capacity
```

### Message Queue Metrics

Monitor via RabbitMQ Management UI:
- Queue depth
- Message rate (publish/deliver)
- Consumer count
- Unacknowledged messages

## Error Handling

### Task Pool Errors

- Task errors are logged but don't stop the pool
- Panics are recovered and logged
- Failed tasks are not retried automatically

### Message Queue Errors

- Handler errors trigger message requeuing
- Connection errors trigger automatic reconnection
- Failed messages can be routed to dead letter queues

## Testing

### Task Pool Tests

```bash
go test ./internal/taskpool/...
```

### Message Queue Tests

```bash
# Start RabbitMQ
docker run -d --name rabbitmq -p 5672:5672 rabbitmq:3-management

# Run tests
go test ./internal/mq/...
```

## Examples

See `examples/async_example.go` for complete working examples of:
- Using task pool alone
- Using message queue alone
- Combining task pool and message queue

## Troubleshooting

### Task Pool Issues

**Problem**: Tasks not executing
- Check pool is not closed
- Check pool size is sufficient
- Check for panics in tasks

**Problem**: High memory usage
- Reduce pool size
- Check for goroutine leaks
- Ensure tasks complete promptly

### Message Queue Issues

**Problem**: Messages not being delivered
- Check RabbitMQ is running
- Check connection URL is correct
- Check exchange and queue bindings

**Problem**: Messages being requeued repeatedly
- Check handler is not returning errors unnecessarily
- Check for infinite error loops
- Configure dead letter queue

**Problem**: Connection drops
- Check network stability
- Check RabbitMQ resource limits
- Monitor reconnection attempts

## Future Enhancements

- [ ] Support for Kafka as alternative message queue
- [ ] Message priority support
- [ ] Delayed message delivery
- [ ] Message batching
- [ ] Event replay functionality
- [ ] Schema validation for events
- [ ] Distributed tracing integration
- [ ] Metrics export to Prometheus
