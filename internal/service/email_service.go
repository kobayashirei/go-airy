package service

import (
	"context"
	"fmt"
)

// EmailService defines the interface for email operations
type EmailService interface {
	SendActivationEmail(ctx context.Context, email, token string) error
	SendVerificationCode(ctx context.Context, email, code string) error
	SendPasswordResetEmail(ctx context.Context, email, token string) error
}

// emailService implements EmailService interface
type emailService struct {
	// In production, this would contain SMTP configuration or email service client
	// For now, we'll use a simple implementation
}

// NewEmailService creates a new email service
func NewEmailService() EmailService {
	return &emailService{}
}

// SendActivationEmail sends an activation email to the user
func (s *emailService) SendActivationEmail(ctx context.Context, email, token string) error {
	// In production, this would send an actual email using SMTP or email service API
	// For now, we'll just log the activation link
	activationLink := fmt.Sprintf("http://localhost:8080/api/v1/auth/activate?token=%s", token)
	fmt.Printf("Activation email sent to %s: %s\n", email, activationLink)
	return nil
}

// SendVerificationCode sends a verification code to the user
func (s *emailService) SendVerificationCode(ctx context.Context, email, code string) error {
	// In production, this would send an actual email
	fmt.Printf("Verification code sent to %s: %s\n", email, code)
	return nil
}

// SendPasswordResetEmail sends a password reset email to the user
func (s *emailService) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	// In production, this would send an actual email
	resetLink := fmt.Sprintf("http://localhost:8080/api/v1/auth/reset-password?token=%s", token)
	fmt.Printf("Password reset email sent to %s: %s\n", email, resetLink)
	return nil
}
