package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/kobayashirei/airy/internal/config"
	"go.uber.org/zap"
)

// Client wraps Elasticsearch client
type Client struct {
	es  *elasticsearch.Client
	log *zap.Logger
}

// NewClient creates a new Elasticsearch client
func NewClient(cfg *config.Config, log *zap.Logger) (*Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.ES.GetAddr()},
	}

	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// Test connection
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch returned error: %s", res.String())
	}

	log.Info("Connected to Elasticsearch")

	return &Client{
		es:  es,
		log: log,
	}, nil
}

// CreateIndex creates an index with the given mapping
func (c *Client) CreateIndex(ctx context.Context, indexName string, mapping string) error {
	req := esapi.IndicesCreateRequest{
		Index: indexName,
		Body:  strings.NewReader(mapping),
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// Ignore "already exists" error
		if res.StatusCode == 400 {
			var e map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&e); err == nil {
				if errType, ok := e["error"].(map[string]interface{})["type"].(string); ok {
					if errType == "resource_already_exists_exception" {
						c.log.Info("Index already exists", zap.String("index", indexName))
						return nil
					}
				}
			}
		}
		return fmt.Errorf("failed to create index: %s", res.String())
	}

	c.log.Info("Created index", zap.String("index", indexName))
	return nil
}

// IndexExists checks if an index exists
func (c *Client) IndexExists(ctx context.Context, indexName string) (bool, error) {
	req := esapi.IndicesExistsRequest{
		Index: []string{indexName},
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// Index indexes a document
func (c *Client) Index(ctx context.Context, indexName, documentID string, document interface{}) error {
	data, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       strings.NewReader(string(data)),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index document: %s", res.String())
	}

	return nil
}

// Update updates a document
func (c *Client) Update(ctx context.Context, indexName, documentID string, document interface{}) error {
	data, err := json.Marshal(map[string]interface{}{
		"doc": document,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.UpdateRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       strings.NewReader(string(data)),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to update document: %s", res.String())
	}

	return nil
}

// Delete deletes a document
func (c *Client) Delete(ctx context.Context, indexName, documentID string) error {
	req := esapi.DeleteRequest{
		Index:      indexName,
		DocumentID: documentID,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("failed to delete document: %s", res.String())
	}

	return nil
}

// Search performs a search query
func (c *Client) Search(ctx context.Context, indexName string, query map[string]interface{}) (*SearchResponse, error) {
	data, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{indexName},
		Body:  strings.NewReader(string(data)),
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search returned error: %s", res.String())
	}

	var searchRes SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&searchRes); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &searchRes, nil
}

// SearchResponse represents Elasticsearch search response
type SearchResponse struct {
	Took int64 `json:"took"`
	Hits struct {
		Total struct {
			Value    int64  `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		Hits []Hit `json:"hits"`
	} `json:"hits"`
}

// Hit represents a search result hit
type Hit struct {
	Index  string                 `json:"_index"`
	ID     string                 `json:"_id"`
	Score  float64                `json:"_score"`
	Source map[string]interface{} `json:"_source"`
}

// Close closes the Elasticsearch client
func (c *Client) Close() error {
	// Elasticsearch v8 client doesn't have a Close method
	return nil
}

// UpdatePostHotnessScore updates only the hotness score field for a post in the index
func (c *Client) UpdatePostHotnessScore(ctx context.Context, postID int64, hotnessScore float64) error {
	documentID := fmt.Sprintf("%d", postID)
	
	// Partial update with just the hotness_score field
	updateDoc := map[string]interface{}{
		"hotness_score": hotnessScore,
	}
	
	data, err := json.Marshal(map[string]interface{}{
		"doc": updateDoc,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal update: %w", err)
	}
	
	req := esapi.UpdateRequest{
		Index:      "posts",
		DocumentID: documentID,
		Body:       strings.NewReader(string(data)),
		Refresh:    "true",
	}
	
	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to update hotness: %w", err)
	}
	defer res.Body.Close()
	
	if res.IsError() {
		// If document doesn't exist, that's okay - it might not be indexed yet
		if res.StatusCode == 404 {
			c.log.Debug("Post not found in index", zap.Int64("post_id", postID))
			return nil
		}
		return fmt.Errorf("failed to update hotness: %s", res.String())
	}
	
	return nil
}
