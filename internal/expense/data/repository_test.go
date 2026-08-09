//go:build integration

package data_test

import (
	"context"
	"testing"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db_test"
	"github.com/Saurrabhh/splittr_be/internal/expense/data"
	"github.com/Saurrabhh/splittr_be/internal/expense/domain"
	"github.com/Saurrabhh/splittr_be/internal/group"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) (*data.DBRepository, *user.DBRepository, *group.DBRepository, func()) {
	ctx := context.Background()
	testDB, cleanup, err := db_test.SetupTestDB(ctx)
	require.NoError(t, err)

	tm := db.NewTransactionManager(testDB)
	expenseRepo := data.NewRepository(testDB, tm)
	userRepo := user.NewRepository(testDB, tm)
	groupRepo := group.NewRepository(testDB, tm)
	return expenseRepo, userRepo, groupRepo, cleanup
}

func createTestUser(t *testing.T, userRepo *user.DBRepository, name string) *user.User {
	ctx := context.Background()
	email := uuid.New().String() + "@example.com"
	u := &user.User{
		ID:          uuid.New().String(),
		FirebaseUID: "fb-" + uuid.New().String(),
		Email:       &email,
		Name:        name,
	}
	err := userRepo.Create(ctx, u)
	require.NoError(t, err)
	return u
}

func createTestGroup(t *testing.T, groupRepo *group.DBRepository, creatorID string, name string) *group.Group {
	ctx := context.Background()
	g := &group.Group{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedBy: &creatorID,
	}
	err := groupRepo.CreateGroup(ctx, g)
	require.NoError(t, err)
	return g
}

