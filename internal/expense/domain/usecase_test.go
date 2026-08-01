package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/expense/domain"
	"github.com/Saurrabhh/splittr_be/internal/group"
	"github.com/Saurrabhh/splittr_be/internal/notification"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockExpenseRepository struct {
	mock.Mock
}

func (m *mockExpenseRepository) CreateExpense(ctx context.Context, e *domain.Expense) error {
	return m.Called(ctx, e).Error(0)
}

func (m *mockExpenseRepository) CreateExpenseSplit(ctx context.Context, s *domain.Split) error {
	return m.Called(ctx, s).Error(0)
}

func (m *mockExpenseRepository) GetExpenseByID(ctx context.Context, id string) (*domain.Expense, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Expense), args.Error(1)
}

func (m *mockExpenseRepository) ListExpenseSplits(ctx context.Context, expenseID string) ([]domain.Split, error) {
	args := m.Called(ctx, expenseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Split), args.Error(1)
}

func (m *mockExpenseRepository) ListExpensesByGroup(ctx context.Context, groupID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.Expense, error) {
	args := m.Called(ctx, groupID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Expense), args.Error(1)
}

func (m *mockExpenseRepository) ListUserPersonalExpenses(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.Expense, error) {
	args := m.Called(ctx, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Expense), args.Error(1)
}

func (m *mockExpenseRepository) ListUserFriendExpenses(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.Expense, error) {
	args := m.Called(ctx, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Expense), args.Error(1)
}

func (m *mockExpenseRepository) DeleteExpense(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockExpenseRepository) GetGroupBalances(ctx context.Context, groupID string) ([]domain.UserBalance, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserBalance), args.Error(1)
}

func (m *mockExpenseRepository) GetFriendBalances(ctx context.Context, userID string) ([]domain.UserBalance, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserBalance), args.Error(1)
}

func (m *mockExpenseRepository) GetGroupPairwiseDebts(ctx context.Context, groupID string) ([]domain.PairwiseDebt, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PairwiseDebt), args.Error(1)
}

func (m *mockExpenseRepository) ListExpenseSplitsByIDs(ctx context.Context, expenseIDs []string) ([]domain.Split, error) {
	args := m.Called(ctx, expenseIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Split), args.Error(1)
}

type mockGroupService struct {
	mock.Mock
}

func (m *mockGroupService) GetGroupDetails(ctx context.Context, groupID, userID string) (*group.Group, []group.Member, error) {
	args := m.Called(ctx, groupID, userID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var members []group.Member
	if args.Get(1) != nil {
		members = args.Get(1).([]group.Member)
	}
	return args.Get(0).(*group.Group), members, args.Error(2)
}

type mockActivityLogger struct {
	mock.Mock
}

func (m *mockActivityLogger) LogEvent(
	ctx context.Context,
	actorID string,
	groupID *string,
	visibleToUserIDs []string,
	event activity.Event,
) (*activity.Activity, error) {
	args := m.Called(ctx, actorID, groupID, visibleToUserIDs, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*activity.Activity), args.Error(1)
}

type mockNotificationSender struct {
	mock.Mock
}

func (m *mockNotificationSender) CreateAlert(ctx context.Context, userID string, actorID *string, activityID *string, alert notification.Alert) error {
	args := m.Called(ctx, userID, actorID, activityID, alert)
	return args.Error(0)
}

type mockTransactor struct {
	fail bool
}

func (m *mockTransactor) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.fail {
		return errors.New("transaction error")
	}
	return fn(ctx)
}

// --- CreateExpense Tests ---

