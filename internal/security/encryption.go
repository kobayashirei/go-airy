package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/bcrypt"
)

// Encryption errors
var (
	ErrInvalidKey        = errors.New("invalid encryption key")
	ErrEncryptionFailed  = errors.New("encryption failed")
	ErrDecryptionFailed  = errors.New("decryption failed")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

// PasswordConfig holds configuration for password hashing
type PasswordConfig struct {
	// Cost is the bcrypt cost factor (4-31, default 10)
	Cost int
}

// DefaultPasswordConfig returns the default password configuration
func DefaultPasswordConfig() PasswordConfig {
	return PasswordConfig{
		Cost: bcrypt.DefaultCost, // 10
	}
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	return HashPasswordWithConfig(password, DefaultPasswordConfig())
}

// HashPasswordWithConfig hashes a password with custom configuration
func HashPasswordWithConfig(password string, config PasswordConfig) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), config.Cost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// IsBcryptHash checks if a string is a valid bcrypt hash
func IsBcryptHash(hash string) bool {
	// Bcrypt hashes start with $2a$, $2b$, or $2y$ and are 60 characters long
	if len(hash) != 60 {
		return false
	}
	if hash[0] != '$' || hash[1] != '2' {
		return false
	}
	if hash[2] != 'a' && hash[2] != 'b' && hash[2] != 'y' {
		return false
	}
	return hash[3] == '$'
}

// Encryptor provides AES-GCM encryption for sensitive data
type Encryptor struct {
	key []byte
}

// NewEncryptor creates a new encryptor with the given key
// Key must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256
func NewEncryptor(key []byte) (*Encryptor, error) {
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, ErrInvalidKey
	}

	return &Encryptor{key: key}, nil
}

// NewEncryptorFromString creates a new encryptor from a base64-encoded key
func NewEncryptorFromString(keyStr string) (*Encryptor, error) {
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, ErrInvalidKey
	}
	return NewEncryptor(key)
}

// Encrypt encrypts plaintext using AES-GCM
func (e *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, ErrEncryptionFailed
	}

	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-GCM
func (e *Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrInvalidCiphertext
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns base64-encoded ciphertext
func (e *Encryptor) EncryptString(plaintext string) (string, error) {
	ciphertext, err := e.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptString decrypts base64-encoded ciphertext and returns the plaintext string
func (e *Encryptor) DecryptString(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	plaintext, err := e.Decrypt(data)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GenerateKey generates a random encryption key of the specified size
// Size must be 16, 24, or 32 bytes
func GenerateKey(size int) ([]byte, error) {
	if size != 16 && size != 24 && size != 32 {
		return nil, ErrInvalidKey
	}

	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	return key, nil
}

// GenerateKeyString generates a random encryption key and returns it as base64
func GenerateKeyString(size int) (string, error) {
	key, err := GenerateKey(size)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// MaskSensitiveData masks sensitive data for logging/display
// Shows only the first and last few characters
func MaskSensitiveData(data string, visibleChars int) string {
	if len(data) <= visibleChars*2 {
		return "****"
	}

	prefix := data[:visibleChars]
	suffix := data[len(data)-visibleChars:]
	masked := prefix + "****" + suffix

	return masked
}

// MaskEmail masks an email address for display
func MaskEmail(email string) string {
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}

	if atIndex == -1 || atIndex < 2 {
		return "****"
	}

	// Show first 2 chars and domain
	return email[:2] + "****" + email[atIndex:]
}

// MaskPhone masks a phone number for display
func MaskPhone(phone string) string {
	if len(phone) < 4 {
		return "****"
	}

	// Show last 4 digits
	return "****" + phone[len(phone)-4:]
}

// SecureCompare performs constant-time comparison of two strings
// This prevents timing attacks
func SecureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// SecureWipe overwrites a byte slice with zeros
// Use this to clear sensitive data from memory
func SecureWipe(data []byte) {
	for i := range data {
		data[i] = 0
	}
}
