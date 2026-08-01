package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
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

func (m *mockUserRepository) GetByEmailOrPhone(ctx context.Context, email, phone string) (*domain.User, error) {
	args := m.Called(ctx, email, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) CreateFriendship(ctx context.Context, userID, friendID string) error {
	return m.Called(ctx, userID, friendID).Error(0)
}

func (m *mockUserRepository) DeleteFriendship(ctx context.Context, userID, friendID string) error {
	return m.Called(ctx, userID, friendID).Error(0)
}

func (m *mockUserRepository) GetFriendship(ctx context.Context, userID, friendID string) (bool, error) {
	args := m.Called(ctx, userID, friendID)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepository) ListFriends(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.User, error) {
	args := m.Called(ctx, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

// --- RegisterUser Tests ---

func TestRegisterUser_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	email := "alice@example.com"
	mockRepo.On("GetByFirebaseUID", ctx, "fb-123").Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	uc := domain.NewUseCase(mockRepo)
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

	uc := domain.NewUseCase(mockRepo)
	u, err := uc.RegisterUser(ctx, "fb-123", &email, nil, "Alice")
	require.NoError(t, err)
	assert.Equal(t, existingUser, u)
	mockRepo.AssertExpectations(t)
}

func TestRegisterUser_MissingFirebaseUID(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()
	email := "alice@example.com"

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.RegisterUser(ctx, "", &email, nil, "Alice")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgInvalidParam)
}

func TestRegisterUser_MissingEmailAndPhone(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.RegisterUser(ctx, "fb-123", nil, nil, "Alice")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgMissingEmailOrPhone)
}

func TestRegisterUser_RepoCreateError(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()
	email := "alice@example.com"

	mockRepo.On("GetByFirebaseUID", ctx, "fb-123").Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(errors.New("db error"))

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.RegisterUser(ctx, "fb-123", &email, nil, "Alice")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeInternal, appErr.Type)
	assert.Contains(t, appErr.Message, "failed to register user")
	mockRepo.AssertExpectations(t)
}

// --- GetUserProfile Tests ---

func TestGetUserProfile_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()
	expectedUser := &domain.User{ID: "usr-1", Name: "Alice"}
	mockRepo.On("GetByID", ctx, "usr-1").Return(expectedUser, nil)

	uc := domain.NewUseCase(mockRepo)
	u, err := uc.GetUserProfile(ctx, "usr-1")
	require.NoError(t, err)
	assert.Equal(t, expectedUser, u)
	mockRepo.AssertExpectations(t)
}

func TestGetUserProfile_EmptyID(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.GetUserProfile(ctx, "")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgInvalidParam)
}

func TestGetUserProfile_NotFound(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()
	mockRepo.On("GetByID", ctx, "non-existent").Return(nil, nil)

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.GetUserProfile(ctx, "non-existent")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeNotFound, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgUserNotFound)
	mockRepo.AssertExpectations(t)
}

func TestGetUserProfile_RepoError(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()
	mockRepo.On("GetByID", ctx, "usr-1").Return(nil, errors.New("db error"))

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.GetUserProfile(ctx, "usr-1")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeInternal, appErr.Type)
	assert.Contains(t, appErr.Message, "failed to retrieve user profile")
	mockRepo.AssertExpectations(t)
}

// --- GetUserByFirebaseUID Tests ---

func TestGetUserByFirebaseUID_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()
	expectedUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123"}
	mockRepo.On("GetByFirebaseUID", ctx, "fb-123").Return(expectedUser, nil)

	uc := domain.NewUseCase(mockRepo)
	u, err := uc.GetUserByFirebaseUID(ctx, "fb-123")
	require.NoError(t, err)
	assert.Equal(t, expectedUser, u)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByFirebaseUID_EmptyUID(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.GetUserByFirebaseUID(ctx, "")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
}

func TestGetUserByFirebaseUID_NotFound(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()
	mockRepo.On("GetByFirebaseUID", ctx, "fb-999").Return(nil, nil)

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.GetUserByFirebaseUID(ctx, "fb-999")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeNotFound, appErr.Type)
}

// --- UpdateProfile Tests ---

func TestUpdateProfile_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()
	existingUser := &domain.User{ID: "usr-1", Name: "Old Name", DefaultCurrency: "INR"}
	mockRepo.On("GetByID", ctx, "usr-1").Return(existingUser, nil)
	mockRepo.On("UpdateUser", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	uc := domain.NewUseCase(mockRepo)
	u, err := uc.UpdateProfile(ctx, "usr-1", "New Name", "USD")
	require.NoError(t, err)
	assert.Equal(t, "New Name", u.Name)
	assert.Equal(t, "USD", u.DefaultCurrency)
	mockRepo.AssertExpectations(t)
}

func TestUpdateProfile_InvalidCurrencyLength(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()
	existingUser := &domain.User{ID: "usr-1", Name: "Alice", DefaultCurrency: "INR"}
	mockRepo.On("GetByID", ctx, "usr-1").Return(existingUser, nil)

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.UpdateProfile(ctx, "usr-1", "Alice", "US")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgInvalidCurrency)
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()
	mockRepo.On("GetByID", ctx, "non-existent").Return(nil, nil)

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.UpdateProfile(ctx, "non-existent", "New Name", "USD")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeNotFound, appErr.Type)
}

