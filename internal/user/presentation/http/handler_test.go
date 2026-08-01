package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/auth"
	"github.com/Saurrabhh/splittr_be/internal/user/domain"
	userhttp "github.com/Saurrabhh/splittr_be/internal/user/presentation/http"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*domain.User, error) {
	args := m.Called(ctx, firebaseUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) Create(ctx context.Context, u *domain.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepository) UpdateUser(ctx context.Context, u *domain.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepository) GetByEmailOrPhone(ctx context.Context, email, phone string) (*domain.User, error) {
	args := m.Called(ctx, email, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) CreateFriendship(ctx context.Context, userID, friendID string) error {
	return m.Called(ctx, userID, friendID).Error(0)
}

func (m *mockUserRepository) DeleteFriendship(ctx context.Context, userID, friendID string) error {
	return m.Called(ctx, userID, friendID).Error(0)
}

func (m *mockUserRepository) GetFriendship(ctx context.Context, userID, friendID string) (bool, error) {
	args := m.Called(ctx, userID, friendID)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepository) ListFriends(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.User, error) {
	args := m.Called(ctx, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func setupHandlerTestRouter(uc *domain.UseCase, identity *auth.Identity) chi.Router {
	r := chi.NewRouter()
	h := userhttp.NewHandler(uc)

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if identity != nil {
				ctx := auth.WithIdentity(r.Context(), identity)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}

	h.RegisterRoutes(r, authMiddleware)
	return r
}

// --- POST /users (Register) Tests ---

func TestHandler_Register_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123", Email: "alice@example.com"}

	email := "alice@example.com"
	expectedUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice", Email: &email}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Run(func(args mock.Arguments) {
		u := args.Get(1).(*domain.User)
		u.ID = "usr-1"
	}).Return(nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, identity)

	body, _ := json.Marshal(map[string]string{"name": "Alice"})
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var respUser domain.User
	err := json.Unmarshal(rr.Body.Bytes(), &respUser)
	require.NoError(t, err)
	assert.Equal(t, expectedUser.ID, respUser.ID)
	assert.Equal(t, expectedUser.Name, respUser.Name)
}

func TestHandler_Register_Unauthorized(t *testing.T) {
	mockRepo := new(mockUserRepository)
	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, nil) // nil identity

	body, _ := json.Marshal(map[string]string{"name": "Alice"})
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandler_Register_BadRequest(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123", Email: "alice@example.com"}
	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, identity)

	// Missing name field
	body, _ := json.Marshal(map[string]string{"name": ""})
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- GET /users/me Tests ---

func TestHandler_GetMe_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, identity)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respUser domain.User
	err := json.Unmarshal(rr.Body.Bytes(), &respUser)
	require.NoError(t, err)
	assert.Equal(t, "usr-1", respUser.ID)
	assert.Equal(t, "Alice", respUser.Name)
}

func TestHandler_GetMe_Unauthorized(t *testing.T) {
	mockRepo := new(mockUserRepository)
	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, nil) // no identity

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandler_GetMe_UserNotFound(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-unknown"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-unknown").Return(nil, nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, identity)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// --- PUT /users/me Tests ---

func TestHandler_UpdateMe_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice", DefaultCurrency: "INR"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("GetByID", mock.Anything, "usr-1").Return(currentUser, nil)
	mockRepo.On("UpdateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, identity)

	body, _ := json.Marshal(map[string]string{"name": "Alice Updated", "defaultCurrency": "USD"})
	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respUser domain.User
	err := json.Unmarshal(rr.Body.Bytes(), &respUser)
	require.NoError(t, err)
	assert.Equal(t, "Alice Updated", respUser.Name)
	assert.Equal(t, "USD", respUser.DefaultCurrency)
}

func TestHandler_UpdateMe_BadRequest_InvalidCurrency(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice", DefaultCurrency: "INR"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("GetByID", mock.Anything, "usr-1").Return(currentUser, nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, identity)

	body, _ := json.Marshal(map[string]string{"defaultCurrency": "INVALID"})
	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_UpdateMe_NotFound(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("GetByID", mock.Anything, "usr-1").Return(nil, nil) // not found in usecase GetByID

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, identity)

	body, _ := json.Marshal(map[string]string{"name": "Alice Updated"})
	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_UpdateMe_Unauthorized(t *testing.T) {
	mockRepo := new(mockUserRepository)
	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, nil)

	body, _ := json.Marshal(map[string]string{"name": "Alice Updated"})
	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- POST /friends Tests ---

func TestHandler_AddFriend_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}
	friendUser := &domain.User{ID: "usr-2", Name: "Bob"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("GetByEmailOrPhone", mock.Anything, "bob@example.com", "").Return(friendUser, nil)
	mockRepo.On("GetFriendship", mock.Anything, "usr-1", "usr-2").Return(false, nil)
	mockRepo.On("CreateFriendship", mock.Anything, "usr-1", "usr-2").Return(nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, identity)

	body, _ := json.Marshal(map[string]string{"friendEmail": "bob@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/friends", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respUser domain.User
	err := json.Unmarshal(rr.Body.Bytes(), &respUser)
	require.NoError(t, err)
	assert.Equal(t, "usr-2", respUser.ID)
}

func TestHandler_AddFriend_BadRequest_MissingFields(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, identity)

	body, _ := json.Marshal(map[string]string{"friendEmail": "", "friendPhone": ""})
	req := httptest.NewRequest(http.MethodPost, "/friends", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_AddFriend_NotFound(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("GetByEmailOrPhone", mock.Anything, "unknown@example.com", "").Return(nil, nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, identity)

	body, _ := json.Marshal(map[string]string{"friendEmail": "unknown@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/friends", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_AddFriend_Unauthorized(t *testing.T) {
	mockRepo := new(mockUserRepository)
	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, nil)

	body, _ := json.Marshal(map[string]string{"friendEmail": "bob@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/friends", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- DELETE /friends/{friendId} Tests ---

func TestHandler_RemoveFriend_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("GetFriendship", mock.Anything, "usr-1", "usr-2").Return(true, nil)
	mockRepo.On("DeleteFriendship", mock.Anything, "usr-1", "usr-2").Return(nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, identity)

	req := httptest.NewRequest(http.MethodDelete, "/friends/usr-2", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestHandler_RemoveFriend_BadRequest_NotFriends(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("GetFriendship", mock.Anything, "usr-1", "usr-2").Return(false, nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, identity)

	req := httptest.NewRequest(http.MethodDelete, "/friends/usr-2", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_RemoveFriend_Unauthorized(t *testing.T) {
	mockRepo := new(mockUserRepository)
	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, nil)

	req := httptest.NewRequest(http.MethodDelete, "/friends/usr-2", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- GET /friends Tests ---

func TestHandler_GetFriends_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}
	friendsList := []domain.User{{ID: "usr-2", Name: "Bob"}}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("ListFriends", mock.Anything, "usr-1", int32(21), mock.Anything, mock.Anything).Return(friendsList, nil)

	uc := domain.NewUseCase(mockRepo)
	router := setupHandlerTestRouter(uc, identity)

	req := httptest.NewRequest(http.MethodGet, "/friends", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp userhttp.ListFriendsResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "usr-2", resp.Data[0].ID)
}
