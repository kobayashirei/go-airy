package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{
			name:    "valid email",
			email:   "test@example.com",
			wantErr: nil,
		},
		{
			name:    "valid email with subdomain",
			email:   "test@mail.example.com",
			wantErr: nil,
		},
		{
			name:    "valid email with plus",
			email:   "test+tag@example.com",
			wantErr: nil,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: ErrInputTooShort,
		},
		{
			name:    "invalid email no @",
			email:   "testexample.com",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "invalid email no domain",
			email:   "test@",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "invalid email no TLD",
			email:   "test@example",
			wantErr: ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateEmail(tt.email)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePhone(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		phone   string
		wantErr error
	}{
		{
			name:    "valid phone with plus",
			phone:   "+1234567890",
			wantErr: nil,
		},
		{
			name:    "valid phone without plus",
			phone:   "1234567890",
			wantErr: nil,
		},
		{
			name:    "empty phone",
			phone:   "",
			wantErr: ErrInputTooShort,
		},
		{
			name:    "invalid phone with letters",
			phone:   "+123abc",
			wantErr: ErrInvalidPhone,
		},
		{
			name:    "invalid phone starting with 0",
			phone:   "0123456789",
			wantErr: ErrInvalidPhone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePhone(tt.phone)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{
			name:     "valid username",
			username: "john_doe",
			wantErr:  nil,
		},
		{
			name:     "valid username with hyphen",
			username: "john-doe",
			wantErr:  nil,
		},
		{
			name:     "valid username alphanumeric",
			username: "john123",
			wantErr:  nil,
		},
		{
			name:     "too short",
			username: "ab",
			wantErr:  ErrInputTooShort,
		},
		{
			name:     "invalid characters",
			username: "john@doe",
			wantErr:  ErrInvalidUsername,
		},
		{
			name:     "invalid with spaces",
			username: "john doe",
			wantErr:  ErrInvalidUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUsername(tt.username)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid password",
			password: "Password123",
			wantErr:  nil,
		},
		{
			name:     "valid password with special chars",
			password: "Password123!@#",
			wantErr:  nil,
		},
		{
			name:     "too short",
			password: "Pass1",
			wantErr:  ErrInputTooShort,
		},
		{
			name:     "no uppercase",
			password: "password123",
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "no lowercase",
			password: "PASSWORD123",
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "no digit",
			password: "PasswordABC",
			wantErr:  ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePassword(tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckSQLInjection(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "normal input",
			input:   "Hello World",
			wantErr: false,
		},
		{
			name:    "SQL comment",
			input:   "test--comment",
			wantErr: true,
		},
		{
			name:    "SQL union",
			input:   "test UNION SELECT * FROM users",
			wantErr: true,
		},
		{
			name:    "SQL drop",
			input:   "test; DROP TABLE users;",
			wantErr: true,
		},
		{
			name:    "SQL select",
			input:   "SELECT * FROM users",
			wantErr: true,
		},
		{
			name:    "SQL insert",
			input:   "INSERT INTO users VALUES",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.CheckSQLInjection(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckXSS(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "normal input",
			input:   "Hello World",
			wantErr: false,
		},
		{
			name:    "script tag",
			input:   "<script>alert('xss')</script>",
			wantErr: true,
		},
		{
			name:    "javascript protocol",
			input:   "javascript:alert('xss')",
			wantErr: true,
		},
		{
			name:    "onclick event",
			input:   `<img onclick="alert('xss')">`,
			wantErr: true,
		},
		{
			name:    "onerror event",
			input:   `<img onerror="alert('xss')">`,
			wantErr: true,
		},
		{
			name:    "document.cookie",
			input:   "document.cookie",
			wantErr: true,
		},
		{
			name:    "eval function",
			input:   "eval(code)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.CheckXSS(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateLength(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   string
		minLen  int
		maxLen  int
		wantErr error
	}{
		{
			name:    "valid length",
			input:   "Hello",
			minLen:  1,
			maxLen:  10,
			wantErr: nil,
		},
		{
			name:    "too short",
			input:   "Hi",
			minLen:  5,
			maxLen:  10,
			wantErr: ErrInputTooShort,
		},
		{
			name:    "too long",
			input:   "Hello World!",
			minLen:  1,
			maxLen:  5,
			wantErr: ErrInputTooLong,
		},
		{
			name:    "unicode characters",
			input:   "你好世界",
			minLen:  1,
			maxLen:  10,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateLength(tt.input, tt.minLen, tt.maxLen)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
