package service

import (
	"context"
	"fmt"

	"github.com/kobayashirei/airy/internal/repository"
)

// PermissionService defines the interface for permission operations
type PermissionService interface {
	CheckPermission(ctx context.Context, roles []string, permission string) (bool, error)
	CheckPermissionWithCircle(ctx context.Context, userID int64, circleID *int64, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID int64, circleID *int64) ([]string, error)
}

// permissionService implements PermissionService interface
type permissionService struct {
	permissionRepo repository.PermissionRepository
	roleRepo       repository.RoleRepository
	userRoleRepo   repository.UserRoleRepository
}

// NewPermissionService creates a new permission service
func NewPermissionService(
	permissionRepo repository.PermissionRepository,
	roleRepo repository.RoleRepository,
	userRoleRepo repository.UserRoleRepository,
) PermissionService {
	return &permissionService{
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
	}
}

// CheckPermission checks if any of the given roles has the specified permission
func (s *permissionService) CheckPermission(ctx context.Context, roles []string, permission string) (bool, error) {
	if len(roles) == 0 {
		return false, nil
	}

	// Get all permissions for each role
	for _, roleName := range roles {
		role, err := s.roleRepo.FindByName(ctx, roleName)
		if err != nil {
			return false, fmt.Errorf("failed to find role %s: %w", roleName, err)
		}
		if role == nil {
			continue
		}

		permissions, err := s.permissionRepo.FindByRoleID(ctx, role.ID)
		if err != nil {
			return false, fmt.Errorf("failed to find permissions for role %d: %w", role.ID, err)
		}

		// Check if the permission exists in the role's permissions
		for _, perm := range permissions {
			if perm.Name == permission {
				return true, nil
			}
		}
	}

	return false, nil
}

// CheckPermissionWithCircle checks if a user has a specific permission, considering circle-specific roles
func (s *permissionService) CheckPermissionWithCircle(ctx context.Context, userID int64, circleID *int64, permission string) (bool, error) {
	// Get user roles for the specific circle (or global if circleID is nil)
	userRoles, err := s.userRoleRepo.FindByUserIDAndCircleID(ctx, userID, circleID)
	if err != nil {
		return false, fmt.Errorf("failed to find user roles: %w", err)
	}

	// If no circle-specific roles found and circleID is not nil, also check global roles
	if len(userRoles) == 0 && circleID != nil {
		userRoles, err = s.userRoleRepo.FindByUserIDAndCircleID(ctx, userID, nil)
		if err != nil {
			return false, fmt.Errorf("failed to find global user roles: %w", err)
		}
	}

	// Extract role names
	roleNames := make([]string, 0, len(userRoles))
	for _, ur := range userRoles {
		role, err := s.roleRepo.FindByID(ctx, ur.RoleID)
		if err != nil {
			return false, fmt.Errorf("failed to find role %d: %w", ur.RoleID, err)
		}
		if role != nil {
			roleNames = append(roleNames, role.Name)
		}
	}

	return s.CheckPermission(ctx, roleNames, permission)
}

// GetUserPermissions retrieves all permissions for a user, considering circle-specific roles
func (s *permissionService) GetUserPermissions(ctx context.Context, userID int64, circleID *int64) ([]string, error) {
	// Get user roles for the specific circle (or global if circleID is nil)
	userRoles, err := s.userRoleRepo.FindByUserIDAndCircleID(ctx, userID, circleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user roles: %w", err)
	}

	// If no circle-specific roles found and circleID is not nil, also check global roles
	if len(userRoles) == 0 && circleID != nil {
		userRoles, err = s.userRoleRepo.FindByUserIDAndCircleID(ctx, userID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to find global user roles: %w", err)
		}
	}

	// Collect all permissions from all roles (using a map to avoid duplicates)
	permissionMap := make(map[string]bool)
	for _, ur := range userRoles {
		permissions, err := s.permissionRepo.FindByRoleID(ctx, ur.RoleID)
		if err != nil {
			return nil, fmt.Errorf("failed to find permissions for role %d: %w", ur.RoleID, err)
		}

		for _, perm := range permissions {
			permissionMap[perm.Name] = true
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(permissionMap))
	for permName := range permissionMap {
		result = append(result, permName)
	}

	return result, nil
}