func TestUpdateProfile_RepoUpdateError(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()
	existingUser := &domain.User{ID: "usr-1", Name: "Alice", DefaultCurrency: "INR"}
	mockRepo.On("GetByID", ctx, "usr-1").Return(existingUser, nil)
	mockRepo.On("UpdateUser", ctx, mock.AnythingOfType("*domain.User")).Return(errors.New("db error"))

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.UpdateProfile(ctx, "usr-1", "New Name", "USD")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeInternal, appErr.Type)
}

// --- AddFriendByEmailOrPhone Tests ---

func TestAddFriendByEmailOrPhone_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	friendUser := &domain.User{ID: "usr-2", Name: "Bob"}
	mockRepo.On("GetByEmailOrPhone", ctx, "bob@example.com", "").Return(friendUser, nil)
	mockRepo.On("GetFriendship", ctx, "usr-1", "usr-2").Return(false, nil)
	mockRepo.On("CreateFriendship", ctx, "usr-1", "usr-2").Return(nil)

	uc := domain.NewUseCase(mockRepo)
	friend, err := uc.AddFriendByEmailOrPhone(ctx, "usr-1", "bob@example.com", "")
	require.NoError(t, err)
	assert.Equal(t, friendUser, friend)
	mockRepo.AssertExpectations(t)
}

func TestAddFriendByEmailOrPhone_AlreadyFriends(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	friendUser := &domain.User{ID: "usr-2", Name: "Bob"}
	mockRepo.On("GetByEmailOrPhone", ctx, "bob@example.com", "").Return(friendUser, nil)
	mockRepo.On("GetFriendship", ctx, "usr-1", "usr-2").Return(true, nil)

	uc := domain.NewUseCase(mockRepo)
	friend, err := uc.AddFriendByEmailOrPhone(ctx, "usr-1", "bob@example.com", "")
	require.NoError(t, err)
	assert.Equal(t, friendUser, friend)
	mockRepo.AssertExpectations(t)
}

func TestAddFriendByEmailOrPhone_MissingParams(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.AddFriendByEmailOrPhone(ctx, "usr-1", "", "")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
}

func TestAddFriendByEmailOrPhone_FriendNotFound(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	mockRepo.On("GetByEmailOrPhone", ctx, "unknown@example.com", "").Return(nil, nil)

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.AddFriendByEmailOrPhone(ctx, "usr-1", "unknown@example.com", "")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeNotFound, appErr.Type)
}

func TestAddFriendByEmailOrPhone_AddSelfError(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	selfUser := &domain.User{ID: "usr-1", Name: "Alice"}
	mockRepo.On("GetByEmailOrPhone", ctx, "alice@example.com", "").Return(selfUser, nil)

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.AddFriendByEmailOrPhone(ctx, "usr-1", "alice@example.com", "")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgSelfFriendError)
}

// --- RemoveFriend Tests ---

func TestRemoveFriend_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	mockRepo.On("GetFriendship", ctx, "usr-1", "usr-2").Return(true, nil)
	mockRepo.On("DeleteFriendship", ctx, "usr-1", "usr-2").Return(nil)

	uc := domain.NewUseCase(mockRepo)
	err := uc.RemoveFriend(ctx, "usr-1", "usr-2")
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRemoveFriend_EmptyFriendID(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	uc := domain.NewUseCase(mockRepo)
	err := uc.RemoveFriend(ctx, "usr-1", "")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
}

func TestRemoveFriend_NotFriends(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	mockRepo.On("GetFriendship", ctx, "usr-1", "usr-2").Return(false, nil)

	uc := domain.NewUseCase(mockRepo)
	err := uc.RemoveFriend(ctx, "usr-1", "usr-2")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgNotFriends)
}

// --- ListFriends Tests ---

func TestListFriends_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	friendsList := []domain.User{
		{ID: "usr-2", Name: "Bob", CreatedAt: time.Now()},
	}
	mockRepo.On("ListFriends", ctx, "usr-1", int32(21), (*time.Time)(nil), (*string)(nil)).Return(friendsList, nil)

	uc := domain.NewUseCase(mockRepo)
	resp, err := uc.ListFriends(ctx, "usr-1", pagination.Params{Limit: 20})
	require.NoError(t, err)
	assert.Equal(t, 1, len(resp.Data))
	assert.Equal(t, "usr-2", resp.Data[0].ID)
	mockRepo.AssertExpectations(t)
}

func TestListFriends_EmptyUserID(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.ListFriends(ctx, "", pagination.Params{Limit: 20})
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
}

func TestListFriends_RepoError(t *testing.T) {
	mockRepo := new(mockUserRepository)
	ctx := context.Background()

	mockRepo.On("ListFriends", ctx, "usr-1", int32(21), (*time.Time)(nil), (*string)(nil)).Return(nil, errors.New("db error"))

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.ListFriends(ctx, "usr-1", pagination.Params{Limit: 20})
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeInternal, appErr.Type)
}
