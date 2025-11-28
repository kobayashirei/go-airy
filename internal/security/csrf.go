package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

// CSRF errors
var (
	ErrInvalidCSRFToken = errors.New("invalid CSRF token")
	ErrMissingCSRFToken = errors.New("missing CSRF token")
	ErrCSRFTokenExpired = errors.New("CSRF token expired")
)

// CSRFConfig holds configuration for CSRF protection
type CSRFConfig struct {
	// TokenLength is the length of the CSRF token in bytes
	TokenLength int
	// TokenExpiration is how long a token is valid
	TokenExpiration time.Duration
	// CookieName is the name of the cookie storing the CSRF token
	CookieName string
	// HeaderName is the name of the header containing the CSRF token
	HeaderName string
	// FormFieldName is the name of the form field containing the CSRF token
	FormFieldName string
	// Secure sets the Secure flag on the cookie
	Secure bool
	// SameSite sets the SameSite attribute on the cookie
	SameSite string
}

// DefaultCSRFConfig returns the default CSRF configuration
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:     32,
		TokenExpiration: 24 * time.Hour,
		CookieName:      "_csrf",
		HeaderName:      "X-CSRF-Token",
		FormFieldName:   "_csrf",
		Secure:          true,
		SameSite:        "Strict",
	}
}

// CSRFToken represents a CSRF token with metadata
type CSRFToken struct {
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// CSRFManager manages CSRF token generation and validation
type CSRFManager struct {
	config CSRFConfig
	tokens map[string]*CSRFToken
	mu     sync.RWMutex
}

// NewCSRFManager creates a new CSRF manager
func NewCSRFManager(config CSRFConfig) *CSRFManager {
	manager := &CSRFManager{
		config: config,
		tokens: make(map[string]*CSRFToken),
	}

	// Start cleanup goroutine
	go manager.cleanupExpiredTokens()

	return manager
}

// GenerateToken generates a new CSRF token
func (m *CSRFManager) GenerateToken() (*CSRFToken, error) {
	// Generate random bytes
	bytes := make([]byte, m.config.TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	// Encode to base64
	token := base64.URLEncoding.EncodeToString(bytes)

	now := time.Now()
	csrfToken := &CSRFToken{
		Token:     token,
		CreatedAt: now,
		ExpiresAt: now.Add(m.config.TokenExpiration),
	}

	// Store token
	m.mu.Lock()
	m.tokens[token] = csrfToken
	m.mu.Unlock()

	return csrfToken, nil
}

// ValidateToken validates a CSRF token
func (m *CSRFManager) ValidateToken(token string) error {
	if token == "" {
		return ErrMissingCSRFToken
	}

	m.mu.RLock()
	storedToken, exists := m.tokens[token]
	m.mu.RUnlock()

	if !exists {
		return ErrInvalidCSRFToken
	}

	if time.Now().After(storedToken.ExpiresAt) {
		// Remove expired token
		m.mu.Lock()
		delete(m.tokens, token)
		m.mu.Unlock()
		return ErrCSRFTokenExpired
	}

	return nil
}

// ValidateTokenPair validates that the cookie token matches the header/form token
func (m *CSRFManager) ValidateTokenPair(cookieToken, requestToken string) error {
	if cookieToken == "" || requestToken == "" {
		return ErrMissingCSRFToken
	}

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(requestToken)) != 1 {
		return ErrInvalidCSRFToken
	}

	return m.ValidateToken(cookieToken)
}

// RevokeToken revokes a CSRF token
func (m *CSRFManager) RevokeToken(token string) {
	m.mu.Lock()
	delete(m.tokens, token)
	m.mu.Unlock()
}

// cleanupExpiredTokens periodically removes expired tokens
func (m *CSRFManager) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for token, csrfToken := range m.tokens {
			if now.After(csrfToken.ExpiresAt) {
				delete(m.tokens, token)
			}
		}
		m.mu.Unlock()
	}
}

// GetConfig returns the CSRF configuration
func (m *CSRFManager) GetConfig() CSRFConfig {
	return m.config
}

// DefaultCSRFManager is the default CSRF manager instance
var DefaultCSRFManager = NewCSRFManager(DefaultCSRFConfig())

// GenerateCSRFToken generates a new CSRF token using the default manager
func GenerateCSRFToken() (*CSRFToken, error) {
	return DefaultCSRFManager.GenerateToken()
}

// ValidateCSRFToken validates a CSRF token using the default manager
func ValidateCSRFToken(token string) error {
	return DefaultCSRFManager.ValidateToken(token)
}

// ValidateCSRFTokenPair validates a token pair using the default manager
func ValidateCSRFTokenPair(cookieToken, requestToken string) error {
	return DefaultCSRFManager.ValidateTokenPair(cookieToken, requestToken)
}
