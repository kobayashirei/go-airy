// Package security provides security utilities for input validation and sanitization
package security

import (
	"html"
	"regexp"
	"strings"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
)

// Sanitizer provides content sanitization functionality
type Sanitizer struct {
	// strictPolicy removes all HTML tags
	strictPolicy *bluemonday.Policy
	// ugcPolicy allows safe HTML for user-generated content
	ugcPolicy *bluemonday.Policy
}

// NewSanitizer creates a new Sanitizer instance
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		strictPolicy: bluemonday.StrictPolicy(),
		ugcPolicy:    bluemonday.UGCPolicy(),
	}
}

// SanitizeStrict removes all HTML tags from the input
// Use this for fields that should never contain HTML (e.g., usernames, titles)
func (s *Sanitizer) SanitizeStrict(input string) string {
	// First, decode any HTML entities to prevent double-encoding
	decoded := html.UnescapeString(input)
	// Remove all HTML tags
	sanitized := s.strictPolicy.Sanitize(decoded)
	// Trim whitespace
	return strings.TrimSpace(sanitized)
}

// SanitizeUGC sanitizes user-generated content while allowing safe HTML
// Use this for content fields that may contain formatted text
func (s *Sanitizer) SanitizeUGC(input string) string {
	return s.ugcPolicy.Sanitize(input)
}

// SanitizeHTML escapes HTML special characters
// Use this when you want to display text as-is without rendering HTML
func (s *Sanitizer) SanitizeHTML(input string) string {
	return html.EscapeString(input)
}

// StripControlCharacters removes control characters from input
func (s *Sanitizer) StripControlCharacters(input string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, input)
}

// SanitizeUsername sanitizes a username
// - Removes HTML tags
// - Strips control characters
// - Trims whitespace
// - Validates allowed characters
func (s *Sanitizer) SanitizeUsername(input string) string {
	// Remove HTML tags
	sanitized := s.SanitizeStrict(input)
	// Strip control characters
	sanitized = s.StripControlCharacters(sanitized)
	// Remove any characters that aren't alphanumeric, underscore, or hyphen
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-]`)
	sanitized = reg.ReplaceAllString(sanitized, "")
	return sanitized
}

// SanitizeEmail sanitizes an email address
func (s *Sanitizer) SanitizeEmail(input string) string {
	// Remove HTML tags
	sanitized := s.SanitizeStrict(input)
	// Strip control characters
	sanitized = s.StripControlCharacters(sanitized)
	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)
	// Trim whitespace
	return strings.TrimSpace(sanitized)
}

// SanitizePhone sanitizes a phone number
func (s *Sanitizer) SanitizePhone(input string) string {
	// Remove HTML tags
	sanitized := s.SanitizeStrict(input)
	// Keep only digits and + sign
	reg := regexp.MustCompile(`[^\d+]`)
	sanitized = reg.ReplaceAllString(sanitized, "")
	return sanitized
}

// SanitizeTitle sanitizes a title field
func (s *Sanitizer) SanitizeTitle(input string) string {
	// Remove HTML tags
	sanitized := s.SanitizeStrict(input)
	// Strip control characters
	sanitized = s.StripControlCharacters(sanitized)
	// Normalize whitespace
	sanitized = normalizeWhitespace(sanitized)
	return sanitized
}

// SanitizeSearchQuery sanitizes a search query
func (s *Sanitizer) SanitizeSearchQuery(input string) string {
	// Remove HTML tags
	sanitized := s.SanitizeStrict(input)
	// Strip control characters
	sanitized = s.StripControlCharacters(sanitized)
	// Escape special characters that could be used for injection
	sanitized = escapeSearchSpecialChars(sanitized)
	return sanitized
}

// normalizeWhitespace replaces multiple whitespace characters with a single space
func normalizeWhitespace(input string) string {
	reg := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(reg.ReplaceAllString(input, " "))
}

// escapeSearchSpecialChars escapes special characters in search queries
func escapeSearchSpecialChars(input string) string {
	// Escape characters that have special meaning in search engines
	specialChars := []string{"\\", "+", "-", "&&", "||", "!", "(", ")", "{", "}", "[", "]", "^", "\"", "~", "*", "?", ":", "/"}
	result := input
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// DefaultSanitizer is the default sanitizer instance
var DefaultSanitizer = NewSanitizer()

// SanitizeStrict is a convenience function using the default sanitizer
func SanitizeStrict(input string) string {
	return DefaultSanitizer.SanitizeStrict(input)
}

// SanitizeUGC is a convenience function using the default sanitizer
func SanitizeUGC(input string) string {
	return DefaultSanitizer.SanitizeUGC(input)
}

// SanitizeHTML is a convenience function using the default sanitizer
func SanitizeHTML(input string) string {
	return DefaultSanitizer.SanitizeHTML(input)
}

// SanitizeUsername is a convenience function using the default sanitizer
func SanitizeUsername(input string) string {
	return DefaultSanitizer.SanitizeUsername(input)
}

// SanitizeEmail is a convenience function using the default sanitizer
func SanitizeEmail(input string) string {
	return DefaultSanitizer.SanitizeEmail(input)
}

// SanitizePhone is a convenience function using the default sanitizer
func SanitizePhone(input string) string {
	return DefaultSanitizer.SanitizePhone(input)
}

// SanitizeTitle is a convenience function using the default sanitizer
func SanitizeTitle(input string) string {
	return DefaultSanitizer.SanitizeTitle(input)
}

// SanitizeSearchQuery is a convenience function using the default sanitizer
func SanitizeSearchQuery(input string) string {
	return DefaultSanitizer.SanitizeSearchQuery(input)
}
