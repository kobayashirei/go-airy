package service

import (
	"context"
	"testing"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCircleRepository is a mock implementation of CircleRepository
type MockCircleRepository struct {
	mock.Mock
}

func (m *MockCircleRepository) Create(ctx context.Context, circle *models.Circle) error {
	args := m.Called(ctx, circle)
	if args.Get(0) != nil {
		circle.ID = 1 // Set ID for created circle
	}
	return args.Error(0)
}

func (m *MockCircleRepository) FindByID(ctx context.Context, id int64) (*models.Circle, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Circle), args.Error(1)
}

func (m *MockCircleRepository) FindByName(ctx context.Context, name string) (*models.Circle, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Circle), args.Error(1)
}

func (m *MockCircleRepository) FindAll(ctx context.Context, limit, offset int) ([]*models.Circle, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Circle), args.Error(1)
}

func (m *MockCircleRepository) FindByCreatorID(ctx context.Context, creatorID int64) ([]*models.Circle, error) {
	args := m.Called(ctx, creatorID)
	return args.Get(0).([]*models.Circle), args.Error(1)
}

func (m *MockCircleRepository) Update(ctx context.Context, circle *models.Circle) error {
	args := m.Called(ctx, circle)
	return args.Error(0)
}

func (m *MockCircleRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCircleRepository) IncrementMemberCount(ctx context.Context, id int64, delta int) error {
	args := m.Called(ctx, id, delta)
	return args.Error(0)
}

func (m *MockCircleRepository) IncrementPostCount(ctx context.Context, id int64, delta int) error {
	args := m.Called(ctx, id, delta)
	return args.Error(0)
}

func (m *MockCircleRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// MockCircleMemberRepository is a mock implementation of CircleMemberRepository
type MockCircleMemberRepository struct {
	mock.Mock
}

func (m *MockCircleMemberRepository) Create(ctx context.Context, member *models.CircleMember) error {
	args := m.Called(ctx, member)
	if args.Get(0) != nil {
		member.ID = 1 // Set ID for created member
	}
	return args.Error(0)
}

func (m *MockCircleMemberRepository) FindByID(ctx context.Context, id int64) (*models.CircleMember, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CircleMember), args.Error(1)
}

func (m *MockCircleMemberRepository) FindByCircleAndUser(ctx context.Context, circleID, userID int64) (*models.CircleMember, error) {
	args := m.Called(ctx, circleID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CircleMember), args.Error(1)
}

func (m *MockCircleMemberRepository) FindByCircleID(ctx context.Context, circleID int64, role string) ([]*models.CircleMember, error) {
	args := m.Called(ctx, circleID, role)
	return args.Get(0).([]*models.CircleMember), args.Error(1)
}

func (m *MockCircleMemberRepository) FindByUserID(ctx context.Context, userID int64) ([]*models.CircleMember, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.CircleMember), args.Error(1)
}

func (m *MockCircleMemberRepository) Update(ctx context.Context, member *models.CircleMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockCircleMemberRepository) UpdateRole(ctx context.Context, id int64, role string) error {
	args := m.Called(ctx, id, role)
	return args.Error(0)
}

func (m *MockCircleMemberRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCircleMemberRepository) DeleteByCircleAndUser(ctx context.Context, circleID, userID int64) error {
	args := m.Called(ctx, circleID, userID)
	return args.Error(0)
}

func (m *MockCircleMemberRepository) CountByCircleID(ctx context.Context, circleID int64, role string) (int64, error) {
	args := m.Called(ctx, circleID, role)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCircleMemberRepository) IsMember(ctx context.Context, circleID, userID int64) (bool, error) {
	args := m.Called(ctx, circleID, userID)
	return args.Get(0).(bool), args.Error(1)
}

