package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/kobayashirei/airy/internal/cache"
	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/repository"
)

var (
	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrInvalidPhone is returned when phone format is invalid
	ErrInvalidPhone = errors.New("invalid phone format")
	// ErrUserExists is returned when user already exists
	ErrUserExists = errors.New("user already exists")
	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidToken is returned when activation token is invalid
	ErrInvalidToken = errors.New("invalid or expired token")
	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidVerificationCode is returned when verification code is invalid
	ErrInvalidVerificationCode = errors.New("invalid or expired verification code")
)

// Email regex pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Phone regex pattern (supports international formats)
var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

// UserService defines the interface for user business logic
type UserService interface {
    Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error)
    Activate(ctx context.Context, token string) error
    Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
    LoginWithCode(ctx context.Context, req LoginWithCodeRequest) (*LoginResponse, error)
    RefreshToken(ctx context.Context, token string) (*RefreshTokenResponse, error)
    ResendActivation(ctx context.Context, identifier string) error
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required_without=Phone"`
	Phone    string `json:"phone" binding:"required_without=Email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// RegisterResponse represents a user registration response
type RegisterResponse struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

// LoginRequest represents a login request with password
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // email, phone, or username
	Password   string `json:"password" binding:"required"`
	ClientIP   string `json:"-"` // Client IP for logging (not from JSON)
}

// LoginWithCodeRequest represents a login request with verification code
type LoginWithCodeRequest struct {
	Identifier string `json:"identifier" binding:"required"` // email or phone
	Code       string `json:"code" binding:"required"`
	ClientIP   string `json:"-"` // Client IP for logging (not from JSON)
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	User         *models.User `json:"user"`
}

// RefreshTokenResponse represents a token refresh response
type RefreshTokenResponse struct {
	Token string `json:"token"`
}

// userService implements UserService interface
type userService struct {
	userRepo     repository.UserRepository
	cacheService cache.Service
	emailService EmailService
	jwtService   JWTService
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repository.UserRepository,
	cacheService cache.Service,
	emailService EmailService,
	jwtService JWTService,
) UserService {
	return &userService{
		userRepo:     userRepo,
		cacheService: cacheService,
		emailService: emailService,
		jwtService:   jwtService,
	}
}

// Register registers a new user
func (s *userService) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	// Validate email format if provided
	if req.Email != "" && !emailRegex.MatchString(req.Email) {
		return nil, ErrInvalidEmail
	}

	// Validate phone format if provided
	if req.Phone != "" && !phoneRegex.MatchString(req.Phone) {
		return nil, ErrInvalidPhone
	}

	// Check for duplicate email
	if req.Email != "" {
		existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if existingUser != nil {
			return nil, ErrUserExists
		}
	}

	// Check for duplicate phone
	if req.Phone != "" {
		existingUser, err := s.userRepo.FindByPhone(ctx, req.Phone)
		if err != nil {
			return nil, fmt.Errorf("failed to check phone: %w", err)
		}
		if existingUser != nil {
			return nil, ErrUserExists
		}
	}

	// Check for duplicate username
	existingUser, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUserExists
	}

	// Hash password using bcrypt
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(passwordHash),
		Status:       "inactive",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate activation token
	token, err := generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate activation token: %w", err)
	}

	// Store activation token in Redis with 24 hour expiration
	tokenKey := cache.ActivationTokenKey(token)
	if err := s.cacheService.Set(ctx, tokenKey, user.ID, 24*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to store activation token: %w", err)
	}

	// Send activation email
	if req.Email != "" {
		if err := s.emailService.SendActivationEmail(ctx, req.Email, token); err != nil {
			// Log error but don't fail registration
			// In production, you might want to queue this for retry
			fmt.Printf("failed to send activation email: %v\n", err)
		}
	}

	return &RegisterResponse{
		UserID:  user.ID,
		Message: "Registration successful. Please check your email to activate your account.",
	}, nil
}

// Activate activates a user account using the activation token
func (s *userService) Activate(ctx context.Context, token string) error {
	// Retrieve user ID from Redis
	tokenKey := cache.ActivationTokenKey(token)
	var userID int64
	if err := s.cacheService.Get(ctx, tokenKey, &userID); err != nil {
		if errors.Is(err, cache.ErrCacheMiss) {
			return ErrInvalidToken
		}
		return fmt.Errorf("failed to retrieve activation token: %w", err)
	}

	// Find user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Update user status to active
	user.Status = "active"
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	// Delete activation token from Redis
	if err := s.cacheService.Delete(ctx, tokenKey); err != nil {
		// Log error but don't fail activation
		fmt.Printf("failed to delete activation token: %v\n", err)
	}

	return nil
}

// Login authenticates a user with password
func (s *userService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Find user by identifier (email, phone, or username)
	var user *models.User
	var err error

	// Try email first
	if emailRegex.MatchString(req.Identifier) {
		user, err = s.userRepo.FindByEmail(ctx, req.Identifier)
	} else if phoneRegex.MatchString(req.Identifier) {
		// Try phone
		user, err = s.userRepo.FindByPhone(ctx, req.Identifier)
	} else {
		// Try username
		user, err = s.userRepo.FindByUsername(ctx, req.Identifier)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check user status
	if user.Status != "active" {
		return nil, errors.New("user account is not active")
	}

	// Generate JWT token
	roles := []string{"user"} // Default role, can be extended to fetch from user_roles table
	token, err := s.jwtService.GenerateToken(user.ID, roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate refresh token (longer expiration)
	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update login info with client IP
	if err := s.userRepo.UpdateLoginInfo(ctx, user.ID, req.ClientIP); err != nil {
		// Log error but don't fail login
		fmt.Printf("failed to update login info: %v\n", err)
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// LoginWithCode authenticates a user with verification code
func (s *userService) LoginWithCode(ctx context.Context, req LoginWithCodeRequest) (*LoginResponse, error) {
	// Verify code from Redis
	codeKey := cache.VerificationCodeKey(req.Identifier)
	var storedCode string
	if err := s.cacheService.Get(ctx, codeKey, &storedCode); err != nil {
		if errors.Is(err, cache.ErrCacheMiss) {
			return nil, ErrInvalidVerificationCode
		}
		return nil, fmt.Errorf("failed to retrieve verification code: %w", err)
	}

	if storedCode != req.Code {
		return nil, ErrInvalidVerificationCode
	}

	// Find user by identifier
	var user *models.User
	var err error

	if emailRegex.MatchString(req.Identifier) {
		user, err = s.userRepo.FindByEmail(ctx, req.Identifier)
	} else if phoneRegex.MatchString(req.Identifier) {
		user, err = s.userRepo.FindByPhone(ctx, req.Identifier)
	} else {
		return nil, errors.New("invalid identifier format")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check user status
	if user.Status != "active" {
		return nil, errors.New("user account is not active")
	}

	// Generate JWT token
	roles := []string{"user"}
	token, err := s.jwtService.GenerateToken(user.ID, roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Delete verification code from Redis
	if err := s.cacheService.Delete(ctx, codeKey); err != nil {
		fmt.Printf("failed to delete verification code: %v\n", err)
	}

	// Update login info with client IP
	if err := s.userRepo.UpdateLoginInfo(ctx, user.ID, req.ClientIP); err != nil {
		fmt.Printf("failed to update login info: %v\n", err)
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// RefreshToken refreshes an access token
func (s *userService) RefreshToken(ctx context.Context, token string) (*RefreshTokenResponse, error) {
	newToken, err := s.jwtService.RefreshToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &RefreshTokenResponse{
		Token: newToken,
	}, nil
}

// ResendActivation regenerates and resends activation token to user's email
func (s *userService) ResendActivation(ctx context.Context, identifier string) error {
    var user *models.User
    var err error
    if emailRegex.MatchString(identifier) {
        user, err = s.userRepo.FindByEmail(ctx, identifier)
    } else if phoneRegex.MatchString(identifier) {
        // If identifier is phone, try phone then fallback by username
        user, err = s.userRepo.FindByPhone(ctx, identifier)
    } else {
        user, err = s.userRepo.FindByUsername(ctx, identifier)
    }
    if err != nil {
        return fmt.Errorf("failed to find user: %w", err)
    }
    if user == nil {
        return ErrUserNotFound
    }
    if user.Status == "active" {
        return errors.New("user already active")
    }
    if user.Email == "" {
        return errors.New("user has no email")
    }

    token, err := generateToken(32)
    if err != nil {
        return fmt.Errorf("failed to generate activation token: %w", err)
    }
    tokenKey := cache.ActivationTokenKey(token)
    if err := s.cacheService.Set(ctx, tokenKey, user.ID, 24*time.Hour); err != nil {
        return fmt.Errorf("failed to store activation token: %w", err)
    }
    if err := s.emailService.SendActivationEmail(ctx, user.Email, token); err != nil {
        return fmt.Errorf("failed to send activation email: %w", err)
    }
    return nil
}

// generateToken generates a random token
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
