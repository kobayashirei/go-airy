package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	password := "SecurePassword123!"

	hash, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
	assert.True(t, IsBcryptHash(hash))
}

func TestHashPasswordEmpty(t *testing.T) {
	_, err := HashPassword("")
	assert.Error(t, err)
}

func TestVerifyPassword(t *testing.T) {
	password := "SecurePassword123!"

	hash, err := HashPassword(password)
	require.NoError(t, err)

	// Correct password
	assert.True(t, VerifyPassword(password, hash))

	// Wrong password
	assert.False(t, VerifyPassword("WrongPassword", hash))
}

func TestIsBcryptHash(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		expected bool
	}{
		{
			name:     "valid bcrypt hash",
			hash:     "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
			expected: true,
		},
		{
			name:     "valid bcrypt hash 2b",
			hash:     "$2b$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
			expected: true,
		},
		{
			name:     "too short",
			hash:     "$2a$10$short",
			expected: false,
		},
		{
			name:     "wrong prefix",
			hash:     "$1a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
			expected: false,
		},
		{
			name:     "plain text",
			hash:     "plaintext",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBcryptHash(tt.hash)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEncryptor(t *testing.T) {
	// Generate a key
	key, err := GenerateKey(32)
	require.NoError(t, err)

	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	plaintext := []byte("Hello, World!")

	// Encrypt
	ciphertext, err := encryptor.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	// Decrypt
	decrypted, err := encryptor.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptorString(t *testing.T) {
	key, err := GenerateKey(32)
	require.NoError(t, err)

	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	plaintext := "Sensitive data to encrypt"

	// Encrypt
	ciphertext, err := encryptor.EncryptString(plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	// Decrypt
	decrypted, err := encryptor.DecryptString(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptorInvalidKey(t *testing.T) {
	// Invalid key size
	_, err := NewEncryptor([]byte("short"))
	assert.ErrorIs(t, err, ErrInvalidKey)

	// Valid key sizes
	for _, size := range []int{16, 24, 32} {
		key, err := GenerateKey(size)
		require.NoError(t, err)

		_, err = NewEncryptor(key)
		assert.NoError(t, err)
	}
}

func TestEncryptorFromString(t *testing.T) {
	keyStr, err := GenerateKeyString(32)
	require.NoError(t, err)

	encryptor, err := NewEncryptorFromString(keyStr)
	require.NoError(t, err)

	plaintext := "Test data"
	ciphertext, err := encryptor.EncryptString(plaintext)
	require.NoError(t, err)

	decrypted, err := encryptor.DecryptString(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestGenerateKey(t *testing.T) {
	// Valid sizes
	for _, size := range []int{16, 24, 32} {
		key, err := GenerateKey(size)
		require.NoError(t, err)
		assert.Len(t, key, size)
	}

	// Invalid size
	_, err := GenerateKey(15)
	assert.ErrorIs(t, err, ErrInvalidKey)
}

func TestMaskSensitiveData(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		visibleChars int
		expected     string
	}{
		{
			name:         "normal data",
			data:         "1234567890",
			visibleChars: 2,
			expected:     "12****90",
		},
		{
			name:         "short data",
			data:         "123",
			visibleChars: 2,
			expected:     "****",
		},
		{
			name:         "exact length",
			data:         "1234",
			visibleChars: 2,
			expected:     "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData(tt.data, tt.visibleChars)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "normal email",
			email:    "test@example.com",
			expected: "te****@example.com",
		},
		{
			name:     "short local part",
			email:    "a@example.com",
			expected: "****",
		},
		{
			name:     "no at sign",
			email:    "invalid",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected string
	}{
		{
			name:     "normal phone",
			phone:    "+1234567890",
			expected: "****7890",
		},
		{
			name:     "short phone",
			phone:    "123",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskPhone(tt.phone)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecureCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "equal strings",
			a:        "hello",
			b:        "hello",
			expected: true,
		},
		{
			name:     "different strings",
			a:        "hello",
			b:        "world",
			expected: false,
		},
		{
			name:     "different lengths",
			a:        "hello",
			b:        "hi",
			expected: false,
		},
		{
			name:     "empty strings",
			a:        "",
			b:        "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SecureCompare(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecureWipe(t *testing.T) {
	data := []byte("sensitive data")
	SecureWipe(data)

	for _, b := range data {
		assert.Equal(t, byte(0), b)
	}
}

func TestEncryptionRoundTrip(t *testing.T) {
	key, err := GenerateKey(32)
	require.NoError(t, err)

	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	// Test various data sizes
	testCases := []string{
		"",
		"a",
		"short",
		"This is a longer piece of text that should still encrypt and decrypt correctly.",
		string(make([]byte, 1000)), // 1KB of null bytes
	}

	for _, plaintext := range testCases {
		ciphertext, err := encryptor.EncryptString(plaintext)
		require.NoError(t, err)

		decrypted, err := encryptor.DecryptString(ciphertext)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	key, err := GenerateKey(32)
	require.NoError(t, err)

	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	// Invalid base64
	_, err = encryptor.DecryptString("not-valid-base64!!!")
	assert.Error(t, err)

	// Too short ciphertext
	_, err = encryptor.Decrypt([]byte("short"))
	assert.ErrorIs(t, err, ErrInvalidCiphertext)
}
