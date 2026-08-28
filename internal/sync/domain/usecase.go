package domain

import (
	"context"

	expensedomain "github.com/Saurrabhh/splittr_be/internal/expense/domain"
	groupdomain "github.com/Saurrabhh/splittr_be/internal/group/domain"
	userdomain "github.com/Saurrabhh/splittr_be/internal/user/domain"
)

type UserSyncService interface {
	SyncFriends(ctx context.Context, lastVersion int64, userID string, limit int32) (*userdomain.FriendSyncResponse, error)
}

type GroupSyncService interface {
	SyncGroups(ctx context.Context, lastVersion int64, userID string, limit int32) (*groupdomain.GroupSyncResponse, error)
}

type ExpenseSyncService interface {
	SyncExpenses(ctx context.Context, lastVersion int64, userID string, limit int32) (*expensedomain.ExpenseSyncResponse, error)
}

type UseCase struct {
	userSync    UserSyncService
	groupSync   GroupSyncService
	expenseSync ExpenseSyncService
}

func NewUseCase(userSync UserSyncService, groupSync GroupSyncService, expenseSync ExpenseSyncService) *UseCase {
	return &UseCase{
		userSync:    userSync,
		groupSync:   groupSync,
		expenseSync: expenseSync,
	}
}

func (u *UseCase) Sync(ctx context.Context, userID string, p SyncParams) (*SyncResponse, error) {
	limit := p.Limit
	if limit <= 0 {
		limit = 100
	}

	friends, err := u.userSync.SyncFriends(ctx, p.FriendsVersion, userID, limit)
	if err != nil {
		return nil, err
	}

	groups, err := u.groupSync.SyncGroups(ctx, p.GroupsVersion, userID, limit)
	if err != nil {
		return nil, err
	}

	expenses, err := u.expenseSync.SyncExpenses(ctx, p.ExpensesVersion, userID, limit)
	if err != nil {
		return nil, err
	}

	return &SyncResponse{
		Friends:  *friends,
		Groups:   *groups,
		Expenses: *expenses,
	}, nil
}
