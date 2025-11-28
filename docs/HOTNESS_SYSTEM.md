# Hotness Ranking System

## Overview

The hotness ranking system calculates and maintains popularity scores for posts based on votes and comments. It supports two popular ranking algorithms: Reddit's hot ranking and Hacker News' ranking.

## Architecture

### Components

1. **HotnessService**: Calculates hotness scores using configurable algorithms
2. **HotnessWorker**: Listens to vote and comment events and triggers recalculation
3. **WorkerManager**: Manages worker lifecycle and message queue subscriptions

### Algorithms

#### Reddit Algorithm

Formula: `log10(max(|score|, 1)) + sign(score) * seconds / 45000`

Where:
- `score = upvotes - downvotes`
- `seconds = time since epoch`
- `45000 ≈ 12.5 hours in seconds`

**Characteristics:**
- Time-weighted: newer posts get a boost
- Logarithmic vote scaling: diminishing returns for more votes
- Considers both positive and negative votes

#### Hacker News Algorithm

Formula: `(score - 1) / (age + 2)^gravity`

Where:
- `score = upvotes - downvotes + 1`
- `age = hours since post creation`
- `gravity = 1.8` (controls decay rate)

**Characteristics:**
- Strong time decay: older posts fall quickly
- Linear vote scaling (within the numerator)
- Gravity factor controls how fast posts age

## Configuration

Set the algorithm in your environment or config file:

```bash
HOTNESS_ALGORITHM=reddit  # or "hackernews"
```

Default: `reddit`

## Usage

### Service Initialization

```go
import (
    "github.com/kobayashirei/airy/internal/service"
    "github.com/kobayashirei/airy/internal/repository"
)

// Create hotness service
hotnessService := service.NewHotnessService(
    postRepo,
    entityCountRepo,
    service.AlgorithmReddit, // or service.AlgorithmHackerNews
)

// Calculate hotness for a post
score, err := hotnessService.CalculateHotness(ctx, post, counts)

// Recalculate and update hotness
newScore, err := hotnessService.RecalculatePostHotness(ctx, postID)
```

### Worker Setup

```go
import (
    "github.com/kobayashirei/airy/internal/service"
    "github.com/kobayashirei/airy/internal/mq"
)

// Create hotness worker
hotnessWorker := service.NewHotnessWorker(
    hotnessService,
    searchClient,
)

// Create worker manager
workerManager := service.NewWorkerManager(
    messageQueue,
    hotnessWorker,
    logger,
)

// Start all workers
if err := workerManager.Start(); err != nil {
    log.Fatal(err)
}

// Stop workers on shutdown
defer workerManager.Stop()
```

## Event Flow

1. User votes on a post or creates a comment
2. Vote/Comment service publishes event to message queue
3. HotnessWorker receives event
4. Worker triggers hotness recalculation
5. New score is saved to database
6. Elasticsearch index is updated with new score

## Message Queue Topics

The hotness worker subscribes to:
- `vote.created` - New vote on a post
- `vote.updated` - Vote changed (up to down or vice versa)
- `vote.deleted` - Vote removed
- `comment.created` - New comment on a post
- `comment.deleted` - Comment removed

## Database Updates

The system updates two locations:
1. **posts.hotness_score** - Main database field
2. **Elasticsearch posts index** - For search and sorting

Updates are atomic and use dedicated repository methods to avoid race conditions.

## Performance Considerations

### Async Processing
- Hotness recalculation happens asynchronously via message queue
- Does not block user requests
- Eventual consistency model

### Caching
- Post cache is not invalidated on hotness updates
- Hotness is primarily used for sorting, not display
- Cache invalidation happens on content updates

### Elasticsearch Sync
- ES updates are best-effort
- Failures are logged but don't fail the operation
- ES can be eventually consistent with database

## Testing

Run hotness service tests:

```bash
go test -v ./internal/service -run TestCalculate
```

## Monitoring

Key metrics to monitor:
- Hotness calculation latency
- Message queue processing rate
- ES sync success rate
- Worker error rate

## Troubleshooting

### Posts not appearing in hot feed
1. Check if hotness worker is running
2. Verify message queue connectivity
3. Check worker logs for errors
4. Verify post has votes or comments

### Hotness scores seem incorrect
1. Verify algorithm configuration
2. Check post timestamps (published_at vs created_at)
3. Verify entity_counts table has correct data
4. Review algorithm parameters

### ES index out of sync
1. Check ES connectivity
2. Review worker logs for ES errors
3. Consider manual reindex if needed
4. Verify ES mapping includes hotness_score field

## Future Enhancements

Potential improvements:
- Custom algorithm parameters via config
- Scheduled batch recalculation for all posts
- Algorithm A/B testing support
- Decay factor configuration
- Comment weight configuration
