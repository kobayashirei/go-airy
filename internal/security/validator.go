package security

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Validation errors
var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidPhone       = errors.New("invalid phone format")
	ErrInvalidUsername    = errors.New("invalid username format")
	ErrInvalidPassword    = errors.New("password does not meet requirements")
	ErrInputTooLong       = errors.New("input exceeds maximum length")
	ErrInputTooShort      = errors.New("input is too short")
	ErrInvalidCharacters  = errors.New("input contains invalid characters")
	ErrPotentialInjection = errors.New("input contains potentially dangerous content")
)

// Validation patterns
var (
	// Email regex pattern (RFC 5322 simplified)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Phone regex pattern (E.164 format)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

	// Username regex pattern (alphanumeric, underscore, hyphen)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

	// SQL injection patterns
	sqlInjectionPatterns = []string{
		"--",
		";--",
		"/*",
		"*/",
		"@@",
		"char(",
		"nchar(",
		"varchar(",
		"nvarchar(",
		"alter ",
		"begin ",
		"cast(",
		"create ",
		"cursor ",
		"declare ",
		"delete ",
		"drop ",
		"end ",
		"exec(",
		"execute(",
		"fetch ",
		"insert ",
		"kill ",
		"select ",
		"sys.",
		"sysobjects",
		"syscolumns",
		"table ",
		"update ",
		"union ",
		"xp_",
	}

	// XSS patterns
	xssPatterns = []string{
		"<script",
		"</script",
		"javascript:",
		"vbscript:",
		"onload=",
		"onerror=",
		"onclick=",
		"onmouseover=",
		"onfocus=",
		"onblur=",
		"onsubmit=",
		"onreset=",
		"onselect=",
		"onchange=",
		"ondblclick=",
		"onkeydown=",
		"onkeypress=",
		"onkeyup=",
		"onmousedown=",
		"onmouseup=",
		"onmousemove=",
		"onmouseout=",
		"expression(",
		"eval(",
		"document.cookie",
		"document.write",
		"window.location",
	}
)

// Validator provides input validation functionality
type Validator struct {
	sanitizer *Sanitizer
}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{
		sanitizer: NewSanitizer(),
	}
}

// ValidateEmail validates an email address
func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return ErrInputTooShort
	}
	if len(email) > 254 {
		return ErrInputTooLong
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidatePhone validates a phone number
func (v *Validator) ValidatePhone(phone string) error {
	if phone == "" {
		return ErrInputTooShort
	}
	if len(phone) > 20 {
		return ErrInputTooLong
	}
	if !phoneRegex.MatchString(phone) {
		return ErrInvalidPhone
	}
	return nil
}

// ValidateUsername validates a username
func (v *Validator) ValidateUsername(username string) error {
	if len(username) < 3 {
		return ErrInputTooShort
	}
	if len(username) > 50 {
		return ErrInputTooLong
	}
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}
	return nil
}

// ValidatePassword validates a password
// Requirements:
// - Minimum 8 characters
// - Maximum 128 characters
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one digit
func (v *Validator) ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrInputTooShort
	}
	if len(password) > 128 {
		return ErrInputTooLong
	}

	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return ErrInvalidPassword
	}

	return nil
}

// ValidateLength validates that input length is within bounds
func (v *Validator) ValidateLength(input string, minLen, maxLen int) error {
	length := utf8.RuneCountInString(input)
	if length < minLen {
		return ErrInputTooShort
	}
	if length > maxLen {
		return ErrInputTooLong
	}
	return nil
}

// CheckSQLInjection checks for potential SQL injection patterns
func (v *Validator) CheckSQLInjection(input string) error {
	lowerInput := strings.ToLower(input)
	for _, pattern := range sqlInjectionPatterns {
		if strings.Contains(lowerInput, pattern) {
			return ErrPotentialInjection
		}
	}
	return nil
}

// CheckXSS checks for potential XSS patterns
func (v *Validator) CheckXSS(input string) error {
	lowerInput := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerInput, pattern) {
			return ErrPotentialInjection
		}
	}
	return nil
}

// ValidateAndSanitize validates and sanitizes input
// Returns the sanitized input and any validation error
func (v *Validator) ValidateAndSanitize(input string, fieldType FieldType, minLen, maxLen int) (string, error) {
	// Check for potential injection attacks
	if err := v.CheckSQLInjection(input); err != nil {
		return "", err
	}
	if err := v.CheckXSS(input); err != nil {
		return "", err
	}

	// Sanitize based on field type
	var sanitized string
	switch fieldType {
	case FieldTypeUsername:
		sanitized = v.sanitizer.SanitizeUsername(input)
		if err := v.ValidateUsername(sanitized); err != nil {
			return "", err
		}
	case FieldTypeEmail:
		sanitized = v.sanitizer.SanitizeEmail(input)
		if err := v.ValidateEmail(sanitized); err != nil {
			return "", err
		}
	case FieldTypePhone:
		sanitized = v.sanitizer.SanitizePhone(input)
		if err := v.ValidatePhone(sanitized); err != nil {
			return "", err
		}
	case FieldTypeTitle:
		sanitized = v.sanitizer.SanitizeTitle(input)
	case FieldTypeContent:
		sanitized = v.sanitizer.SanitizeUGC(input)
	case FieldTypeSearch:
		sanitized = v.sanitizer.SanitizeSearchQuery(input)
	default:
		sanitized = v.sanitizer.SanitizeStrict(input)
	}

	// Validate length
	if err := v.ValidateLength(sanitized, minLen, maxLen); err != nil {
		return "", err
	}

	return sanitized, nil
}

// FieldType represents the type of field being validated
type FieldType int

const (
	FieldTypeGeneric FieldType = iota
	FieldTypeUsername
	FieldTypeEmail
	FieldTypePhone
	FieldTypeTitle
	FieldTypeContent
	FieldTypeSearch
)

// DefaultValidator is the default validator instance
var DefaultValidator = NewValidator()

// ValidateEmail is a convenience function using the default validator
func ValidateEmail(email string) error {
	return DefaultValidator.ValidateEmail(email)
}

// ValidatePhone is a convenience function using the default validator
func ValidatePhone(phone string) error {
	return DefaultValidator.ValidatePhone(phone)
}

// ValidateUsername is a convenience function using the default validator
func ValidateUsername(username string) error {
	return DefaultValidator.ValidateUsername(username)
}

// ValidatePassword is a convenience function using the default validator
func ValidatePassword(password string) error {
	return DefaultValidator.ValidatePassword(password)
}

// CheckSQLInjection is a convenience function using the default validator
func CheckSQLInjection(input string) error {
	return DefaultValidator.CheckSQLInjection(input)
}

// CheckXSS is a convenience function using the default validator
func CheckXSS(input string) error {
	return DefaultValidator.CheckXSS(input)
}

// ValidateAndSanitize is a convenience function using the default validator
func ValidateAndSanitize(input string, fieldType FieldType, minLen, maxLen int) (string, error) {
	return DefaultValidator.ValidateAndSanitize(input, fieldType, minLen, maxLen)
}
