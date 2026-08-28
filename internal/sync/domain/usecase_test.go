package domain_test

import (
	"context"
	"errors"
	"testing"

	expensedomain "github.com/Saurrabhh/splittr_be/internal/expense/domain"
	groupdomain "github.com/Saurrabhh/splittr_be/internal/group/domain"
	syncdomain "github.com/Saurrabhh/splittr_be/internal/sync/domain"
	userdomain "github.com/Saurrabhh/splittr_be/internal/user/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserSyncService struct {
	mock.Mock
}

func (m *mockUserSyncService) SyncFriends(ctx context.Context, lastVersion int64, userID string, limit int32) (*userdomain.FriendSyncResponse, error) {
	args := m.Called(ctx, lastVersion, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userdomain.FriendSyncResponse), args.Error(1)
}

type mockGroupSyncService struct {
	mock.Mock
}

func (m *mockGroupSyncService) SyncGroups(ctx context.Context, lastVersion int64, userID string, limit int32) (*groupdomain.GroupSyncResponse, error) {
	args := m.Called(ctx, lastVersion, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*groupdomain.GroupSyncResponse), args.Error(1)
}

type mockExpenseSyncService struct {
	mock.Mock
}

func (m *mockExpenseSyncService) SyncExpenses(ctx context.Context, lastVersion int64, userID string, limit int32) (*expensedomain.ExpenseSyncResponse, error) {
	args := m.Called(ctx, lastVersion, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*expensedomain.ExpenseSyncResponse), args.Error(1)
}

func TestSync_Success(t *testing.T) {
	mockUser := new(mockUserSyncService)
	mockGroup := new(mockGroupSyncService)
	mockExpense := new(mockExpenseSyncService)
	ctx := context.Background()
	userID := "usr-1"

	mockUser.On("SyncFriends", ctx, int64(10), userID, int32(100)).Return(&userdomain.FriendSyncResponse{
		NewVersion: 15,
		Updated: []userdomain.FriendshipSyncRecord{
			{UserID: "usr-1", FriendID: "usr-2", Status: userdomain.Accepted, SyncVersion: 15},
		},
		DeletedIDs: []string{"usr-3"},
	}, nil)

	mockGroup.On("SyncGroups", ctx, int64(20), userID, int32(100)).Return(&groupdomain.GroupSyncResponse{
		NewVersion: 25,
		Updated: []groupdomain.Group{
			{ID: "grp-1", Name: "Trip", SyncVersion: 25},
		},
		DeletedIDs: []string{"grp-2"},
	}, nil)

	mockExpense.On("SyncExpenses", ctx, int64(30), userID, int32(100)).Return(&expensedomain.ExpenseSyncResponse{
		NewVersion: 35,
		Updated: []expensedomain.ExpenseWithSplits{
			{Expense: expensedomain.Expense{ID: "exp-1", Amount: 50.0, SyncVersion: 35}},
		},
		DeletedIDs: []string{"exp-2"},
	}, nil)

	uc := syncdomain.NewUseCase(mockUser, mockGroup, mockExpense)
	params := syncdomain.SyncParams{
		FriendsVersion:  10,
		GroupsVersion:   20,
		ExpensesVersion: 30,
		Limit:           100,
	}

	resp, err := uc.Sync(ctx, userID, params)
	require.NoError(t, err)
	assert.Equal(t, int64(15), resp.Friends.NewVersion)
	assert.Len(t, resp.Friends.Updated, 1)
	assert.Equal(t, []string{"usr-3"}, resp.Friends.DeletedIDs)

	assert.Equal(t, int64(25), resp.Groups.NewVersion)
	assert.Len(t, resp.Groups.Updated, 1)
	assert.Equal(t, []string{"grp-2"}, resp.Groups.DeletedIDs)

	assert.Equal(t, int64(35), resp.Expenses.NewVersion)
	assert.Len(t, resp.Expenses.Updated, 1)
	assert.Equal(t, []string{"exp-2"}, resp.Expenses.DeletedIDs)

	mockUser.AssertExpectations(t)
	mockGroup.AssertExpectations(t)
	mockExpense.AssertExpectations(t)
}

func TestSync_UserSyncError(t *testing.T) {
	mockUser := new(mockUserSyncService)
	mockGroup := new(mockGroupSyncService)
	mockExpense := new(mockExpenseSyncService)
	ctx := context.Background()
	userID := "usr-1"

	mockUser.On("SyncFriends", ctx, int64(10), userID, int32(100)).Return(nil, errors.New("user sync error"))

	uc := syncdomain.NewUseCase(mockUser, mockGroup, mockExpense)
	params := syncdomain.SyncParams{
		FriendsVersion: 10,
	}

	_, err := uc.Sync(ctx, userID, params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user sync error")
}
