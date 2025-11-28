package service

import (
	"context"

	"github.com/kobayashirei/airy/internal/models"
	"github.com/kobayashirei/airy/internal/mq"
	"github.com/kobayashirei/airy/internal/repository"
	"github.com/stretchr/testify/mock"
)

// MockCommentRepository is a mock implementation of CommentRepository
type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) Create(ctx context.Context, comment *models.Comment) error {
	args := m.Called(ctx, comment)
	// Simulate ID assignment
	if args.Error(0) == nil && comment.ID == 0 {
		comment.ID = 1
	}
	return args.Error(0)
}

func (m *MockCommentRepository) FindByID(ctx context.Context, id int64) (*models.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Comment), args.Error(1)
}

func (m *MockCommentRepository) FindByPostID(ctx context.Context, postID int64) ([]*models.Comment, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Comment), args.Error(1)
}

func (m *MockCommentRepository) FindByParentID(ctx context.Context, parentID int64) ([]*models.Comment, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]*models.Comment), args.Error(1)
}

func (m *MockCommentRepository) FindRootComments(ctx context.Context, postID int64, limit, offset int) ([]*models.Comment, error) {
	args := m.Called(ctx, postID, limit, offset)
	return args.Get(0).([]*models.Comment), args.Error(1)
}

func (m *MockCommentRepository) Update(ctx context.Context, comment *models.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommentRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockCommentRepository) CountByPostID(ctx context.Context, postID int64) (int64, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCommentRepository) Count(ctx context.Context, opts repository.CommentListOptions) (int64, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(int64), args.Error(1)
}

// MockPostRepository is a mock implementation of PostRepository
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(ctx context.Context, post *models.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) FindByID(ctx context.Context, id int64) (*models.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) Update(ctx context.Context, post *models.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) List(ctx context.Context, opts repository.PostListOptions) ([]*models.Post, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Post), args.Error(1)
}

func (m *MockPostRepository) Count(ctx context.Context, opts repository.PostListOptions) (int64, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPostRepository) IncrementViewCount(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) UpdateHotnessScore(ctx context.Context, id int64, score float64) error {
	args := m.Called(ctx, id, score)
	return args.Error(0)
}

func (m *MockPostRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockPostRepository) CountByDate(ctx context.Context, date string) (int64, error) {
	args := m.Called(ctx, date)
	return args.Get(0).(int64), args.Error(1)
}

// MockMessageQueue is a mock implementation of MessageQueue
type MockMessageQueue struct {
	mock.Mock
}

func (m *MockMessageQueue) Publish(ctx context.Context, topic string, message interface{}) error {
	args := m.Called(ctx, topic, message)
	return args.Error(0)
}

func (m *MockMessageQueue) Subscribe(topic string, handler mq.MessageHandler) error {
	args := m.Called(topic, handler)
	return args.Error(0)
}

func (m *MockMessageQueue) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockVoteRepository is a mock implementation of VoteRepository
type MockVoteRepository struct {
	mock.Mock
}

func (m *MockVoteRepository) Create(ctx context.Context, vote *models.Vote) error {
	args := m.Called(ctx, vote)
	// Simulate ID assignment
	if args.Error(0) == nil && vote.ID == 0 {
		vote.ID = 1
	}
	return args.Error(0)
}

func (m *MockVoteRepository) FindByID(ctx context.Context, id int64) (*models.Vote, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vote), args.Error(1)
}

func (m *MockVoteRepository) FindByUserAndEntity(ctx context.Context, userID int64, entityType string, entityID int64) (*models.Vote, error) {
	args := m.Called(ctx, userID, entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vote), args.Error(1)
}

func (m *MockVoteRepository) Update(ctx context.Context, vote *models.Vote) error {
	args := m.Called(ctx, vote)
	return args.Error(0)
}

func (m *MockVoteRepository) Upsert(ctx context.Context, vote *models.Vote) error {
	args := m.Called(ctx, vote)
	return args.Error(0)
}

func (m *MockVoteRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockVoteRepository) DeleteByUserAndEntity(ctx context.Context, userID int64, entityType string, entityID int64) error {
	args := m.Called(ctx, userID, entityType, entityID)
	return args.Error(0)
}

func (m *MockVoteRepository) CountByEntity(ctx context.Context, entityType string, entityID int64, voteType string) (int64, error) {
	args := m.Called(ctx, entityType, entityID, voteType)
	return args.Get(0).(int64), args.Error(1)
}

// MockPublisher is a mock implementation of Publisher
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) PublishVoteCreated(ctx context.Context, voteID, userID int64, entityType string, entityID int64, voteType string) error {
	args := m.Called(ctx, voteID, userID, entityType, entityID, voteType)
	return args.Error(0)
}

func (m *MockPublisher) PublishVoteUpdated(ctx context.Context, voteID, userID int64, entityType string, entityID int64, oldVoteType, newVoteType string) error {
	args := m.Called(ctx, voteID, userID, entityType, entityID, oldVoteType, newVoteType)
	return args.Error(0)
}

