package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/repository"
)

var (
	ErrCircleNotFound    = errors.New("circle not found")
	ErrCircleNameExists  = errors.New("circle name already exists")
	ErrAlreadyMember     = errors.New("user is already a member")
	ErrPendingApproval   = errors.New("membership request is pending approval")
	ErrNotMember         = errors.New("user is not a member")
	ErrInvalidRole       = errors.New("invalid role")
	ErrCannotApproveSelf = errors.New("cannot approve own membership")
	ErrNotAuthorized     = errors.New("not authorized to perform this action")
)

// CircleService defines the interface for circle business logic
type CircleService interface {
	CreateCircle(ctx context.Context, req CreateCircleRequest) (*models.Circle, error)
	GetCircle(ctx context.Context, circleID int64) (*models.Circle, error)
	JoinCircle(ctx context.Context, circleID, userID int64) error
	ApproveMember(ctx context.Context, circleID, userID, approverID int64) error
	AssignModerator(ctx context.Context, circleID, userID, assignerID int64) error
	GetCircleMembers(ctx context.Context, circleID int64, role string) ([]*models.CircleMember, error)
	IsUserModerator(ctx context.Context, circleID, userID int64) (bool, error)
}

// circleService implements CircleService interface
type circleService struct {
	circleRepo       repository.CircleRepository
	circleMemberRepo repository.CircleMemberRepository
	userRepo         repository.UserRepository
	userRoleRepo     repository.UserRoleRepository
	roleRepo         repository.RoleRepository
}

// NewCircleService creates a new circle service
func NewCircleService(
	circleRepo repository.CircleRepository,
	circleMemberRepo repository.CircleMemberRepository,
	userRepo repository.UserRepository,
	userRoleRepo repository.UserRoleRepository,
	roleRepo repository.RoleRepository,
) CircleService {
	return &circleService{
		circleRepo:       circleRepo,
		circleMemberRepo: circleMemberRepo,
		userRepo:         userRepo,
		userRoleRepo:     userRoleRepo,
		roleRepo:         roleRepo,
	}
}

// CreateCircleRequest represents the request to create a circle
type CreateCircleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=1000"`
	Avatar      string `json:"avatar" binding:"omitempty,url"`
	Background  string `json:"background" binding:"omitempty,url"`
	Status      string `json:"status" binding:"omitempty,oneof=public semi_public private"`
	JoinRule    string `json:"join_rule" binding:"omitempty,oneof=free approval"`
	CreatorID   int64  `json:"creator_id" binding:"required"`
}

// CreateCircle creates a new circle
// Requirement 7.1: WHEN a User creates a Circle THEN the System SHALL store the Circle information with creator ID and initial status
func (s *circleService) CreateCircle(ctx context.Context, req CreateCircleRequest) (*models.Circle, error) {
	// Check if user exists
	user, err := s.userRepo.FindByID(ctx, req.CreatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check if circle name already exists
	existingCircle, err := s.circleRepo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check circle name: %w", err)
	}
	if existingCircle != nil {
		return nil, ErrCircleNameExists
	}

	// Set defaults
	status := req.Status
	if status == "" {
		status = "public"
	}
	joinRule := req.JoinRule
	if joinRule == "" {
		joinRule = "free"
	}

	// Create circle
	circle := &models.Circle{
		Name:        req.Name,
		Description: req.Description,
		Avatar:      req.Avatar,
		Background:  req.Background,
		CreatorID:   req.CreatorID,
		Status:      status,
		JoinRule:    joinRule,
		MemberCount: 1, // Creator is the first member
		PostCount:   0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.circleRepo.Create(ctx, circle); err != nil {
		return nil, fmt.Errorf("failed to create circle: %w", err)
	}

	// Add creator as a member with moderator role
	member := &models.CircleMember{
		CircleID:  circle.ID,
		UserID:    req.CreatorID,
		Role:      "moderator",
		JoinedAt:  time.Now(),
		CreatedAt: time.Now(),
	}

	if err := s.circleMemberRepo.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add creator as member: %w", err)
	}

	return circle, nil
}

// GetCircle retrieves a circle by ID
func (s *circleService) GetCircle(ctx context.Context, circleID int64) (*models.Circle, error) {
	circle, err := s.circleRepo.FindByID(ctx, circleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find circle: %w", err)
	}
	if circle == nil {
		return nil, ErrCircleNotFound
	}
	return circle, nil
}

