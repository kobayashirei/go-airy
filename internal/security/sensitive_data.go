package security

import (
	"errors"
	"sync"
)

// SensitiveDataService provides encryption/decryption for sensitive data fields
type SensitiveDataService struct {
	encryptor *Encryptor
	mu        sync.RWMutex
}

var (
	// ErrEncryptorNotInitialized is returned when the encryptor is not initialized
	ErrEncryptorNotInitialized = errors.New("encryptor not initialized")
	// ErrEmptyData is returned when trying to encrypt empty data
	ErrEmptyData = errors.New("data cannot be empty")
)

// SensitiveDataConfig holds configuration for sensitive data encryption
type SensitiveDataConfig struct {
	// EncryptionKey is the base64-encoded AES encryption key
	EncryptionKey string
	// BcryptCost is the cost factor for bcrypt password hashing
	BcryptCost int
}

// NewSensitiveDataService creates a new sensitive data service
// If encryptionKey is empty, encryption will be disabled (not recommended for production)
func NewSensitiveDataService(config SensitiveDataConfig) (*SensitiveDataService, error) {
	svc := &SensitiveDataService{}

	if config.EncryptionKey != "" {
		encryptor, err := NewEncryptorFromString(config.EncryptionKey)
		if err != nil {
			return nil, err
		}
		svc.encryptor = encryptor
	}

	return svc, nil
}

// IsEncryptionEnabled returns true if encryption is enabled
func (s *SensitiveDataService) IsEncryptionEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.encryptor != nil
}

// EncryptField encrypts a sensitive field value
// Returns the original value if encryption is disabled
func (s *SensitiveDataService) EncryptField(value string) (string, error) {
	if value == "" {
		return "", nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.encryptor == nil {
		// Encryption disabled, return original value
		return value, nil
	}

	return s.encryptor.EncryptString(value)
}

// DecryptField decrypts a sensitive field value
// Returns the original value if encryption is disabled
func (s *SensitiveDataService) DecryptField(value string) (string, error) {
	if value == "" {
		return "", nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.encryptor == nil {
		// Encryption disabled, return original value
		return value, nil
	}

	return s.encryptor.DecryptString(value)
}


// EncryptEmail encrypts an email address for storage
func (s *SensitiveDataService) EncryptEmail(email string) (string, error) {
	return s.EncryptField(email)
}

// DecryptEmail decrypts an email address
func (s *SensitiveDataService) DecryptEmail(email string) (string, error) {
	return s.DecryptField(email)
}

// EncryptPhone encrypts a phone number for storage
func (s *SensitiveDataService) EncryptPhone(phone string) (string, error) {
	return s.EncryptField(phone)
}

// DecryptPhone decrypts a phone number
func (s *SensitiveDataService) DecryptPhone(phone string) (string, error) {
	return s.DecryptField(phone)
}

// EncryptAddress encrypts an address for storage
func (s *SensitiveDataService) EncryptAddress(address string) (string, error) {
	return s.EncryptField(address)
}

// DecryptAddress decrypts an address
func (s *SensitiveDataService) DecryptAddress(address string) (string, error) {
	return s.DecryptField(address)
}

// SensitiveFields represents a collection of sensitive fields that can be encrypted/decrypted
type SensitiveFields struct {
	Email   string
	Phone   string
	Address string
}

// EncryptFields encrypts all sensitive fields in the struct
func (s *SensitiveDataService) EncryptFields(fields *SensitiveFields) error {
	if fields == nil {
		return nil
	}

	var err error

	if fields.Email != "" {
		fields.Email, err = s.EncryptEmail(fields.Email)
		if err != nil {
			return err
		}
	}

	if fields.Phone != "" {
		fields.Phone, err = s.EncryptPhone(fields.Phone)
		if err != nil {
			return err
		}
	}

	if fields.Address != "" {
		fields.Address, err = s.EncryptAddress(fields.Address)
		if err != nil {
			return err
		}
	}

	return nil
}

// DecryptFields decrypts all sensitive fields in the struct
func (s *SensitiveDataService) DecryptFields(fields *SensitiveFields) error {
	if fields == nil {
		return nil
	}

	var err error

	if fields.Email != "" {
		fields.Email, err = s.DecryptEmail(fields.Email)
		if err != nil {
			return err
		}
	}

	if fields.Phone != "" {
		fields.Phone, err = s.DecryptPhone(fields.Phone)
		if err != nil {
			return err
		}
	}

	if fields.Address != "" {
		fields.Address, err = s.DecryptAddress(fields.Address)
		if err != nil {
			return err
		}
	}

	return nil
}

// HashPasswordWithCost hashes a password using bcrypt with the specified cost
func HashPasswordWithCost(password string, cost int) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	if cost < 4 || cost > 31 {
		cost = 10 // Default cost
	}

	config := PasswordConfig{Cost: cost}
	return HashPasswordWithConfig(password, config)
}