func TestCreateExpense_Success_EqualSplit(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	mockGroupSvc := new(mockGroupService)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	creatorID := "usr-1"
	paidBy := "usr-1"
	inputs := []domain.InputSplit{
		{UserID: "usr-1"},
		{UserID: "usr-2"},
	}

	mockRepo.On("CreateExpense", ctx, mock.AnythingOfType("*domain.Expense")).Return(nil)
	mockRepo.On("CreateExpenseSplit", ctx, mock.AnythingOfType("*domain.Split")).Return(nil)
	mockRepo.On("ListExpenseSplits", ctx, mock.AnythingOfType("string")).Return([]domain.Split{
		{ExpenseID: "exp-1", UserID: "usr-1", Amount: 50.0},
		{ExpenseID: "exp-1", UserID: "usr-2", Amount: 50.0},
	}, nil)

	mockAct.On("LogEvent",
		ctx, creatorID, (*string)(nil), mock.Anything, mock.Anything,
	).Return(&activity.Activity{ID: "act-1"}, nil)

	mockNotif.On("CreateAlert", ctx, "usr-2", &creatorID, mock.Anything, mock.Anything).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockGroupSvc, mockAct, mockNotif)
	exp, splits, err := uc.CreateExpense(ctx, "Dinner", 100.0, "INR", "Food", nil, paidBy, domain.SplitTypeEqual, inputs, creatorID)

	require.NoError(t, err)
	assert.NotNil(t, exp)
	assert.Equal(t, "Dinner", exp.Description)
	assert.Equal(t, 100.0, exp.Amount)
	assert.Len(t, splits, 2)

	mockRepo.AssertExpectations(t)
	mockAct.AssertExpectations(t)
	mockNotif.AssertExpectations(t)
}

func TestCreateExpense_Success_GroupExactSplit(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	mockGroupSvc := new(mockGroupService)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	creatorID := "usr-1"
	paidBy := "usr-1"
	amt1 := 60.0
	amt2 := 40.0
	inputs := []domain.InputSplit{
		{UserID: "usr-1", Amount: &amt1},
		{UserID: "usr-2", Amount: &amt2},
	}

	members := []group.Member{
		{GroupID: groupID, UserID: "usr-1", Role: group.MemberRoleAdmin},
		{GroupID: groupID, UserID: "usr-2", Role: group.MemberRoleMember},
	}

	mockGroupSvc.On("GetGroupDetails", ctx, groupID, creatorID).Return(&group.Group{ID: groupID}, members, nil)
	mockGroupSvc.On("GetGroupDetails", mock.Anything, groupID, creatorID).Return(&group.Group{ID: groupID}, members, nil)

	mockRepo.On("CreateExpense", ctx, mock.AnythingOfType("*domain.Expense")).Return(nil)
	mockRepo.On("CreateExpenseSplit", ctx, mock.AnythingOfType("*domain.Split")).Return(nil)
	mockRepo.On("ListExpenseSplits", ctx, mock.AnythingOfType("string")).Return([]domain.Split{
		{ExpenseID: "exp-1", UserID: "usr-1", Amount: 60.0},
		{ExpenseID: "exp-1", UserID: "usr-2", Amount: 40.0},
	}, nil)

	mockAct.On("LogEvent",
		ctx, creatorID, &groupID, ([]string)(nil), mock.Anything,
	).Return(&activity.Activity{ID: "act-1"}, nil)

	mockNotif.On("CreateAlert", ctx, "usr-2", &creatorID, mock.Anything, mock.Anything).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockGroupSvc, mockAct, mockNotif)
	exp, splits, err := uc.CreateExpense(ctx, "Hotel", 100.0, "INR", "Travel", &groupID, paidBy, domain.SplitTypeExact, inputs, creatorID)

	require.NoError(t, err)
	assert.NotNil(t, exp)
	assert.Equal(t, &groupID, exp.GroupID)
	assert.Len(t, splits, 2)
}

func TestCreateExpense_Success_PercentageSplit(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	mockGroupSvc := new(mockGroupService)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	creatorID := "usr-1"
	paidBy := "usr-1"
	p1 := 70.0
	p2 := 30.0
	inputs := []domain.InputSplit{
		{UserID: "usr-1", Percentage: &p1},
		{UserID: "usr-2", Percentage: &p2},
	}

	mockRepo.On("CreateExpense", ctx, mock.AnythingOfType("*domain.Expense")).Return(nil)
	mockRepo.On("CreateExpenseSplit", ctx, mock.AnythingOfType("*domain.Split")).Return(nil)
	mockRepo.On("ListExpenseSplits", ctx, mock.AnythingOfType("string")).Return([]domain.Split{
		{ExpenseID: "exp-1", UserID: "usr-1", Amount: 140.0},
		{ExpenseID: "exp-1", UserID: "usr-2", Amount: 60.0},
	}, nil)

	mockAct.On("LogEvent",
		ctx, creatorID, (*string)(nil), mock.Anything, mock.Anything,
	).Return(&activity.Activity{ID: "act-1"}, nil)

	mockNotif.On("CreateAlert", ctx, "usr-2", &creatorID, mock.Anything, mock.Anything).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockGroupSvc, mockAct, mockNotif)
	exp, splits, err := uc.CreateExpense(ctx, "Party", 200.0, "USD", "Entertainment", nil, paidBy, domain.SplitTypePercentage, inputs, creatorID)

	require.NoError(t, err)
	assert.NotNil(t, exp)
	assert.Len(t, splits, 2)
}

