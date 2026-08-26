package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/user/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*domain.User, error) {
	args := m.Called(ctx, firebaseUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) Create(ctx context.Context, u *domain.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepository) UpdateUser(ctx context.Context, u *domain.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepository) UpdateAvatar(ctx context.Context, userID string, avatarURL string) (*domain.User, error) {
	args := m.Called(ctx, userID, avatarURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) GetByEmailOrPhone(ctx context.Context, email, phone string) (*domain.User, error) {
	args := m.Called(ctx, email, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) GetByEmailOrPhoneWithSettings(ctx context.Context, email, phone string) (*domain.UserWithSettings, error) {
	args := m.Called(ctx, email, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserWithSettings), args.Error(1)
}

func (m *mockUserRepository) CreateDefaultSettings(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockUserRepository) GetSettings(ctx context.Context, userID string) (*domain.UserSettings, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserSettings), args.Error(1)
}

func (m *mockUserRepository) UpsertSettings(ctx context.Context, settings *domain.UserSettings) error {
	return m.Called(ctx, settings).Error(0)
}

func (m *mockUserRepository) CreateFriendship(ctx context.Context, userID, friendID string, status domain.FriendshipStatus, actionUserID string) error {
	return m.Called(ctx, userID, friendID, status, actionUserID).Error(0)
}

func (m *mockUserRepository) UpdateFriendshipStatus(ctx context.Context, userID, friendID string, status domain.FriendshipStatus, actionUserID string) error {
	return m.Called(ctx, userID, friendID, status, actionUserID).Error(0)
}

func (m *mockUserRepository) DeleteFriendship(ctx context.Context, userID, friendID string) error {
	return m.Called(ctx, userID, friendID).Error(0)
}

func (m *mockUserRepository) GetFriendship(ctx context.Context, userID, friendID string) (*domain.Friendship, error) {
	args := m.Called(ctx, userID, friendID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Friendship), args.Error(1)
}

func (m *mockUserRepository) ListFriends(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.User, error) {
	args := m.Called(ctx, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *mockUserRepository) ListFriendsByStatus(ctx context.Context, userID string, status domain.FriendshipStatus) ([]domain.FriendWithStatus, error) {
	args := m.Called(ctx, userID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.FriendWithStatus), args.Error(1)
}

func (m *mockUserRepository) SyncFriendsBySequence(ctx context.Context, lastVersion int64, userID string, limit int32) ([]domain.FriendshipSyncRecord, error) {
	args := m.Called(ctx, lastVersion, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.FriendshipSyncRecord), args.Error(1)
}

// --- RegisterUser Tests ---

func TestRegisterUser_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	email := "alice@example.com"
	mockRepo.On("GetByFirebaseUID", ctx, "fb-123").Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
	mockRepo.On("CreateDefaultSettings", ctx, mock.AnythingOfType("string")).Return(nil)

	uc := domain.NewUseCase(mockRepo, nil)
	u, err := uc.RegisterUser(ctx, "fb-123", &email, nil, "Alice")
	require.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, "fb-123", u.FirebaseUID)
	assert.Equal(t, "Alice", u.Name)
	assert.Equal(t, &email, u.Email)
	mockRepo.AssertExpectations(t)
}

func TestRegisterUser_AlreadyExists(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	email := "alice@example.com"
	existingUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice", Email: &email}
	mockRepo.On("GetByFirebaseUID", ctx, "fb-123").Return(existingUser, nil)

	uc := domain.NewUseCase(mockRepo, nil)
	u, err := uc.RegisterUser(ctx, "fb-123", &email, nil, "Alice")
	require.NoError(t, err)
	assert.Equal(t, existingUser, u)
	mockRepo.AssertExpectations(t)
}

// --- AddFriendByEmailOrPhone Tests ---

func TestAddFriendByEmailOrPhone_PendingWhenAutoAcceptFalse(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	friendUserWithSettings := &domain.UserWithSettings{
		User:                     domain.User{ID: "usr-2", Name: "Bob"},
		AutoAcceptFriendRequests: false,
	}
	mockRepo.On("GetByEmailOrPhoneWithSettings", ctx, "bob@example.com", "").Return(friendUserWithSettings, nil)
	mockRepo.On("GetFriendship", ctx, "usr-1", "usr-2").Return(nil, nil)
	mockRepo.On("CreateFriendship", ctx, "usr-1", "usr-2", domain.Pending, "usr-1").Return(nil)

	uc := domain.NewUseCase(mockRepo, nil)
	friend, err := uc.AddFriendByEmailOrPhone(ctx, "usr-1", "bob@example.com", "")
	require.NoError(t, err)
	assert.Equal(t, "usr-2", friend.ID)
	assert.Equal(t, domain.Pending, friend.Status)
	mockRepo.AssertExpectations(t)
}

func TestAddFriendByEmailOrPhone_AcceptedWhenAutoAcceptTrue(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	friendUserWithSettings := &domain.UserWithSettings{
		User:                     domain.User{ID: "usr-2", Name: "Bob"},
		AutoAcceptFriendRequests: true,
	}
	mockRepo.On("GetByEmailOrPhoneWithSettings", ctx, "bob@example.com", "").Return(friendUserWithSettings, nil)
	mockRepo.On("GetFriendship", ctx, "usr-1", "usr-2").Return(nil, nil)
	mockRepo.On("CreateFriendship", ctx, "usr-1", "usr-2", domain.Accepted, "usr-1").Return(nil)

	uc := domain.NewUseCase(mockRepo, nil)
	friend, err := uc.AddFriendByEmailOrPhone(ctx, "usr-1", "bob@example.com", "")
	require.NoError(t, err)
	assert.Equal(t, "usr-2", friend.ID)
	assert.Equal(t, domain.Accepted, friend.Status)
	mockRepo.AssertExpectations(t)
}

func TestAddFriendByEmailOrPhone_BlockedUserError(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	friendUserWithSettings := &domain.UserWithSettings{
		User:                     domain.User{ID: "usr-2", Name: "Bob"},
		AutoAcceptFriendRequests: true,
	}
	mockRepo.On("GetByEmailOrPhoneWithSettings", ctx, "bob@example.com", "").Return(friendUserWithSettings, nil)
	mockRepo.On("GetFriendship", ctx, "usr-1", "usr-2").Return(&domain.Friendship{
		UserID:   "usr-1",
		FriendID: "usr-2",
		Status:   domain.Blocked,
	}, nil)

	uc := domain.NewUseCase(mockRepo, nil)
	_, err := uc.AddFriendByEmailOrPhone(ctx, "usr-1", "bob@example.com", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "blocked")
}

// --- UpdateFriendshipStatus Tests ---

func TestUpdateFriendshipStatus_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	mockRepo.On("GetFriendship", ctx, "usr-1", "usr-2").Return(&domain.Friendship{
		UserID:   "usr-2",
		FriendID: "usr-1",
		Status:   domain.Pending,
	}, nil)
	mockRepo.On("UpdateFriendshipStatus", ctx, "usr-1", "usr-2", domain.Accepted, "usr-1").Return(nil)
	mockRepo.On("GetByID", ctx, "usr-2").Return(&domain.User{ID: "usr-2", Name: "Bob"}, nil)

	uc := domain.NewUseCase(mockRepo, nil)
	friend, err := uc.UpdateFriendshipStatus(ctx, "usr-1", "usr-2", domain.Accepted)
	require.NoError(t, err)
	assert.Equal(t, "usr-2", friend.ID)
	assert.Equal(t, domain.Accepted, friend.Status)
	mockRepo.AssertExpectations(t)
}

// --- UserSettings Tests ---

func TestGetUserSettings_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	expectedSettings := &domain.UserSettings{UserID: "usr-1", AutoAcceptFriendRequests: true}
	mockRepo.On("GetSettings", ctx, "usr-1").Return(expectedSettings, nil)

	uc := domain.NewUseCase(mockRepo, nil)
	settings, err := uc.GetUserSettings(ctx, "usr-1")
	require.NoError(t, err)
	assert.Equal(t, expectedSettings, settings)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUserSettings_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	mockRepo.On("UpsertSettings", ctx, mock.AnythingOfType("*domain.UserSettings")).Return(nil)

	uc := domain.NewUseCase(mockRepo, nil)
	settings, err := uc.UpdateUserSettings(ctx, "usr-1", true)
	require.NoError(t, err)
	assert.True(t, settings.AutoAcceptFriendRequests)
	mockRepo.AssertExpectations(t)
}
