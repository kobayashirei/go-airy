package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeStrict(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "HTML tags removed",
			input:    "<script>alert('xss')</script>Hello",
			expected: "Hello",
		},
		{
			name:     "nested HTML tags",
			input:    "<div><p>Hello</p></div>",
			expected: "Hello",
		},
		{
			name:     "HTML entities decoded",
			input:    "&lt;script&gt;",
			expected: "",
		},
		{
			name:     "whitespace trimmed",
			input:    "  Hello World  ",
			expected: "Hello World",
		},
		{
			name:     "img tag with onerror",
			input:    `<img src="x" onerror="alert('xss')">`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeStrict(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeUsername(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid username",
			input:    "john_doe",
			expected: "john_doe",
		},
		{
			name:     "username with hyphen",
			input:    "john-doe",
			expected: "john-doe",
		},
		{
			name:     "username with special chars removed",
			input:    "john@doe!",
			expected: "johndoe",
		},
		{
			name:     "username with HTML",
			input:    "<b>john</b>",
			expected: "john",
		},
		{
			name:     "username with spaces removed",
			input:    "john doe",
			expected: "johndoe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeUsername(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeEmail(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid email",
			input:    "test@example.com",
			expected: "test@example.com",
		},
		{
			name:     "email with uppercase",
			input:    "Test@Example.COM",
			expected: "test@example.com",
		},
		{
			name:     "email with whitespace",
			input:    "  test@example.com  ",
			expected: "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeEmail(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizePhone(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid phone",
			input:    "+1234567890",
			expected: "+1234567890",
		},
		{
			name:     "phone with dashes",
			input:    "+1-234-567-890",
			expected: "+1234567890",
		},
		{
			name:     "phone with spaces",
			input:    "+1 234 567 890",
			expected: "+1234567890",
		},
		{
			name:     "phone with parentheses",
			input:    "+1 (234) 567-890",
			expected: "+1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizePhone(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeTitle(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain title",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "title with HTML",
			input:    "<b>Hello</b> World",
			expected: "Hello World",
		},
		{
			name:     "title with multiple spaces",
			input:    "Hello    World",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeUGC(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name        string
		input       string
		shouldAllow bool
	}{
		{
			name:        "script tag removed",
			input:       "<script>alert('xss')</script>",
			shouldAllow: false,
		},
		{
			name:        "safe HTML allowed",
			input:       "<p>Hello <strong>World</strong></p>",
			shouldAllow: true,
		},
		{
			name:        "link allowed",
			input:       `<a href="https://example.com">Link</a>`,
			shouldAllow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeUGC(tt.input)
			if tt.shouldAllow {
				assert.NotEmpty(t, result)
			} else {
				assert.NotContains(t, result, "<script")
			}
		})
	}
}

func TestStripControlCharacters(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no control chars",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "preserves newlines",
			input:    "Hello\nWorld",
			expected: "Hello\nWorld",
		},
		{
			name:     "preserves tabs",
			input:    "Hello\tWorld",
			expected: "Hello\tWorld",
		},
		{
			name:     "removes null byte",
			input:    "Hello\x00World",
			expected: "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.StripControlCharacters(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
