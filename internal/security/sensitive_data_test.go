package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSensitiveDataService(t *testing.T) {
	// Generate a valid key
	keyStr, err := GenerateKeyString(32)
	require.NoError(t, err)

	t.Run("with valid encryption key", func(t *testing.T) {
		svc, err := NewSensitiveDataService(SensitiveDataConfig{
			EncryptionKey: keyStr,
			BcryptCost:    10,
		})
		require.NoError(t, err)
		assert.NotNil(t, svc)
		assert.True(t, svc.IsEncryptionEnabled())
	})

	t.Run("without encryption key", func(t *testing.T) {
		svc, err := NewSensitiveDataService(SensitiveDataConfig{
			EncryptionKey: "",
			BcryptCost:    10,
		})
		require.NoError(t, err)
		assert.NotNil(t, svc)
		assert.False(t, svc.IsEncryptionEnabled())
	})

	t.Run("with invalid encryption key", func(t *testing.T) {
		_, err := NewSensitiveDataService(SensitiveDataConfig{
			EncryptionKey: "invalid-key",
			BcryptCost:    10,
		})
		assert.Error(t, err)
	})
}

func TestSensitiveDataService_EncryptDecryptField(t *testing.T) {
	keyStr, err := GenerateKeyString(32)
	require.NoError(t, err)

	svc, err := NewSensitiveDataService(SensitiveDataConfig{
		EncryptionKey: keyStr,
		BcryptCost:    10,
	})
	require.NoError(t, err)

	t.Run("encrypt and decrypt field", func(t *testing.T) {
		original := "sensitive-data-123"
		encrypted, err := svc.EncryptField(original)
		require.NoError(t, err)
		assert.NotEqual(t, original, encrypted)

		decrypted, err := svc.DecryptField(encrypted)
		require.NoError(t, err)
		assert.Equal(t, original, decrypted)
	})

	t.Run("empty field", func(t *testing.T) {
		encrypted, err := svc.EncryptField("")
		require.NoError(t, err)
		assert.Equal(t, "", encrypted)

		decrypted, err := svc.DecryptField("")
		require.NoError(t, err)
		assert.Equal(t, "", decrypted)
	})
}

func TestSensitiveDataService_EncryptDecryptEmail(t *testing.T) {
	keyStr, err := GenerateKeyString(32)
	require.NoError(t, err)

	svc, err := NewSensitiveDataService(SensitiveDataConfig{
		EncryptionKey: keyStr,
		BcryptCost:    10,
	})
	require.NoError(t, err)

	email := "test@example.com"
	encrypted, err := svc.EncryptEmail(email)
	require.NoError(t, err)
	assert.NotEqual(t, email, encrypted)

	decrypted, err := svc.DecryptEmail(encrypted)
	require.NoError(t, err)
	assert.Equal(t, email, decrypted)
}

func TestSensitiveDataService_EncryptDecryptPhone(t *testing.T) {
	keyStr, err := GenerateKeyString(32)
	require.NoError(t, err)

	svc, err := NewSensitiveDataService(SensitiveDataConfig{
		EncryptionKey: keyStr,
		BcryptCost:    10,
	})
	require.NoError(t, err)

	phone := "+1234567890"
	encrypted, err := svc.EncryptPhone(phone)
	require.NoError(t, err)
	assert.NotEqual(t, phone, encrypted)

	decrypted, err := svc.DecryptPhone(encrypted)
	require.NoError(t, err)
	assert.Equal(t, phone, decrypted)
}

func TestSensitiveDataService_EncryptDecryptFields(t *testing.T) {
	keyStr, err := GenerateKeyString(32)
	require.NoError(t, err)

	svc, err := NewSensitiveDataService(SensitiveDataConfig{
		EncryptionKey: keyStr,
		BcryptCost:    10,
	})
	require.NoError(t, err)

	t.Run("encrypt and decrypt all fields", func(t *testing.T) {
		fields := &SensitiveFields{
			Email:   "test@example.com",
			Phone:   "+1234567890",
			Address: "123 Main St",
		}

		originalEmail := fields.Email
		originalPhone := fields.Phone
		originalAddress := fields.Address

		err := svc.EncryptFields(fields)
		require.NoError(t, err)

		assert.NotEqual(t, originalEmail, fields.Email)
		assert.NotEqual(t, originalPhone, fields.Phone)
		assert.NotEqual(t, originalAddress, fields.Address)

		err = svc.DecryptFields(fields)
		require.NoError(t, err)

		assert.Equal(t, originalEmail, fields.Email)
		assert.Equal(t, originalPhone, fields.Phone)
		assert.Equal(t, originalAddress, fields.Address)
	})

	t.Run("nil fields", func(t *testing.T) {
		err := svc.EncryptFields(nil)
		assert.NoError(t, err)

		err = svc.DecryptFields(nil)
		assert.NoError(t, err)
	})

	t.Run("partial fields", func(t *testing.T) {
		fields := &SensitiveFields{
			Email: "test@example.com",
		}

		err := svc.EncryptFields(fields)
		require.NoError(t, err)

		err = svc.DecryptFields(fields)
		require.NoError(t, err)

		assert.Equal(t, "test@example.com", fields.Email)
		assert.Equal(t, "", fields.Phone)
		assert.Equal(t, "", fields.Address)
	})
}

func TestSensitiveDataService_DisabledEncryption(t *testing.T) {
	svc, err := NewSensitiveDataService(SensitiveDataConfig{
		EncryptionKey: "",
		BcryptCost:    10,
	})
	require.NoError(t, err)
	assert.False(t, svc.IsEncryptionEnabled())

	t.Run("encrypt returns original value", func(t *testing.T) {
		original := "sensitive-data"
		encrypted, err := svc.EncryptField(original)
		require.NoError(t, err)
		assert.Equal(t, original, encrypted)
	})

	t.Run("decrypt returns original value", func(t *testing.T) {
		original := "sensitive-data"
		decrypted, err := svc.DecryptField(original)
		require.NoError(t, err)
		assert.Equal(t, original, decrypted)
	})
}

func TestHashPasswordWithCost(t *testing.T) {
	t.Run("valid password and cost", func(t *testing.T) {
		hash, err := HashPasswordWithCost("password123", 10)
		require.NoError(t, err)
		assert.True(t, IsBcryptHash(hash))
		assert.True(t, VerifyPassword("password123", hash))
	})

	t.Run("empty password", func(t *testing.T) {
		_, err := HashPasswordWithCost("", 10)
		assert.Error(t, err)
	})

	t.Run("invalid cost uses default", func(t *testing.T) {
		hash, err := HashPasswordWithCost("password123", 0)
		require.NoError(t, err)
		assert.True(t, IsBcryptHash(hash))
	})

	t.Run("cost too high uses default", func(t *testing.T) {
		hash, err := HashPasswordWithCost("password123", 100)
		require.NoError(t, err)
		assert.True(t, IsBcryptHash(hash))
	})
}
