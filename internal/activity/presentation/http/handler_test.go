package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity/domain"
	activityhttp "github.com/Saurrabhh/splittr_be/internal/activity/presentation/http"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockActivityRepository struct {
	mock.Mock
}

func (m *mockActivityRepository) CreateActivity(ctx context.Context, act *domain.Activity, rawPayload []byte) error {
	return m.Called(ctx, act, rawPayload).Error(0)
}

func (m *mockActivityRepository) CreateActivityVisibility(ctx context.Context, activityID string, userID string) error {
	return m.Called(ctx, activityID, userID).Error(0)
}

func (m *mockActivityRepository) ListUserActivities(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.Activity, error) {
	args := m.Called(ctx, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Activity), args.Error(1)
}

func (m *mockActivityRepository) ListGroupFeed(ctx context.Context, groupID string, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.Activity, error) {
	args := m.Called(ctx, groupID, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Activity), args.Error(1)
}

func setupActivityHandlerTestRouter(uc *domain.UseCase, currentUser *user.User) chi.Router {
	r := chi.NewRouter()
	h := activityhttp.NewHandler(uc)

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

// --- GET /activities Tests ---

func TestHandler_ListActivities_Success(t *testing.T) {
	mockRepo := new(mockActivityRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}

	expectedActs := []domain.Activity{
		{ID: "act-1", Description: "Activity 1", ActionType: domain.ActionTypeExpenseCreated, EntityType: domain.EntityTypeExpense, CreatedAt: time.Now()},
	}

	mockRepo.On("ListUserActivities", mock.Anything, currentUser.ID, int32(21), (*time.Time)(nil), (*string)(nil)).Return(expectedActs, nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupActivityHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/activities", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp pagination.Response[domain.Activity]
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "act-1", resp.Data[0].ID)
}

func TestHandler_ListActivities_Unauthorized(t *testing.T) {
	mockRepo := new(mockActivityRepository)
	uc := domain.NewUseCase(mockRepo)
	router := setupActivityHandlerTestRouter(uc, nil) // Unauthenticated

	req := httptest.NewRequest(http.MethodGet, "/activities", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandler_ListActivities_InternalServerError(t *testing.T) {
	mockRepo := new(mockActivityRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}

	mockRepo.On("ListUserActivities", mock.Anything, currentUser.ID, int32(21), (*time.Time)(nil), (*string)(nil)).Return(nil, errors.New("db read failure"))

	uc := domain.NewUseCase(mockRepo)
	router := setupActivityHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/activities", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
