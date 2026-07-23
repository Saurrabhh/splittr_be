package activity_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupActivityHandlerTestRouter(uc *activity.UseCase, currentUser *user.User) chi.Router {
	r := chi.NewRouter()
	h := activity.NewHandler(uc)

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

	expectedActs := []activity.Activity{
		{ID: "act-1", Description: "Activity 1", ActionType: activity.ActionTypeExpenseCreated, EntityType: activity.EntityTypeExpense, CreatedAt: time.Now()},
	}

	mockRepo.On("ListUserActivities", mock.Anything, currentUser.ID, int32(21), (*time.Time)(nil), (*string)(nil)).Return(expectedActs, nil)

	uc := activity.NewUseCase(mockRepo)
	router := setupActivityHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/activities", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp activity.ListActivitiesResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "act-1", resp.Data[0].ID)
}

func TestHandler_ListActivities_Unauthorized(t *testing.T) {
	mockRepo := new(mockActivityRepository)
	uc := activity.NewUseCase(mockRepo)
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

	uc := activity.NewUseCase(mockRepo)
	router := setupActivityHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/activities", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