// MockRoleRepository is a mock implementation of RoleRepository
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(ctx context.Context, role *models.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) FindByID(ctx context.Context, id int64) (*models.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleRepository) FindAll(ctx context.Context) ([]*models.Role, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *models.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestCreateCircle tests the CreateCircle method
func TestCreateCircle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful circle creation", func(t *testing.T) {
		// Setup mocks
		circleRepo := new(MockCircleRepository)
		circleMemberRepo := new(MockCircleMemberRepository)
		userRepo := new(MockUserRepository)
		userRoleRepo := new(MockUserRoleRepository)
		roleRepo := new(MockRoleRepository)

		// Mock expectations
		user := &models.User{ID: 1, Username: "testuser"}
		userRepo.On("FindByID", ctx, int64(1)).Return(user, nil)
		circleRepo.On("FindByName", ctx, "Test Circle").Return(nil, nil)
		circleRepo.On("Create", ctx, mock.AnythingOfType("*models.Circle")).Return(nil)
		circleMemberRepo.On("Create", ctx, mock.AnythingOfType("*models.CircleMember")).Return(nil)

		// Create service
		service := NewCircleService(circleRepo, circleMemberRepo, userRepo, userRoleRepo, roleRepo)

		// Test
		req := CreateCircleRequest{
			Name:        "Test Circle",
			Description: "A test circle",
			CreatorID:   1,
		}
		circle, err := service.CreateCircle(ctx, req)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, circle)
		assert.Equal(t, "Test Circle", circle.Name)
		assert.Equal(t, int64(1), circle.CreatorID)
		assert.Equal(t, "public", circle.Status)
		assert.Equal(t, "free", circle.JoinRule)
		assert.Equal(t, 1, circle.MemberCount)

		circleRepo.AssertExpectations(t)
		circleMemberRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("circle name already exists", func(t *testing.T) {
		// Setup mocks
		circleRepo := new(MockCircleRepository)
		circleMemberRepo := new(MockCircleMemberRepository)
		userRepo := new(MockUserRepository)
		userRoleRepo := new(MockUserRoleRepository)
		roleRepo := new(MockRoleRepository)

		// Mock expectations
		user := &models.User{ID: 1, Username: "testuser"}
		existingCircle := &models.Circle{ID: 1, Name: "Test Circle"}
		userRepo.On("FindByID", ctx, int64(1)).Return(user, nil)
		circleRepo.On("FindByName", ctx, "Test Circle").Return(existingCircle, nil)

		// Create service
		service := NewCircleService(circleRepo, circleMemberRepo, userRepo, userRoleRepo, roleRepo)

		// Test
		req := CreateCircleRequest{
			Name:      "Test Circle",
			CreatorID: 1,
		}
		circle, err := service.CreateCircle(ctx, req)

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, ErrCircleNameExists, err)
		assert.Nil(t, circle)

		userRepo.AssertExpectations(t)
		circleRepo.AssertExpectations(t)
	})
}

// TestJoinCircle tests the JoinCircle method
func TestJoinCircle(t *testing.T) {
	ctx := context.Background()

	t.Run("free join circle", func(t *testing.T) {
		// Setup mocks
		circleRepo := new(MockCircleRepository)
		circleMemberRepo := new(MockCircleMemberRepository)
		userRepo := new(MockUserRepository)
		userRoleRepo := new(MockUserRoleRepository)
		roleRepo := new(MockRoleRepository)

		// Mock expectations
		circle := &models.Circle{ID: 1, Name: "Test Circle", JoinRule: "free"}
		user := &models.User{ID: 2, Username: "testuser"}
		circleRepo.On("FindByID", ctx, int64(1)).Return(circle, nil)
		userRepo.On("FindByID", ctx, int64(2)).Return(user, nil)
		circleMemberRepo.On("FindByCircleAndUser", ctx, int64(1), int64(2)).Return(nil, nil)
		circleMemberRepo.On("Create", ctx, mock.AnythingOfType("*models.CircleMember")).Return(nil)
		circleRepo.On("IncrementMemberCount", ctx, int64(1), 1).Return(nil)

		// Create service
		service := NewCircleService(circleRepo, circleMemberRepo, userRepo, userRoleRepo, roleRepo)

		// Test
		err := service.JoinCircle(ctx, 1, 2)

		// Assertions
		assert.NoError(t, err)

		circleRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		circleMemberRepo.AssertExpectations(t)
	})

	t.Run("approval required circle", func(t *testing.T) {
		// Setup mocks
		circleRepo := new(MockCircleRepository)
		circleMemberRepo := new(MockCircleMemberRepository)
		userRepo := new(MockUserRepository)
		userRoleRepo := new(MockUserRoleRepository)
		roleRepo := new(MockRoleRepository)

		// Mock expectations
		circle := &models.Circle{ID: 1, Name: "Test Circle", JoinRule: "approval"}
		user := &models.User{ID: 2, Username: "testuser"}
		circleRepo.On("FindByID", ctx, int64(1)).Return(circle, nil)
		userRepo.On("FindByID", ctx, int64(2)).Return(user, nil)
		circleMemberRepo.On("FindByCircleAndUser", ctx, int64(1), int64(2)).Return(nil, nil)
		circleMemberRepo.On("Create", ctx, mock.AnythingOfType("*models.CircleMember")).Return(nil)

		// Create service
		service := NewCircleService(circleRepo, circleMemberRepo, userRepo, userRoleRepo, roleRepo)

		// Test
		err := service.JoinCircle(ctx, 1, 2)

		// Assertions
		assert.NoError(t, err)

		circleRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		circleMemberRepo.AssertExpectations(t)
		// Note: IncrementMemberCount should NOT be called for approval-required circles
	})

	t.Run("already a member", func(t *testing.T) {
		// Setup mocks
		circleRepo := new(MockCircleRepository)
		circleMemberRepo := new(MockCircleMemberRepository)
		userRepo := new(MockUserRepository)
		userRoleRepo := new(MockUserRoleRepository)
		roleRepo := new(MockRoleRepository)

		// Mock expectations
		circle := &models.Circle{ID: 1, Name: "Test Circle", JoinRule: "free"}
		user := &models.User{ID: 2, Username: "testuser"}
		member := &models.CircleMember{ID: 1, CircleID: 1, UserID: 2, Role: "member"}
		circleRepo.On("FindByID", ctx, int64(1)).Return(circle, nil)
		userRepo.On("FindByID", ctx, int64(2)).Return(user, nil)
		circleMemberRepo.On("FindByCircleAndUser", ctx, int64(1), int64(2)).Return(member, nil)

		// Create service
		service := NewCircleService(circleRepo, circleMemberRepo, userRepo, userRoleRepo, roleRepo)

		// Test
		err := service.JoinCircle(ctx, 1, 2)

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, ErrAlreadyMember, err)

		circleRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		circleMemberRepo.AssertExpectations(t)
	})
}

