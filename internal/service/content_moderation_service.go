package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrModerationFailed is returned when content moderation fails
	ErrModerationFailed = errors.New("content moderation failed")
)

// ModerationResult represents the result of content moderation
type ModerationResult struct {
	Status   string   `json:"status"`   // "pass", "review", "reject"
	Reason   string   `json:"reason"`   // Reason for the status
	Keywords []string `json:"keywords"` // Flagged keywords if any
}

// ContentModerationService defines the interface for content moderation operations
type ContentModerationService interface {
	CheckContent(ctx context.Context, content string) (*ModerationResult, error)
}

// contentModerationService implements ContentModerationService interface
type contentModerationService struct {
	// In production, this would contain API client for third-party moderation service
	// For now, we'll use a simple keyword-based implementation
	bannedKeywords    []string
	suspiciousKeywords []string
}

// NewContentModerationService creates a new content moderation service
func NewContentModerationService() ContentModerationService {
	return &contentModerationService{
		// Example banned keywords (in production, this would be more comprehensive)
		bannedKeywords: []string{
			"spam", "scam", "illegal", "violence", "hate",
		},
		// Example suspicious keywords that require review
		suspiciousKeywords: []string{
			"advertisement", "promotion", "buy now", "click here",
		},
	}
}

// CheckContent checks content for policy violations
// This is a simplified implementation. In production, this would call a third-party API
// like AWS Comprehend, Google Cloud Natural Language API, or a custom ML model
func (s *contentModerationService) CheckContent(ctx context.Context, content string) (*ModerationResult, error) {
	if content == "" {
		return &ModerationResult{
			Status: "pass",
			Reason: "empty content",
		}, nil
	}

	// Convert to lowercase for case-insensitive matching
	lowerContent := strings.ToLower(content)

	// Check for banned keywords
	for _, keyword := range s.bannedKeywords {
		if strings.Contains(lowerContent, keyword) {
			return &ModerationResult{
				Status:   "reject",
				Reason:   "content contains banned keywords",
				Keywords: []string{keyword},
			}, nil
		}
	}

	// Check for suspicious keywords
	flaggedKeywords := []string{}
	for _, keyword := range s.suspiciousKeywords {
		if strings.Contains(lowerContent, keyword) {
			flaggedKeywords = append(flaggedKeywords, keyword)
		}
	}

	if len(flaggedKeywords) > 0 {
		return &ModerationResult{
			Status:   "review",
			Reason:   "content contains suspicious keywords",
			Keywords: flaggedKeywords,
		}, nil
	}

	// Content passed all checks
	return &ModerationResult{
		Status: "pass",
		Reason: "content passed moderation",
	}, nil
}

// MapModerationStatusToPostStatus maps moderation result status to post status
// This implements the requirement for status mapping (Requirements 4.3, 16.2, 16.3, 16.4)
func MapModerationStatusToPostStatus(moderationStatus string) string {
	switch moderationStatus {
	case "pass":
		return "published"
	case "review":
		return "pending"
	case "reject":
		return "hidden"
	default:
		return "pending" // Default to pending for unknown statuses
	}
}

// IntegrationExample shows how to integrate with a third-party API
// This is a template for production implementation
type ThirdPartyModerationClient struct {
	apiKey  string
	baseURL string
}

// CheckContentWithAPI is an example of how to call a third-party moderation API
func (c *ThirdPartyModerationClient) CheckContentWithAPI(ctx context.Context, content string) (*ModerationResult, error) {
	// In production, this would:
	// 1. Make HTTP request to third-party API
	// 2. Parse the response
	// 3. Map the response to ModerationResult
	// 4. Handle errors and retries
	
	// Example pseudo-code:
	// resp, err := http.Post(c.baseURL + "/moderate", body)
	// if err != nil {
	//     return nil, fmt.Errorf("API request failed: %w", err)
	// }
	// Parse response and return ModerationResult
	
	return nil, fmt.Errorf("not implemented - integrate with actual API")
}