func TestCreateExpense_ValidationErrors(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	mockGroupSvc := new(mockGroupService)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	uc := domain.NewUseCase(mockRepo, mockTx, mockGroupSvc, mockAct, mockNotif)

	// Empty description
	_, _, err := uc.CreateExpense(ctx, "", 100.0, "INR", "Cat", nil, "usr-1", domain.SplitTypeEqual, []domain.InputSplit{{UserID: "usr-1"}}, "usr-1")
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "description is required")

	// Invalid total amount <= 0
	_, _, err = uc.CreateExpense(ctx, "Desc", 0.0, "INR", "Cat", nil, "usr-1", domain.SplitTypeEqual, []domain.InputSplit{{UserID: "usr-1"}}, "usr-1")
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgInvalidAmount)

	// Empty splits inputs
	_, _, err = uc.CreateExpense(ctx, "Desc", 100.0, "INR", "Cat", nil, "usr-1", domain.SplitTypeEqual, nil, "usr-1")
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgInvalidSplit)

	// Exact split sum mismatch (total 100, splits 50+40=90)
	amt1 := 50.0
	amt2 := 40.0
	inputsMismatch := []domain.InputSplit{
		{UserID: "usr-1", Amount: &amt1},
		{UserID: "usr-2", Amount: &amt2},
	}
	_, _, err = uc.CreateExpense(ctx, "Desc", 100.0, "INR", "Cat", nil, "usr-1", domain.SplitTypeExact, inputsMismatch, "usr-1")
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "does not match total expense amount")

	// Exact split missing amount
	inputsMissingAmt := []domain.InputSplit{
		{UserID: "usr-1"},
	}
	_, _, err = uc.CreateExpense(ctx, "Desc", 100.0, "INR", "Cat", nil, "usr-1", domain.SplitTypeExact, inputsMissingAmt, "usr-1")
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "amount is required for each user")

	// Percentage split sum != 100
	p1 := 50.0
	p2 := 40.0
	inputsPctMismatch := []domain.InputSplit{
		{UserID: "usr-1", Percentage: &p1},
		{UserID: "usr-2", Percentage: &p2},
	}
	_, _, err = uc.CreateExpense(ctx, "Desc", 100.0, "INR", "Cat", nil, "usr-1", domain.SplitTypePercentage, inputsPctMismatch, "usr-1")
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "must equal 100%")

	// Non-member payer rejection in group
	groupID := "grp-1"
	members := []group.Member{
		{GroupID: groupID, UserID: "usr-2", Role: group.MemberRoleMember},
	}
	mockGroupSvc.On("GetGroupDetails", ctx, groupID, "usr-1").Return(&group.Group{ID: groupID}, members, nil).Once()

	_, _, err = uc.CreateExpense(ctx, "Desc", 100.0, "INR", "Cat", &groupID, "usr-1", domain.SplitTypeEqual, []domain.InputSplit{{UserID: "usr-2"}}, "usr-1")
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "payer must be a member of the group")

	// Non-member split user rejection in group
	membersWithPayer := []group.Member{
		{GroupID: groupID, UserID: "usr-1", Role: group.MemberRoleAdmin},
	}
	mockGroupSvc.On("GetGroupDetails", ctx, groupID, "usr-1").Return(&group.Group{ID: groupID}, membersWithPayer, nil).Once()

	_, _, err = uc.CreateExpense(ctx, "Desc", 100.0, "INR", "Cat", &groupID, "usr-1", domain.SplitTypeEqual, []domain.InputSplit{{UserID: "usr-non-member"}}, "usr-1")
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgSplitUserNotMember)
}

