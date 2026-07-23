package user_test

import (
	"context"
	"errors"
	"testing"

	"time"

	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*user.User, error) {
	args := m.Called(ctx, firebaseUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepository) Create(ctx context.Context, u *user.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepository) UpdateUser(ctx context.Context, u *user.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepository) GetByEmailOrPhone(ctx context.Context, email, phone string) (*user.User, error) {
	args := m.Called(ctx, email, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
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

func (m *mockUserRepository) ListFriends(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]user.User, error) {
	// Not needed for simple test
	return nil, nil
}

func TestGetUserProfile_Error(t *testing.T) {
	mockRepo := new(mockUserRepository)
	mockRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, errors.New("not found"))

	uc := user.NewUseCase(mockRepo)
	_, err := uc.GetUserProfile(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve user profile")
	mockRepo.AssertExpectations(t)
}
