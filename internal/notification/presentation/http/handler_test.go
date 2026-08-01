package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/notification/domain"
	notifhttp "github.com/Saurrabhh/splittr_be/internal/notification/presentation/http"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockNotificationRepository struct {
	mock.Mock
}

func (m *mockNotificationRepository) CreateNotification(ctx context.Context, notif *domain.Notification) error {
	return m.Called(ctx, notif).Error(0)
}

func (m *mockNotificationRepository) ListUserNotifications(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.Notification, error) {
	args := m.Called(ctx, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Notification), args.Error(1)
}

func (m *mockNotificationRepository) MarkNotificationAsRead(ctx context.Context, id, userID string) error {
	return m.Called(ctx, id, userID).Error(0)
}

func (m *mockNotificationRepository) MarkAllNotificationsAsRead(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func setupHandlerTestRouter(uc *domain.UseCase, currentUser *user.User) chi.Router {
	r := chi.NewRouter()
	h := notifhttp.NewHandler(uc)

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

// --- GET /notifications Tests ---

func TestHandler_List_Success(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}

	expectedNotifs := []domain.Notification{
		{ID: "notif-1", UserID: currentUser.ID, Title: "Title 1", Content: "Content 1", CreatedAt: time.Now()},
	}

	mockRepo.On("ListUserNotifications", mock.Anything, currentUser.ID, int32(21), (*time.Time)(nil), (*string)(nil)).Return(expectedNotifs, nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp pagination.Response[domain.Notification]
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "notif-1", resp.Data[0].ID)
}

func TestHandler_List_Unauthorized(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, nil) // Unauthenticated

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandler_List_InternalServerError(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}

	mockRepo.On("ListUserNotifications", mock.Anything, currentUser.ID, int32(21), (*time.Time)(nil), (*string)(nil)).Return(nil, errors.New("db error"))

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- POST /notifications/{id}/read Tests ---

func TestHandler_MarkAsRead_Success(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}
	notifID := "notif-123"

	mockRepo.On("MarkNotificationAsRead", mock.Anything, notifID, currentUser.ID).Return(nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodPost, "/notifications/"+notifID+"/read", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp response.MessageResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "notification marked as read", resp.Message)
}

func TestHandler_MarkAsRead_BadRequest(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}

	// When router matches route, if ID is somehow invalid / validation error returns Bad Request
	// Note: uc.MarkAsRead(ctx, "", userID) returns validation error
	uc := domain.NewUseCase(mockRepo)
	router := chi.NewRouter()
	h := notifhttp.NewHandler(uc)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := user.WithUser(r.Context(), currentUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	// Route registered with custom param or handler call directly to simulate empty ID parameter
	router.Post("/notifications/{id}/read", h.MarkAsRead)

	// Custom request targeting handler with empty chi url param (or directly calling handler for empty ID validation)
	req := httptest.NewRequest(http.MethodPost, "/notifications//read", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	// chi router won't match double slash or will fail to extract param
	// Let's also test via direct call to handler or route
	if rr.Code != http.StatusBadRequest {
		// Test directly via chi context with empty URLParam "id"
		r := httptest.NewRequest(http.MethodPost, "/notifications/test/read", nil)
		r = r.WithContext(user.WithUser(r.Context(), currentUser))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "") // empty ID
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		rr2 := httptest.NewRecorder()
		h.MarkAsRead(rr2, r)
		assert.Equal(t, http.StatusBadRequest, rr2.Code)
	}
}

func TestHandler_MarkAsRead_Unauthorized(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, nil) // Unauthenticated

	req := httptest.NewRequest(http.MethodPost, "/notifications/notif-123/read", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandler_MarkAsRead_InternalServerError(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}
	notifID := "notif-123"

	mockRepo.On("MarkNotificationAsRead", mock.Anything, notifID, currentUser.ID).Return(errors.New("db error"))

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodPost, "/notifications/"+notifID+"/read", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- POST /notifications/read-all Tests ---

func TestHandler_MarkAllAsRead_Success(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}

	mockRepo.On("MarkAllNotificationsAsRead", mock.Anything, currentUser.ID).Return(nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodPost, "/notifications/read-all", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp response.MessageResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "all notifications marked as read", resp.Message)
}

func TestHandler_MarkAllAsRead_Unauthorized(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, nil) // Unauthenticated

	req := httptest.NewRequest(http.MethodPost, "/notifications/read-all", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandler_MarkAllAsRead_InternalServerError(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}

	mockRepo.On("MarkAllNotificationsAsRead", mock.Anything, currentUser.ID).Return(errors.New("db error"))

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodPost, "/notifications/read-all", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
