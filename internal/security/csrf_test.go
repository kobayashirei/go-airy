package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSRFManager_GenerateToken(t *testing.T) {
	config := DefaultCSRFConfig()
	manager := NewCSRFManager(config)

	token, err := manager.GenerateToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token.Token)
	assert.False(t, token.CreatedAt.IsZero())
	assert.False(t, token.ExpiresAt.IsZero())
	assert.True(t, token.ExpiresAt.After(token.CreatedAt))
}

func TestCSRFManager_ValidateToken(t *testing.T) {
	config := DefaultCSRFConfig()
	manager := NewCSRFManager(config)

	// Generate a token
	token, err := manager.GenerateToken()
	require.NoError(t, err)

	// Validate the token
	err = manager.ValidateToken(token.Token)
	assert.NoError(t, err)

	// Validate an invalid token
	err = manager.ValidateToken("invalid-token")
	assert.ErrorIs(t, err, ErrInvalidCSRFToken)

	// Validate empty token
	err = manager.ValidateToken("")
	assert.ErrorIs(t, err, ErrMissingCSRFToken)
}

func TestCSRFManager_ValidateTokenPair(t *testing.T) {
	config := DefaultCSRFConfig()
	manager := NewCSRFManager(config)

	// Generate a token
	token, err := manager.GenerateToken()
	require.NoError(t, err)

	// Valid pair
	err = manager.ValidateTokenPair(token.Token, token.Token)
	assert.NoError(t, err)

	// Mismatched pair
	err = manager.ValidateTokenPair(token.Token, "different-token")
	assert.ErrorIs(t, err, ErrInvalidCSRFToken)

	// Missing cookie token
	err = manager.ValidateTokenPair("", token.Token)
	assert.ErrorIs(t, err, ErrMissingCSRFToken)

	// Missing request token
	err = manager.ValidateTokenPair(token.Token, "")
	assert.ErrorIs(t, err, ErrMissingCSRFToken)
}

func TestCSRFManager_RevokeToken(t *testing.T) {
	config := DefaultCSRFConfig()
	manager := NewCSRFManager(config)

	// Generate a token
	token, err := manager.GenerateToken()
	require.NoError(t, err)

	// Validate the token
	err = manager.ValidateToken(token.Token)
	assert.NoError(t, err)

	// Revoke the token
	manager.RevokeToken(token.Token)

	// Token should no longer be valid
	err = manager.ValidateToken(token.Token)
	assert.ErrorIs(t, err, ErrInvalidCSRFToken)
}

func TestCSRFManager_TokenExpiration(t *testing.T) {
	config := CSRFConfig{
		TokenLength:     32,
		TokenExpiration: 100 * time.Millisecond, // Very short expiration for testing
		CookieName:      "_csrf",
		HeaderName:      "X-CSRF-Token",
		FormFieldName:   "_csrf",
		Secure:          true,
		SameSite:        "Strict",
	}
	manager := NewCSRFManager(config)

	// Generate a token
	token, err := manager.GenerateToken()
	require.NoError(t, err)

	// Token should be valid immediately
	err = manager.ValidateToken(token.Token)
	assert.NoError(t, err)

	// Wait for token to expire
	time.Sleep(150 * time.Millisecond)

	// Token should be expired
	err = manager.ValidateToken(token.Token)
	assert.ErrorIs(t, err, ErrCSRFTokenExpired)
}

func TestCSRFManager_UniqueTokens(t *testing.T) {
	config := DefaultCSRFConfig()
	manager := NewCSRFManager(config)

	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := manager.GenerateToken()
		require.NoError(t, err)
		assert.False(t, tokens[token.Token], "Token should be unique")
		tokens[token.Token] = true
	}
}

func TestDefaultCSRFConfig(t *testing.T) {
	config := DefaultCSRFConfig()

	assert.Equal(t, 32, config.TokenLength)
	assert.Equal(t, 24*time.Hour, config.TokenExpiration)
	assert.Equal(t, "_csrf", config.CookieName)
	assert.Equal(t, "X-CSRF-Token", config.HeaderName)
	assert.Equal(t, "_csrf", config.FormFieldName)
	assert.True(t, config.Secure)
	assert.Equal(t, "Strict", config.SameSite)
}

func TestDefaultCSRFManager(t *testing.T) {
	// Test that default manager works
	token, err := GenerateCSRFToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token.Token)

	err = ValidateCSRFToken(token.Token)
	assert.NoError(t, err)

	err = ValidateCSRFTokenPair(token.Token, token.Token)
	assert.NoError(t, err)
}