// JoinCircle allows a user to join a circle
// Requirement 7.2: WHEN a User requests to join a Circle THEN the System SHALL process the request based on the Circle join rules
// Requirement 7.3: WHERE a Circle requires approval for joining THEN the System SHALL create a pending membership record for Moderator review
func (s *circleService) JoinCircle(ctx context.Context, circleID, userID int64) error {
	// Check if circle exists
	circle, err := s.circleRepo.FindByID(ctx, circleID)
	if err != nil {
		return fmt.Errorf("failed to find circle: %w", err)
	}
	if circle == nil {
		return ErrCircleNotFound
	}

	// Check if user exists
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Check if user is already a member or has pending request
	existingMember, err := s.circleMemberRepo.FindByCircleAndUser(ctx, circleID, userID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if existingMember != nil {
		if existingMember.Role == "pending" {
			return ErrPendingApproval
		}
		return ErrAlreadyMember
	}

	// Determine role based on join rule
	role := "member"
	if circle.JoinRule == "approval" {
		role = "pending"
	}

	// Create membership record
	member := &models.CircleMember{
		CircleID:  circleID,
		UserID:    userID,
		Role:      role,
		JoinedAt:  time.Now(),
		CreatedAt: time.Now(),
	}

	if err := s.circleMemberRepo.Create(ctx, member); err != nil {
		return fmt.Errorf("failed to create membership: %w", err)
	}

	// If free join, increment member count
	if circle.JoinRule == "free" {
		if err := s.circleRepo.IncrementMemberCount(ctx, circleID, 1); err != nil {
			return fmt.Errorf("failed to increment member count: %w", err)
		}
	}

	return nil
}

// ApproveMember approves a pending membership request
// Requirement 7.4: WHEN a Moderator approves a membership request THEN the System SHALL update the User-Circle relationship to active member
func (s *circleService) ApproveMember(ctx context.Context, circleID, userID, approverID int64) error {
	// Check if approver is a moderator
	isModerator, err := s.IsUserModerator(ctx, circleID, approverID)
	if err != nil {
		return fmt.Errorf("failed to check moderator status: %w", err)
	}
	if !isModerator {
		return ErrNotAuthorized
	}

	// Cannot approve self
	if userID == approverID {
		return ErrCannotApproveSelf
	}

	// Find the pending membership
	member, err := s.circleMemberRepo.FindByCircleAndUser(ctx, circleID, userID)
	if err != nil {
		return fmt.Errorf("failed to find membership: %w", err)
	}
	if member == nil {
		return ErrNotMember
	}
	if member.Role != "pending" {
		return errors.New("membership is not pending approval")
	}

	// Update role to member
	member.Role = "member"
	member.JoinedAt = time.Now()
	if err := s.circleMemberRepo.Update(ctx, member); err != nil {
		return fmt.Errorf("failed to update membership: %w", err)
	}

	// Increment member count
	if err := s.circleRepo.IncrementMemberCount(ctx, circleID, 1); err != nil {
		return fmt.Errorf("failed to increment member count: %w", err)
	}

	return nil
}

// AssignModerator assigns a user as a moderator of a circle
// Requirement 7.5: WHEN a User is assigned as Moderator THEN the System SHALL grant Circle-specific moderation Permissions
func (s *circleService) AssignModerator(ctx context.Context, circleID, userID, assignerID int64) error {
	// Check if assigner is a moderator or creator
	isModerator, err := s.IsUserModerator(ctx, circleID, assignerID)
	if err != nil {
		return fmt.Errorf("failed to check moderator status: %w", err)
	}
	if !isModerator {
		return ErrNotAuthorized
	}

	// Check if user is a member
	member, err := s.circleMemberRepo.FindByCircleAndUser(ctx, circleID, userID)
	if err != nil {
		return fmt.Errorf("failed to find membership: %w", err)
	}
	if member == nil {
		return ErrNotMember
	}
	if member.Role == "pending" {
		return errors.New("cannot assign moderator to pending member")
	}

	// Update role to moderator
	member.Role = "moderator"
	if err := s.circleMemberRepo.Update(ctx, member); err != nil {
		return fmt.Errorf("failed to update membership: %w", err)
	}

	// Assign moderator role in user_roles table for circle-specific permissions
	moderatorRole, err := s.roleRepo.FindByName(ctx, "moderator")
	if err != nil {
		return fmt.Errorf("failed to find moderator role: %w", err)
	}
	if moderatorRole == nil {
		return errors.New("moderator role not found in system")
	}

	userRole := &models.UserRole{
		UserID:    userID,
		RoleID:    moderatorRole.ID,
		CircleID:  &circleID,
		CreatedAt: time.Now(),
	}

	if err := s.userRoleRepo.Create(ctx, userRole); err != nil {
		// Ignore duplicate key errors (user might already have the role)
		if !errors.Is(err, errors.New("duplicate key")) {
			return fmt.Errorf("failed to assign moderator role: %w", err)
		}
	}

	return nil
}

// GetCircleMembers retrieves members of a circle, optionally filtered by role
func (s *circleService) GetCircleMembers(ctx context.Context, circleID int64, role string) ([]*models.CircleMember, error) {
	members, err := s.circleMemberRepo.FindByCircleID(ctx, circleID, role)
	if err != nil {
		return nil, fmt.Errorf("failed to find members: %w", err)
	}
	return members, nil
}

// IsUserModerator checks if a user is a moderator of a circle
func (s *circleService) IsUserModerator(ctx context.Context, circleID, userID int64) (bool, error) {
	member, err := s.circleMemberRepo.FindByCircleAndUser(ctx, circleID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to find membership: %w", err)
	}
	if member == nil {
		return false, nil
	}
	return member.Role == "moderator", nil
}
