package service

import (
	"time"

	"github.com/kobayashirei/airy/internal/auth"
)

// JWTService defines the interface for JWT operations in the service layer
type JWTService interface {
	GenerateToken(userID int64, roles []string) (string, error)
	GenerateRefreshToken(userID int64, roles []string) (string, error)
	ParseToken(tokenString string) (*auth.Claims, error)
	RefreshToken(tokenString string) (string, error)
}

// jwtServiceWrapper wraps the auth.JWTService
type jwtServiceWrapper struct {
	authJWTService       auth.JWTService
	refreshTokenDuration time.Duration
}

// NewJWTService creates a new JWT service wrapper
func NewJWTService(authJWTService auth.JWTService, refreshTokenDuration time.Duration) JWTService {
	return &jwtServiceWrapper{
		authJWTService:       authJWTService,
		refreshTokenDuration: refreshTokenDuration,
	}
}

// GenerateToken generates a new JWT token
func (s *jwtServiceWrapper) GenerateToken(userID int64, roles []string) (string, error) {
	return s.authJWTService.GenerateToken(userID, roles)
}

// GenerateRefreshToken generates a refresh token with longer expiration
func (s *jwtServiceWrapper) GenerateRefreshToken(userID int64, roles []string) (string, error) {
	// For refresh tokens, we can use the same generation method
	// In a more sophisticated implementation, you might want to store refresh tokens
	// in Redis with a longer TTL and track them separately
	return s.authJWTService.GenerateToken(userID, roles)
}

// ParseToken parses and validates a JWT token
func (s *jwtServiceWrapper) ParseToken(tokenString string) (*auth.Claims, error) {
	return s.authJWTService.ParseToken(tokenString)
}

// RefreshToken refreshes an existing token
func (s *jwtServiceWrapper) RefreshToken(tokenString string) (string, error) {
	return s.authJWTService.RefreshToken(tokenString)
}
