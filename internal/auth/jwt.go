package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when the token has expired
	ErrExpiredToken = errors.New("token has expired")
	// ErrTokenNotFound is returned when no token is provided
	ErrTokenNotFound = errors.New("token not found")
)

// Claims represents the JWT claims
type Claims struct {
	UserID int64    `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService interface {
	GenerateToken(userID int64, roles []string) (string, error)
	ParseToken(tokenString string) (*Claims, error)
	RefreshToken(tokenString string) (string, error)
}

// jwtService implements JWTService interface
type jwtService struct {
	secretKey  []byte
	expiration time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, expiration time.Duration) JWTService {
	return &jwtService{
		secretKey:  []byte(secretKey),
		expiration: expiration,
	}
}

// GenerateToken generates a new JWT token
func (s *jwtService) GenerateToken(userID int64, roles []string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ParseToken parses and validates a JWT token
func (s *jwtService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshToken refreshes an existing token
func (s *jwtService) RefreshToken(tokenString string) (string, error) {
	// Parse the token without validating expiration
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", ErrInvalidToken
	}

	// Generate a new token with the same user ID and roles
	return s.GenerateToken(claims.UserID, claims.Roles)
}