// TestApproveMember tests the ApproveMember method
func TestApproveMember(t *testing.T) {
	ctx := context.Background()

	t.Run("successful approval", func(t *testing.T) {
		// Setup mocks
		circleRepo := new(MockCircleRepository)
		circleMemberRepo := new(MockCircleMemberRepository)
		userRepo := new(MockUserRepository)
		userRoleRepo := new(MockUserRoleRepository)
		roleRepo := new(MockRoleRepository)

		// Mock expectations
		approverMember := &models.CircleMember{ID: 1, CircleID: 1, UserID: 2, Role: "moderator"}
		pendingMember := &models.CircleMember{ID: 2, CircleID: 1, UserID: 3, Role: "pending"}
		circleMemberRepo.On("FindByCircleAndUser", ctx, int64(1), int64(2)).Return(approverMember, nil)
		circleMemberRepo.On("FindByCircleAndUser", ctx, int64(1), int64(3)).Return(pendingMember, nil)
		circleMemberRepo.On("Update", ctx, mock.AnythingOfType("*models.CircleMember")).Return(nil)
		circleRepo.On("IncrementMemberCount", ctx, int64(1), 1).Return(nil)

		// Create service
		service := NewCircleService(circleRepo, circleMemberRepo, userRepo, userRoleRepo, roleRepo)

		// Test
		err := service.ApproveMember(ctx, 1, 3, 2)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, "member", pendingMember.Role)

		circleMemberRepo.AssertExpectations(t)
		circleRepo.AssertExpectations(t)
	})

	t.Run("not authorized", func(t *testing.T) {
		// Setup mocks
		circleRepo := new(MockCircleRepository)
		circleMemberRepo := new(MockCircleMemberRepository)
		userRepo := new(MockUserRepository)
		userRoleRepo := new(MockUserRoleRepository)
		roleRepo := new(MockRoleRepository)

		// Mock expectations - approver is not a moderator
		approverMember := &models.CircleMember{ID: 1, CircleID: 1, UserID: 2, Role: "member"}
		circleMemberRepo.On("FindByCircleAndUser", ctx, int64(1), int64(2)).Return(approverMember, nil)

		// Create service
		service := NewCircleService(circleRepo, circleMemberRepo, userRepo, userRoleRepo, roleRepo)

		// Test
		err := service.ApproveMember(ctx, 1, 3, 2)

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, ErrNotAuthorized, err)

		circleMemberRepo.AssertExpectations(t)
	})
}

// TestAssignModerator tests the AssignModerator method
func TestAssignModerator(t *testing.T) {
	ctx := context.Background()

	t.Run("successful moderator assignment", func(t *testing.T) {
		// Setup mocks
		circleRepo := new(MockCircleRepository)
		circleMemberRepo := new(MockCircleMemberRepository)
		userRepo := new(MockUserRepository)
		userRoleRepo := new(MockUserRoleRepository)
		roleRepo := new(MockRoleRepository)

		// Mock expectations
		assignerMember := &models.CircleMember{ID: 1, CircleID: 1, UserID: 2, Role: "moderator"}
		targetMember := &models.CircleMember{ID: 2, CircleID: 1, UserID: 3, Role: "member"}
		moderatorRole := &models.Role{ID: 2, Name: "moderator"}
		
		circleMemberRepo.On("FindByCircleAndUser", ctx, int64(1), int64(2)).Return(assignerMember, nil)
		circleMemberRepo.On("FindByCircleAndUser", ctx, int64(1), int64(3)).Return(targetMember, nil)
		circleMemberRepo.On("Update", ctx, mock.AnythingOfType("*models.CircleMember")).Return(nil)
		roleRepo.On("FindByName", ctx, "moderator").Return(moderatorRole, nil)
		userRoleRepo.On("Create", ctx, mock.AnythingOfType("*models.UserRole")).Return(nil)

		// Create service
		service := NewCircleService(circleRepo, circleMemberRepo, userRepo, userRoleRepo, roleRepo)

		// Test
		err := service.AssignModerator(ctx, 1, 3, 2)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, "moderator", targetMember.Role)

		circleMemberRepo.AssertExpectations(t)
		roleRepo.AssertExpectations(t)
		userRoleRepo.AssertExpectations(t)
	})
}
