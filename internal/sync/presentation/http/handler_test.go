package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	expensedomain "github.com/Saurrabhh/splittr_be/internal/expense/domain"
	groupdomain "github.com/Saurrabhh/splittr_be/internal/group/domain"
	"github.com/Saurrabhh/splittr_be/internal/response"
	syncdomain "github.com/Saurrabhh/splittr_be/internal/sync/domain"
	synchttp "github.com/Saurrabhh/splittr_be/internal/sync/presentation/http"
	"github.com/Saurrabhh/splittr_be/internal/user"
	userdomain "github.com/Saurrabhh/splittr_be/internal/user/domain"
	"github.com/go-chi/chi/v5"
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

func setupHandlerTestRouter(uc *syncdomain.UseCase, currentUser *user.User) chi.Router {
	r := chi.NewRouter()
	h := synchttp.NewHandler(uc)

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

func TestHandler_Sync_Success(t *testing.T) {
	mockUser := new(mockUserSyncService)
	mockGroup := new(mockGroupSyncService)
	mockExpense := new(mockExpenseSyncService)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}

	mockUser.On("SyncFriends", mock.Anything, int64(10), currentUser.ID, int32(50)).Return(&userdomain.FriendSyncResponse{
		NewVersion: 15,
		Updated: []userdomain.FriendshipSyncRecord{
			{UserID: "usr-1", FriendID: "usr-2", Status: userdomain.Accepted, SyncVersion: 15},
		},
		DeletedIDs: []string{"usr-3"},
	}, nil)

	mockGroup.On("SyncGroups", mock.Anything, int64(20), currentUser.ID, int32(50)).Return(&groupdomain.GroupSyncResponse{
		NewVersion: 25,
		Updated: []groupdomain.Group{
			{ID: "grp-1", Name: "Trip", SyncVersion: 25},
		},
		DeletedIDs: []string{"grp-2"},
	}, nil)

	mockExpense.On("SyncExpenses", mock.Anything, int64(30), currentUser.ID, int32(50)).Return(&expensedomain.ExpenseSyncResponse{
		NewVersion: 35,
		Updated: []expensedomain.ExpenseWithSplits{
			{Expense: expensedomain.Expense{ID: "exp-1", Amount: 50.0, SyncVersion: 35}},
		},
		DeletedIDs: []string{"exp-2"},
	}, nil)

	uc := syncdomain.NewUseCase(mockUser, mockGroup, mockExpense)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/sync?friendsVersion=10&groupsVersion=20&expensesVersion=30&limit=50", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp syncdomain.SyncResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, int64(15), resp.Friends.NewVersion)
	assert.Equal(t, int64(25), resp.Groups.NewVersion)
	assert.Equal(t, int64(35), resp.Expenses.NewVersion)
}

func TestHandler_Sync_Unauthorized(t *testing.T) {
	mockUser := new(mockUserSyncService)
	mockGroup := new(mockGroupSyncService)
	mockExpense := new(mockExpenseSyncService)

	uc := syncdomain.NewUseCase(mockUser, mockGroup, mockExpense)
	router := setupHandlerTestRouter(uc, nil)

	req := httptest.NewRequest(http.MethodGet, "/sync", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandler_Sync_InternalServerError(t *testing.T) {
	mockUser := new(mockUserSyncService)
	mockGroup := new(mockGroupSyncService)
	mockExpense := new(mockExpenseSyncService)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}

	mockUser.On("SyncFriends", mock.Anything, int64(0), currentUser.ID, int32(100)).Return(nil, errors.New("db error"))

	uc := syncdomain.NewUseCase(mockUser, mockGroup, mockExpense)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/sync", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