func TestRepository_CreateExpense_And_GetByID(t *testing.T) {
	expenseRepo, userRepo, _, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	payer := createTestUser(t, userRepo, "Payer Alice")
	creator := createTestUser(t, userRepo, "Creator Bob")
	expID := uuid.New().String()

	e := &domain.Expense{
		ID:          expID,
		Description: "Dinner at Restaurant",
		Amount:      120.50,
		Currency:    "USD",
		Category:    "Food",
		PaidBy:      payer.ID,
		CreatedBy:   creator.ID,
		IsPayment:   false,
	}

	err := expenseRepo.CreateExpense(ctx, e)
	require.NoError(t, err)
	assert.False(t, e.CreatedAt.IsZero())
	assert.False(t, e.UpdatedAt.IsZero())

	// GetExpenseByID Success
	fetched, err := expenseRepo.GetExpenseByID(ctx, expID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, expID, fetched.ID)
	assert.Equal(t, "Dinner at Restaurant", fetched.Description)
	assert.Equal(t, 120.50, fetched.Amount)
	assert.Equal(t, "USD", fetched.Currency)
	assert.Equal(t, "", fetched.Category)
	assert.Equal(t, payer.ID, fetched.PaidBy)
	assert.Equal(t, creator.ID, fetched.CreatedBy)
	assert.False(t, fetched.IsPayment)

	// GetExpenseByID Invalid UUID
	_, err = expenseRepo.GetExpenseByID(ctx, "invalid-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid uuid")

	// GetExpenseByID Not Found
	notFound, err := expenseRepo.GetExpenseByID(ctx, uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestRepository_CreateExpenseSplit_And_ListSplits(t *testing.T) {
	expenseRepo, userRepo, _, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	u1 := createTestUser(t, userRepo, "User 1")
	u2 := createTestUser(t, userRepo, "User 2")
	expID := uuid.New().String()

	e := &domain.Expense{
		ID:          expID,
		Description: "Movie Tickets",
		Amount:      50.00,
		Currency:    "INR",
		Category:    "Entertainment",
		PaidBy:      u1.ID,
		CreatedBy:   u1.ID,
	}
	err := expenseRepo.CreateExpense(ctx, e)
	require.NoError(t, err)

	splitVal := 50.0
	s1 := &domain.Split{
		ExpenseID:  expID,
		UserID:     u1.ID,
		Amount:     25.00,
		SplitType:  domain.SplitTypeEqual,
		SplitValue: &splitVal,
	}
	s2 := &domain.Split{
		ExpenseID:  expID,
		UserID:     u2.ID,
		Amount:     25.00,
		SplitType:  domain.SplitTypeEqual,
		SplitValue: &splitVal,
	}

	err = expenseRepo.CreateExpenseSplit(ctx, s1)
	require.NoError(t, err)

	err = expenseRepo.CreateExpenseSplit(ctx, s2)
	require.NoError(t, err)

	splits, err := expenseRepo.ListExpenseSplits(ctx, expID)
	require.NoError(t, err)
	assert.Len(t, splits, 2)

	userIDs := map[string]float64{
		splits[0].UserID: splits[0].Amount,
		splits[1].UserID: splits[1].Amount,
	}
	assert.Equal(t, 25.00, userIDs[u1.ID])
	assert.Equal(t, 25.00, userIDs[u2.ID])
}

func TestRepository_ListExpensesByGroup(t *testing.T) {
	expenseRepo, userRepo, groupRepo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	u := createTestUser(t, userRepo, "Group Admin")
	g := createTestGroup(t, groupRepo, u.ID, "Vacation Group")

	exp1 := &domain.Expense{
		ID:          uuid.New().String(),
		Description: "Flight",
		Amount:      500.00,
		Currency:    "USD",
		Category:    "Travel",
		GroupID:     &g.ID,
		PaidBy:      u.ID,
		CreatedBy:   u.ID,
	}
	exp2 := &domain.Expense{
		ID:          uuid.New().String(),
		Description: "Hotel",
		Amount:      300.00,
		Currency:    "USD",
		Category:    "Travel",
		GroupID:     &g.ID,
		PaidBy:      u.ID,
		CreatedBy:   u.ID,
	}

	require.NoError(t, expenseRepo.CreateExpense(ctx, exp1))
	require.NoError(t, expenseRepo.CreateExpense(ctx, exp2))

	expenses, err := expenseRepo.ListExpensesByGroup(ctx, g.ID, 10, nil, nil)
	require.NoError(t, err)
	assert.Len(t, expenses, 2)
}

func TestRepository_ListUserPersonalExpenses(t *testing.T) {
	expenseRepo, userRepo, _, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	u := createTestUser(t, userRepo, "Solo Spender")
	expID := uuid.New().String()

	e := &domain.Expense{
		ID:          expID,
		Description: "Coffee",
		Amount:      5.00,
		Currency:    "USD",
		Category:    "Food",
		GroupID:     nil,
		PaidBy:      u.ID,
		CreatedBy:   u.ID,
	}
	require.NoError(t, expenseRepo.CreateExpense(ctx, e))

	s := &domain.Split{
		ExpenseID: expID,
		UserID:    u.ID,
		Amount:    5.00,
		SplitType: domain.SplitTypeExact,
	}
	require.NoError(t, expenseRepo.CreateExpenseSplit(ctx, s))

	personal, err := expenseRepo.ListUserPersonalExpenses(ctx, u.ID, 10, nil, nil)
	require.NoError(t, err)
	assert.Len(t, personal, 1)
	assert.Equal(t, "Coffee", personal[0].Description)
}

func TestRepository_ListUserFriendExpenses(t *testing.T) {
	expenseRepo, userRepo, _, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	u1 := createTestUser(t, userRepo, "Friend 1")
	u2 := createTestUser(t, userRepo, "Friend 2")
	expID := uuid.New().String()

	e := &domain.Expense{
		ID:          expID,
		Description: "Shared Cab",
		Amount:      40.00,
		Currency:    "USD",
		Category:    "Transport",
		GroupID:     nil,
		PaidBy:      u1.ID,
		CreatedBy:   u1.ID,
	}
	require.NoError(t, expenseRepo.CreateExpense(ctx, e))

	s1 := &domain.Split{ExpenseID: expID, UserID: u1.ID, Amount: 20.00, SplitType: domain.SplitTypeEqual}
	s2 := &domain.Split{ExpenseID: expID, UserID: u2.ID, Amount: 20.00, SplitType: domain.SplitTypeEqual}
	require.NoError(t, expenseRepo.CreateExpenseSplit(ctx, s1))
	require.NoError(t, expenseRepo.CreateExpenseSplit(ctx, s2))

	friendExp, err := expenseRepo.ListUserFriendExpenses(ctx, u1.ID, 10, nil, nil)
	require.NoError(t, err)
	assert.Len(t, friendExp, 1)
	assert.Equal(t, "Shared Cab", friendExp[0].Description)
}

func TestRepository_DeleteExpense(t *testing.T) {
	expenseRepo, userRepo, _, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	u := createTestUser(t, userRepo, "User Deleter")
	expID := uuid.New().String()

	e := &domain.Expense{
		ID:          expID,
		Description: "To be deleted",
		Amount:      15.00,
		Currency:    "INR",
		Category:    "Other",
		PaidBy:      u.ID,
		CreatedBy:   u.ID,
	}
	require.NoError(t, expenseRepo.CreateExpense(ctx, e))

	err := expenseRepo.DeleteExpense(ctx, expID)
	require.NoError(t, err)

	// Fetching deleted expense should return nil (soft deleted)
	fetched, err := expenseRepo.GetExpenseByID(ctx, expID)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestRepository_GetGroupBalances_And_Pairwise(t *testing.T) {
	expenseRepo, userRepo, groupRepo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	u1 := createTestUser(t, userRepo, "Alice")
	u2 := createTestUser(t, userRepo, "Bob")
	g := createTestGroup(t, groupRepo, u1.ID, "Balance Group")

	_, err = groupRepo.AddGroupMembers(ctx, g.ID, []string{u1.ID}, group.MemberRoleAdmin, group.MemberStatusActive)
	require.NoError(t, err)
	_, err = groupRepo.AddGroupMembers(ctx, g.ID, []string{u2.ID}, group.MemberRoleMember, group.MemberStatusActive)
	require.NoError(t, err)

	expID := uuid.New().String()
	e := &domain.Expense{
		ID:          expID,
		Description: "Group Lunch",
		Amount:      100.00,
		Currency:    "USD",
		Category:    "Food",
		GroupID:     &g.ID,
		PaidBy:      u1.ID,
		CreatedBy:   u1.ID,
	}
	require.NoError(t, expenseRepo.CreateExpense(ctx, e))

	s1 := &domain.Split{ExpenseID: expID, UserID: u1.ID, Amount: 50.00, SplitType: domain.SplitTypeEqual}
	s2 := &domain.Split{ExpenseID: expID, UserID: u2.ID, Amount: 50.00, SplitType: domain.SplitTypeEqual}
	require.NoError(t, expenseRepo.CreateExpenseSplit(ctx, s1))
	require.NoError(t, expenseRepo.CreateExpenseSplit(ctx, s2))

	balances, err := expenseRepo.GetGroupBalances(ctx, g.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, balances)

	pairwise, err := expenseRepo.GetGroupPairwiseDebts(ctx, g.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, pairwise)
}

func TestRepository_GetFriendBalances(t *testing.T) {
	expenseRepo, userRepo, _, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	u1 := createTestUser(t, userRepo, "Friend Alice")
	u2 := createTestUser(t, userRepo, "Friend Bob")

	expID := uuid.New().String()
	e := &domain.Expense{
		ID:          expID,
		Description: "Dinner",
		Amount:      100.00,
		Currency:    "USD",
		Category:    "Food",
		GroupID:     nil,
		PaidBy:      u1.ID,
		CreatedBy:   u1.ID,
	}
	require.NoError(t, expenseRepo.CreateExpense(ctx, e))

	s1 := &domain.Split{ExpenseID: expID, UserID: u1.ID, Amount: 50.00, SplitType: domain.SplitTypeEqual}
	s2 := &domain.Split{ExpenseID: expID, UserID: u2.ID, Amount: 50.00, SplitType: domain.SplitTypeEqual}
	require.NoError(t, expenseRepo.CreateExpenseSplit(ctx, s1))
	require.NoError(t, expenseRepo.CreateExpenseSplit(ctx, s2))

	balances, err := expenseRepo.GetFriendBalances(ctx, u1.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, balances)
}
