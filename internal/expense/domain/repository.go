package domain

import (
	"context"
	"time"
)

// Repository defines the storage contract for the expense domain.
type Repository interface {
	CreateExpense(ctx context.Context, e *Expense) error
	CreateExpenseSplit(ctx context.Context, s *Split) error
	GetExpenseByID(ctx context.Context, id string) (*Expense, error)
	ListExpenseSplits(ctx context.Context, expenseID string) ([]Split, error)
	ListExpenseSplitsByIDs(ctx context.Context, expenseIDs []string) ([]Split, error)
	ListExpensesByGroup(ctx context.Context, groupID string, limit int32, lastTime *time.Time, lastID *string) ([]Expense, error)
	ListUserPersonalExpenses(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]Expense, error)
	ListUserFriendExpenses(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]Expense, error)
	DeleteExpense(ctx context.Context, id string) error
	GetGroupBalances(ctx context.Context, groupID string) ([]UserBalance, error)
	GetFriendBalances(ctx context.Context, userID string) ([]UserBalance, error)
	GetGroupPairwiseDebts(ctx context.Context, groupID string) ([]PairwiseDebt, error)
}