func TestCreateExpense_TransactionFailure(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	mockGroupSvc := new(mockGroupService)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{fail: true}
	ctx := context.Background()

	uc := domain.NewUseCase(mockRepo, mockTx, mockGroupSvc, mockAct, mockNotif)
	_, _, err := uc.CreateExpense(ctx, "Dinner", 100.0, "INR", "Food", nil, "usr-1", domain.SplitTypeEqual, []domain.InputSplit{{UserID: "usr-1"}}, "usr-1")

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeInternal, appErr.Type)
	assert.Contains(t, appErr.Message, "create expense transaction failed")
}

func TestCreateExpense_RefetchSplitsError(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	creatorID := "usr-1"
	inputs := []domain.InputSplit{{UserID: "usr-1"}}

	mockRepo.On("CreateExpense", ctx, mock.AnythingOfType("*domain.Expense")).Return(nil)
	mockRepo.On("CreateExpenseSplit", ctx, mock.AnythingOfType("*domain.Split")).Return(nil)
	mockRepo.On("ListExpenseSplits", ctx, mock.AnythingOfType("string")).Return([]domain.Split{
		{ExpenseID: "exp-1", UserID: "usr-1", Amount: 100.0},
	}, nil).Once()
	mockRepo.On("ListExpenseSplits", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("db down")).Once()

	mockAct.On("LogEvent",
		ctx, creatorID, (*string)(nil), mock.Anything, mock.Anything,
	).Return(&activity.Activity{ID: "act-1"}, nil)

	uc := domain.NewUseCase(mockRepo, mockTx, nil, mockAct, mockNotif)
	_, _, err := uc.CreateExpense(ctx, "Dinner", 100.0, "INR", "Food", nil, creatorID, domain.SplitTypeEqual, inputs, creatorID)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeInternal, appErr.Type)
	assert.Contains(t, appErr.Message, "failed to load expense splits")
}

// --- SettleUp Tests ---

func TestSettleUp_Success(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	mockGroupSvc := new(mockGroupService)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	paidBy := "usr-1"
	receivedBy := "usr-2"
	createdBy := "usr-1"

	mockRepo.On("CreateExpense", ctx, mock.AnythingOfType("*domain.Expense")).Return(nil)
	mockRepo.On("CreateExpenseSplit", ctx, mock.AnythingOfType("*domain.Split")).Return(nil)
	mockRepo.On("ListExpenseSplits", ctx, mock.AnythingOfType("string")).Return([]domain.Split{
		{ExpenseID: "exp-1", UserID: receivedBy, Amount: 50.0},
	}, nil)

	mockAct.On("LogEvent",
		ctx, createdBy, (*string)(nil), mock.Anything, mock.Anything,
	).Return(&activity.Activity{ID: "act-1"}, nil)

	mockNotif.On("CreateAlert", ctx, receivedBy, &paidBy, mock.Anything, mock.Anything).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockGroupSvc, mockAct, mockNotif)
	exp, split, err := uc.SettleUp(ctx, 50.0, "INR", nil, paidBy, receivedBy, createdBy)

	require.NoError(t, err)
	assert.NotNil(t, exp)
	assert.True(t, exp.IsPayment)
	assert.Equal(t, 50.0, exp.Amount)
	assert.NotNil(t, split)
	assert.Equal(t, receivedBy, split.UserID)
}

func TestSettleUp_ValidationErrors(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	mockGroupSvc := new(mockGroupService)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	uc := domain.NewUseCase(mockRepo, mockTx, mockGroupSvc, nil, nil)

	// Amount <= 0
	_, _, err := uc.SettleUp(ctx, 0.0, "INR", nil, "usr-1", "usr-2", "usr-1")
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)

	// paidBy == receivedBy
	_, _, err = uc.SettleUp(ctx, 50.0, "INR", nil, "usr-1", "usr-1", "usr-1")
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgSamePayerPayee)

	// empty receivedBy
	_, _, err = uc.SettleUp(ctx, 50.0, "INR", nil, "usr-1", "", "usr-1")
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgMissingRecipient)
}

// --- GetExpenseDetails Tests ---

