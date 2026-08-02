package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/expense/domain"
	expensehttp "github.com/Saurrabhh/splittr_be/internal/expense/presentation/http"
	"github.com/Saurrabhh/splittr_be/internal/group"
	"github.com/Saurrabhh/splittr_be/internal/notification"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
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

func (m *mockGroupService) GetGroupDetails(ctx context.Context, groupID, userID string) (*group.Group, error) {
	args := m.Called(ctx, groupID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*group.Group), args.Error(1)
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

func setupHandlerTestRouter(uc *domain.UseCase, currentUser *user.User) chi.Router {
	r := chi.NewRouter()
	h := expensehttp.NewHandler(uc)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if currentUser != nil {
				ctx := user.WithUser(r.Context(), currentUser)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
			} else {
				response.Unauthorized(w, "unauthorized")
			}
		})
	})

	h.RegisterRoutes(r)
	return r
}

// --- POST /expenses ---

func TestHandler_CreateExpense_Success(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	mockGroupSvc := new(mockGroupService)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}

	currentUser := &user.User{ID: "usr-1", Name: "Alice"}

	mockRepo.On("CreateExpense", mock.Anything, mock.AnythingOfType("*domain.Expense")).Return(nil)
	mockRepo.On("CreateExpenseSplit", mock.Anything, mock.AnythingOfType("*domain.Split")).Return(nil)
	mockRepo.On("ListExpenseSplits", mock.Anything, mock.AnythingOfType("string")).Return([]domain.Split{
		{ExpenseID: "exp-1", UserID: currentUser.ID, Amount: 100.0},
	}, nil)

	mockAct.On("LogEvent",
		mock.Anything, currentUser.ID, (*string)(nil), mock.Anything, mock.Anything,
	).Return(&activity.Activity{ID: "act-1"}, nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockGroupSvc, mockAct, mockNotif)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]interface{}{
		"description": "Groceries",
		"amount":      100.0,
		"currency":    "INR",
		"splitType":   "EQUAL",
		"splits": []map[string]interface{}{
			{"userId": currentUser.ID},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var resp domain.ExpenseWithSplits
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Groceries", resp.Expense.Description)
}

func TestHandler_CreateExpense_BadRequest(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}
	router := setupHandlerTestRouter(uc, currentUser)

	// Amount <= 0 validation error
	body, _ := json.Marshal(map[string]interface{}{
		"description": "Lunch",
		"amount":      -10.0,
		"splitType":   "EQUAL",
		"splits": []map[string]interface{}{
			{"userId": currentUser.ID},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_CreateExpense_Unauthorized(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	router := setupHandlerTestRouter(uc, nil) // no user context

	body, _ := json.Marshal(map[string]interface{}{
		"description": "Lunch",
		"amount":      100.0,
	})

	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- POST /expenses/settle ---

func TestHandler_SettleUp_Success(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}

	currentUser := &user.User{ID: "usr-1", Name: "Alice"}

	mockRepo.On("CreateExpense", mock.Anything, mock.AnythingOfType("*domain.Expense")).Return(nil)
	mockRepo.On("CreateExpenseSplit", mock.Anything, mock.AnythingOfType("*domain.Split")).Return(nil)
	mockRepo.On("ListExpenseSplits", mock.Anything, mock.AnythingOfType("string")).Return([]domain.Split{
		{ExpenseID: "exp-1", UserID: "usr-2", Amount: 50.0},
	}, nil)

	mockAct.On("LogEvent",
		mock.Anything, currentUser.ID, (*string)(nil), mock.Anything, mock.Anything,
	).Return(&activity.Activity{ID: "act-1"}, nil)

	mockNotif.On("CreateAlert", mock.Anything, "usr-2", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, nil, mockAct, mockNotif)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]interface{}{
		"amount":     50.0,
		"currency":   "INR",
		"paidBy":     currentUser.ID,
		"receivedBy": "usr-2",
	})

	req := httptest.NewRequest(http.MethodPost, "/expenses/settle", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestHandler_SettleUp_BadRequest(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]interface{}{
		"amount":     50.0,
		"paidBy":     currentUser.ID,
		"receivedBy": currentUser.ID, // paidBy == receivedBy is invalid
	})

	req := httptest.NewRequest(http.MethodPost, "/expenses/settle", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- GET /expenses ---

func TestHandler_ListExpenses_GroupSuccess(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	mockGroupSvc := new(mockGroupService)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}
	groupID := "grp-1"

	mockGroupSvc.On("GetGroupDetails", mock.Anything, groupID, currentUser.ID).Return(&group.Group{ID: groupID}, nil)
	mockRepo.On("ListExpensesByGroup", mock.Anything, groupID, mock.Anything, mock.Anything, mock.Anything).Return([]domain.Expense{
		{ID: "exp-1", Description: "Dinner"},
	}, nil)
	mockRepo.On("ListExpenseSplitsByIDs", mock.Anything, mock.Anything).Return([]domain.Split{}, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, mockGroupSvc, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/expenses?groupId="+groupID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandler_ListExpenses_PersonalSuccess(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}

	mockRepo.On("ListUserPersonalExpenses", mock.Anything, currentUser.ID, mock.Anything, mock.Anything, mock.Anything).Return([]domain.Expense{
		{ID: "exp-1", Description: "Coffee"},
	}, nil)
	mockRepo.On("ListExpenseSplitsByIDs", mock.Anything, mock.Anything).Return([]domain.Split{}, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/expenses?personal=true", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandler_ListExpenses_FriendSuccess(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}
	friendID := "usr-2"

	mockRepo.On("ListUserFriendExpenses", mock.Anything, currentUser.ID, mock.Anything, mock.Anything, mock.Anything).Return([]domain.Expense{
		{ID: "exp-1", Description: "Taxi"},
	}, nil)
	mockRepo.On("ListExpenseSplitsByIDs", mock.Anything, mock.Anything).Return([]domain.Split{}, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/expenses?friendId="+friendID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandler_ListExpenses_MissingFilter(t *testing.T) {
	uc := domain.NewUseCase(new(mockExpenseRepository), &mockTransactor{}, nil, nil, nil)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/expenses", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- GET /expenses/{id} ---

func TestHandler_GetDetails_Success(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}
	expID := "exp-1"

	exp := &domain.Expense{ID: expID, Description: "Lunch", PaidBy: currentUser.ID, CreatedBy: currentUser.ID}
	splits := []domain.Split{{ExpenseID: expID, UserID: currentUser.ID, Amount: 50.0}}

	mockRepo.On("GetExpenseByID", mock.Anything, expID).Return(exp, nil)
	mockRepo.On("ListExpenseSplits", mock.Anything, expID).Return(splits, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/expenses/"+expID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandler_GetDetails_NotFound(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}
	expID := "non-existent"

	mockRepo.On("GetExpenseByID", mock.Anything, expID).Return((*domain.Expense)(nil), nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/expenses/"+expID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_GetDetails_Forbidden(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	currentUser := &user.User{ID: "usr-unrelated", Name: "Stranger"}
	expID := "exp-1"

	exp := &domain.Expense{ID: expID, Description: "Lunch", PaidBy: "usr-creator", CreatedBy: "usr-creator"}
	splits := []domain.Split{{ExpenseID: expID, UserID: "usr-other", Amount: 50.0}}

	mockRepo.On("GetExpenseByID", mock.Anything, expID).Return(exp, nil)
	mockRepo.On("ListExpenseSplits", mock.Anything, expID).Return(splits, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/expenses/"+expID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// --- DELETE /expenses/{id} ---

func TestHandler_Delete_Success(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	currentUser := &user.User{ID: "usr-creator", Name: "Creator"}
	expID := "exp-1"

	exp := &domain.Expense{ID: expID, CreatedBy: currentUser.ID}

	mockRepo.On("GetExpenseByID", mock.Anything, expID).Return(exp, nil)
	mockRepo.On("DeleteExpense", mock.Anything, expID).Return(nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/expenses/"+expID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestHandler_Delete_Forbidden(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	currentUser := &user.User{ID: "usr-member", Name: "Member"}
	expID := "exp-1"

	exp := &domain.Expense{ID: expID, CreatedBy: "usr-creator"}

	mockRepo.On("GetExpenseByID", mock.Anything, expID).Return(exp, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/expenses/"+expID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_Delete_NotFound(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}
	expID := "non-existent"

	mockRepo.On("GetExpenseByID", mock.Anything, expID).Return((*domain.Expense)(nil), nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/expenses/"+expID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- GET /balances ---

func TestHandler_GetBalances_Success(t *testing.T) {
	mockRepo := new(mockExpenseRepository)
	mockGroupSvc := new(mockGroupService)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}
	groupID := "grp-1"

	mockGroupSvc.On("GetGroupDetails", mock.Anything, groupID, currentUser.ID).Return(&group.Group{ID: groupID}, nil)
	mockRepo.On("GetGroupBalances", mock.Anything, groupID).Return([]domain.UserBalance{
		{UserID: "usr-1", UserName: "Alice", NetBalance: 50.0},
	}, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, mockGroupSvc, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/balances?groupId="+groupID+"&simplified=true", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
