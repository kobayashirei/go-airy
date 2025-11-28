# Search System

This package implements the Elasticsearch integration for the Airy platform.

## Overview

The search system provides full-text search capabilities for posts and users using Elasticsearch. It includes:

- Elasticsearch client wrapper
- Index mappings for posts and users
- Automatic index synchronization via message queue
- Full-text search with filtering and sorting

## Components

### Client (`client.go`)

The `Client` struct wraps the Elasticsearch Go client and provides methods for:

- Creating indices with mappings
- Indexing documents
- Updating documents
- Deleting documents
- Searching documents

### Mappings (`mapping.go`)

Defines Elasticsearch index mappings for:

- **Posts Index**: Stores post data with fields for title, content, author, circle, tags, etc.
- **Users Index**: Stores user data with fields for username, bio, follower count, etc.

### Search Service (`service/search_service.go`)

The `SearchService` interface provides methods for:

- **Post Operations**:
  - `IndexPost`: Index a post in Elasticsearch
  - `UpdatePost`: Update a post in Elasticsearch
  - `DeletePost`: Delete a post from Elasticsearch
  - `SearchPosts`: Search for posts with filtering and sorting

- **User Operations**:
  - `IndexUser`: Index a user in Elasticsearch
  - `UpdateUser`: Update a user in Elasticsearch
  - `DeleteUser`: Delete a user from Elasticsearch
  - `SearchUsers`: Search for users

- **Initialization**:
  - `InitializeIndices`: Create the necessary Elasticsearch indices

### Search Consumer (`service/search_consumer.go`)

The `SearchConsumer` listens to message queue events and automatically synchronizes the search index:

- **Post Events**:
  - `post.published`: Index new posts
  - `post.updated`: Update existing posts
  - `post.deleted`: Remove posts from index

- **User Events**:
  - `user.registered`: Index new users

### Search Handler (`handler/search_handler.go`)

The `SearchHandler` provides HTTP endpoints for search:

- `GET /api/v1/search/posts`: Search for posts
  - Query parameters:
    - `keyword`: Search keyword (searches title, content, summary)
    - `circle_id`: Filter by circle ID
    - `tags`: Filter by tags (comma-separated)
    - `sort_by`: Sort by "time", "hotness", or "relevance"
    - `page`: Page number (default: 1)
    - `page_size`: Results per page (default: 20, max: 100)

- `GET /api/v1/search/users`: Search for users
  - Query parameters:
    - `keyword`: Search keyword (searches username, bio)
    - `page`: Page number (default: 1)
    - `page_size`: Results per page (default: 20, max: 100)

## Configuration

The search system requires the following environment variables:

```env
ES_HOST=localhost
ES_PORT=9200
```

## Usage

### Initialize Search Service

```go
// Create Elasticsearch client
esClient, err := search.NewClient(cfg, log)
if err != nil {
    log.Fatal("Failed to create ES client", err)
}

// Create search service
searchService := service.NewSearchService(
    esClient,
    userRepo,
    postRepo,
    circleRepo,
    log,
)

// Initialize indices
if err := searchService.InitializeIndices(ctx); err != nil {
    log.Fatal("Failed to initialize indices", err)
}
```

### Set Up Search Consumer

```go
// Create search consumer
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
    log.Fatal("Failed to subscribe to events", err)
}
```

### Register Search Routes

```go
// Create search handler
searchHandler := handler.NewSearchHandler(searchService)

// Register routes
api := router.Group("/api/v1")
searchHandler.RegisterRoutes(api)
```

## Search Features

### Post Search

- **Full-text search**: Searches across title (3x weight), content, and summary (2x weight)
- **Filtering**:
  - By circle ID
  - By tags (multiple tags supported)
  - Only published posts are returned
- **Sorting**:
  - By relevance (default): Uses Elasticsearch scoring
  - By time: Sorts by published date (newest first)
  - By hotness: Sorts by hotness score (highest first)
- **Pagination**: Supports page-based pagination

### User Search

- **Full-text search**: Searches across username (3x weight) and bio
- **Filtering**: Only active users are returned
- **Sorting**: By relevance and follower count
- **Pagination**: Supports page-based pagination

## Index Synchronization

The search index is automatically synchronized with the database through message queue events:

1. When a post is published, the `post.published` event triggers indexing
2. When a post is updated, the `post.updated` event triggers re-indexing
3. When a post is deleted, the `post.deleted` event triggers removal from index
4. When a user registers, the `user.registered` event triggers indexing

This ensures eventual consistency between the database and search index.

## Requirements Validation

This implementation validates the following requirements:

- **Requirement 9.1**: Posts are asynchronously synced to Elasticsearch when published
- **Requirement 9.2**: Full-text search is performed using Elasticsearch
- **Requirement 9.3**: Search results include sorting by time and hotness
- **Requirement 9.4**: Filtering by circle ID and tags is supported
- **Requirement 9.5**: Post updates and deletes are synced to Elasticsearch

## Performance Considerations

- **Indexing**: Documents are indexed with `refresh=true` for immediate visibility (can be optimized for bulk operations)
- **Search**: Uses Elasticsearch's built-in caching and optimization
- **Pagination**: Efficient offset-based pagination (consider cursor-based for deep pagination)
- **Filtering**: Uses term and terms queries for efficient filtering

## Future Enhancements

- Add autocomplete/suggestions
- Implement faceted search
- Add search analytics
- Support for more complex queries (date ranges, numeric ranges)
- Bulk indexing for initial data load
- Search result highlighting