func TestGetExpenseDetails_Success(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	ctx := context.Background()

	expID := "exp-1"
	userID := "usr-1"
	expectedExp := &domain.Expense{ID: expID, Description: "Lunch", PaidBy: userID, CreatedBy: userID}
	expectedSplits := []domain.Split{{ExpenseID: expID, UserID: userID, Amount: 30.0}}

	mockRepo.On("GetExpenseByID", ctx, expID).Return(expectedExp, nil)
	mockRepo.On("ListExpenseSplits", ctx, expID).Return(expectedSplits, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	e, splits, err := uc.GetExpenseDetails(ctx, expID, userID)

	require.NoError(t, err)
	assert.Equal(t, expectedExp.ID, e.ID)
	assert.Len(t, splits, 1)
}

func TestGetExpenseDetails_NotFound(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	ctx := context.Background()

	expID := "non-existent"
	mockRepo.On("GetExpenseByID", ctx, expID).Return((*domain.Expense)(nil), nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	_, _, err := uc.GetExpenseDetails(ctx, expID, "usr-1")

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeNotFound, appErr.Type)
}

func TestGetExpenseDetails_Forbidden(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	ctx := context.Background()

	expID := "exp-1"
	expectedExp := &domain.Expense{ID: expID, Description: "Lunch", PaidBy: "usr-creator", CreatedBy: "usr-creator"}
	expectedSplits := []domain.Split{{ExpenseID: expID, UserID: "usr-other", Amount: 30.0}}

	mockRepo.On("GetExpenseByID", ctx, expID).Return(expectedExp, nil)
	mockRepo.On("ListExpenseSplits", ctx, expID).Return(expectedSplits, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	_, _, err := uc.GetExpenseDetails(ctx, expID, "usr-unrelated")

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeForbidden, appErr.Type)
	assert.Contains(t, appErr.Message, "access denied")
}

// --- ListExpenses Tests ---

func TestListExpenses_Group(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	mockGroupSvc := new(mockGroupService)
	ctx := context.Background()

	groupID := "grp-1"
	userID := "usr-1"
	p := pagination.Params{Limit: 10}

	mockGroupSvc.On("GetGroupDetails", ctx, groupID, userID).Return(&group.Group{ID: groupID}, []group.Member{}, nil)
	mockRepo.On("ListExpensesByGroup", ctx, groupID, int32(11), (*time.Time)(nil), (*string)(nil)).Return([]domain.Expense{
		{ID: "exp-1", Description: "Trip dinner"},
	}, nil)
	mockRepo.On("ListExpenseSplitsByIDs", ctx, []string{"exp-1"}).Return([]domain.Split{}, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, mockGroupSvc, nil, nil)
	resp, err := uc.ListExpenses(ctx, "group", groupID, userID, p)

	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "exp-1", resp.Data[0].Expense.ID)
}

func TestListExpenses_InvalidFilterType(t *testing.T) {
	uc := domain.NewUseCase(new(mockExpenseRepository), &mockTransactor{}, nil, nil, nil)
	_, err := uc.ListExpenses(context.Background(), "invalid", "id", "usr-1", pagination.Params{Limit: 10})

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
}

// --- DeleteExpense Authorization Checks (creator vs admin vs regular member) ---

func TestDeleteExpense_Creator_Success(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	ctx := context.Background()

	expID := "exp-1"
	creatorID := "usr-creator"
	e := &domain.Expense{ID: expID, CreatedBy: creatorID}

	mockRepo.On("GetExpenseByID", ctx, expID).Return(e, nil)
	mockRepo.On("DeleteExpense", ctx, expID).Return(nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	err := uc.DeleteExpense(ctx, expID, creatorID)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteExpense_NonCreator_Forbidden(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	ctx := context.Background()

	expID := "exp-1"
	creatorID := "usr-creator"
	regularMemberID := "usr-member"
	adminID := "usr-admin"
	e := &domain.Expense{ID: expID, CreatedBy: creatorID}

	mockRepo.On("GetExpenseByID", ctx, expID).Return(e, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)

	// Regular member attempt
	err := uc.DeleteExpense(ctx, expID, regularMemberID)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeForbidden, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgExpenseCreatorOnly)

	// Admin attempt (since expense deletion is restricted to creator)
	err = uc.DeleteExpense(ctx, expID, adminID)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeForbidden, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgExpenseCreatorOnly)
}

func TestDeleteExpense_NotFound(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	ctx := context.Background()

	expID := "non-existent"
	mockRepo.On("GetExpenseByID", ctx, expID).Return((*domain.Expense)(nil), nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	err := uc.DeleteExpense(ctx, expID, "usr-1")

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeNotFound, appErr.Type)
}