func (m *MockPublisher) PublishVoteDeleted(ctx context.Context, voteID, userID int64, entityType string, entityID int64, voteType string) error {
	args := m.Called(ctx, voteID, userID, entityType, entityID, voteType)
	return args.Error(0)
}

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	if args.Error(0) == nil && user.ID == 0 {
		user.ID = 1
	}
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLoginInfo(ctx context.Context, id int64, ip string) error {
	args := m.Called(ctx, id, ip)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, opts repository.UserListOptions) ([]*models.User, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context, opts repository.UserListOptions) (int64, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CountByLastLogin(ctx context.Context, date string) (int64, error) {
	args := m.Called(ctx, date)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// MockUserRoleRepository is a mock implementation of UserRoleRepository
type MockUserRoleRepository struct {
	mock.Mock
}

func (m *MockUserRoleRepository) Create(ctx context.Context, userRole *models.UserRole) error {
	args := m.Called(ctx, userRole)
	return args.Error(0)
}

func (m *MockUserRoleRepository) FindByUserID(ctx context.Context, userID int64) ([]*models.UserRole, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserRole), args.Error(1)
}

func (m *MockUserRoleRepository) FindByUserIDAndCircleID(ctx context.Context, userID int64, circleID *int64) ([]*models.UserRole, error) {
	args := m.Called(ctx, userID, circleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserRole), args.Error(1)
}

func (m *MockUserRoleRepository) FindRolesByUserID(ctx context.Context, userID int64) ([]*models.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockUserRoleRepository) Delete(ctx context.Context, userID, roleID int64, circleID *int64) error {
	args := m.Called(ctx, userID, roleID, circleID)
	return args.Error(0)
}

func (m *MockUserRoleRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockUserProfileRepository is a mock implementation of UserProfileRepository
type MockUserProfileRepository struct {
	mock.Mock
}

func (m *MockUserProfileRepository) Create(ctx context.Context, profile *models.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockUserProfileRepository) FindByUserID(ctx context.Context, userID int64) (*models.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserProfile), args.Error(1)
}

func (m *MockUserProfileRepository) Update(ctx context.Context, profile *models.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockUserProfileRepository) IncrementFollowerCount(ctx context.Context, userID int64, delta int) error {
	args := m.Called(ctx, userID, delta)
	return args.Error(0)
}

func (m *MockUserProfileRepository) IncrementFollowingCount(ctx context.Context, userID int64, delta int) error {
	args := m.Called(ctx, userID, delta)
	return args.Error(0)
}

func (m *MockUserProfileRepository) UpdatePoints(ctx context.Context, userID int64, points int) error {
	args := m.Called(ctx, userID, points)
	return args.Error(0)
}

// MockNotificationRepository is a mock implementation of NotificationRepository
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	args := m.Called(ctx, notification)
	if args.Error(0) == nil && notification.ID == 0 {
		notification.ID = 1
	}
	return args.Error(0)
}

func (m *MockNotificationRepository) FindByID(ctx context.Context, id int64) (*models.Notification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) FindByReceiverID(ctx context.Context, receiverID int64, limit, offset int) ([]*models.Notification, error) {
	args := m.Called(ctx, receiverID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) FindUnreadByReceiverID(ctx context.Context, receiverID int64, limit, offset int) ([]*models.Notification, error) {
	args := m.Called(ctx, receiverID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) Update(ctx context.Context, notification *models.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkAsRead(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkAllAsRead(ctx context.Context, receiverID int64) error {
	args := m.Called(ctx, receiverID)
	return args.Error(0)
}

func (m *MockNotificationRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNotificationRepository) CountUnreadByReceiverID(ctx context.Context, receiverID int64) (int64, error) {
	args := m.Called(ctx, receiverID)
	return args.Get(0).(int64), args.Error(1)
}

// MockConversationRepository is a mock implementation of ConversationRepository
type MockConversationRepository struct {
	mock.Mock
}

func (m *MockConversationRepository) Create(ctx context.Context, conversation *models.Conversation) error {
	args := m.Called(ctx, conversation)
	if args.Error(0) == nil && conversation.ID == 0 {
		conversation.ID = 1
	}
	return args.Error(0)
}

func (m *MockConversationRepository) FindByID(ctx context.Context, id int64) (*models.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Conversation), args.Error(1)
}

func (m *MockConversationRepository) FindByUsers(ctx context.Context, user1ID, user2ID int64) (*models.Conversation, error) {
	args := m.Called(ctx, user1ID, user2ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Conversation), args.Error(1)
}

func (m *MockConversationRepository) FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]*models.Conversation, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Conversation), args.Error(1)
}

func (m *MockConversationRepository) Update(ctx context.Context, conversation *models.Conversation) error {
	args := m.Called(ctx, conversation)
	return args.Error(0)
}

func (m *MockConversationRepository) UpdateLastMessageAt(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockConversationRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockMessageRepository is a mock implementation of MessageRepository
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(ctx context.Context, message *models.Message) error {
	args := m.Called(ctx, message)
	if args.Error(0) == nil && message.ID == 0 {
		message.ID = 1
	}
	return args.Error(0)
}

func (m *MockMessageRepository) FindByID(ctx context.Context, id int64) (*models.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageRepository) FindByConversationID(ctx context.Context, conversationID int64, limit, offset int) ([]*models.Message, error) {
	args := m.Called(ctx, conversationID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockMessageRepository) FindUnreadByConversationAndReceiver(ctx context.Context, conversationID, receiverID int64) ([]*models.Message, error) {
	args := m.Called(ctx, conversationID, receiverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockMessageRepository) Update(ctx context.Context, message *models.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) MarkAsRead(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMessageRepository) MarkConversationAsRead(ctx context.Context, conversationID, receiverID int64) error {
	args := m.Called(ctx, conversationID, receiverID)
	return args.Error(0)
}

func (m *MockMessageRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMessageRepository) CountUnreadByReceiver(ctx context.Context, receiverID int64) (int64, error) {
	args := m.Called(ctx, receiverID)
	return args.Get(0).(int64), args.Error(1)
}
