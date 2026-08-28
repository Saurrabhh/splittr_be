package domain

import (
	expensedomain "github.com/Saurrabhh/splittr_be/internal/expense/domain"
	groupdomain "github.com/Saurrabhh/splittr_be/internal/group/domain"
	userdomain "github.com/Saurrabhh/splittr_be/internal/user/domain"
)

type SyncParams struct {
	FriendsVersion  int64
	GroupsVersion   int64
	ExpensesVersion int64
	Limit           int32
}

type SyncResponse struct {
	Friends  userdomain.FriendSyncResponse     `json:"friends"`
	Groups   groupdomain.GroupSyncResponse     `json:"groups"`
	Expenses expensedomain.ExpenseSyncResponse `json:"expenses"`
} // @name Sync.SyncResponse
