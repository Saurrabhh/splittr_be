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

func (m *mockUserRepository) UpdateAvatar(ctx context.Context, userID string, avatarURL string) (*domain.User, error) {
	args := m.Called(ctx, userID, avatarURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) GetByEmailOrPhone(ctx context.Context, email, phone string) (*domain.User, error) {
	args := m.Called(ctx, email, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) GetByEmailOrPhoneWithSettings(ctx context.Context, email, phone string) (*domain.UserWithSettings, error) {
	args := m.Called(ctx, email, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserWithSettings), args.Error(1)
}

func (m *mockUserRepository) CreateDefaultSettings(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockUserRepository) GetSettings(ctx context.Context, userID string) (*domain.UserSettings, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserSettings), args.Error(1)
}

func (m *mockUserRepository) UpsertSettings(ctx context.Context, settings *domain.UserSettings) error {
	return m.Called(ctx, settings).Error(0)
}

func (m *mockUserRepository) CreateFriendship(ctx context.Context, userID, friendID string, status domain.FriendshipStatus, actionUserID string) error {
	return m.Called(ctx, userID, friendID, status, actionUserID).Error(0)
}

func (m *mockUserRepository) UpdateFriendshipStatus(ctx context.Context, userID, friendID string, status domain.FriendshipStatus, actionUserID string) error {
	return m.Called(ctx, userID, friendID, status, actionUserID).Error(0)
}

func (m *mockUserRepository) DeleteFriendship(ctx context.Context, userID, friendID string) error {
	return m.Called(ctx, userID, friendID).Error(0)
}

func (m *mockUserRepository) GetFriendship(ctx context.Context, userID, friendID string) (*domain.Friendship, error) {
	args := m.Called(ctx, userID, friendID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Friendship), args.Error(1)
}

func (m *mockUserRepository) ListFriends(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.User, error) {
	args := m.Called(ctx, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *mockUserRepository) ListFriendsByStatus(ctx context.Context, userID string, status domain.FriendshipStatus) ([]domain.FriendWithStatus, error) {
	args := m.Called(ctx, userID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.FriendWithStatus), args.Error(1)
}

func (m *mockUserRepository) SyncFriendsBySequence(ctx context.Context, lastVersion int64, userID string, limit int32) ([]domain.FriendshipSyncRecord, error) {
	args := m.Called(ctx, lastVersion, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.FriendshipSyncRecord), args.Error(1)
}

func (m *mockUserRepository) GetFriendTombstonesBySequence(ctx context.Context, lastVersion int64, userID string, limit int32) ([]domain.Tombstone, error) {
	args := m.Called(ctx, lastVersion, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Tombstone), args.Error(1)
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
	mockRepo.On("CreateDefaultSettings", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	uc := domain.NewUseCase(mockRepo, nil)
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

// --- GET/PATCH /users/me Tests ---

func TestHandler_UpdateMe_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice", DefaultCurrency: "USD"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("GetByID", mock.Anything, "usr-1").Return(currentUser, nil)
	mockRepo.On("UpdateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	uc := domain.NewUseCase(mockRepo, nil)
	router := setupHandlerTestRouter(uc, identity)

	body, _ := json.Marshal(map[string]string{"name": "Alice Updated", "defaultCurrency": "EUR"})
	req := httptest.NewRequest(http.MethodPatch, "/users/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp domain.User
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Alice Updated", resp.Name)
	assert.Equal(t, "EUR", resp.DefaultCurrency)
}

// --- GET/PATCH /users/me/settings Tests ---

func TestHandler_GetSettings_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("GetSettings", mock.Anything, "usr-1").Return(&domain.UserSettings{UserID: "usr-1", AutoAcceptFriendRequests: true}, nil)

	uc := domain.NewUseCase(mockRepo, nil)
	router := setupHandlerTestRouter(uc, identity)

	req := httptest.NewRequest(http.MethodGet, "/users/me/settings", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp userhttp.UserSettingsResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.AutoAcceptFriendRequests)
}

func TestHandler_UpdateSettings_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("UpsertSettings", mock.Anything, mock.AnythingOfType("*domain.UserSettings")).Return(nil)

	uc := domain.NewUseCase(mockRepo, nil)
	router := setupHandlerTestRouter(uc, identity)

	body, _ := json.Marshal(map[string]bool{"autoAcceptFriendRequests": true})
	req := httptest.NewRequest(http.MethodPatch, "/users/me/settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp userhttp.UserSettingsResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.AutoAcceptFriendRequests)
}

// --- POST /friends Tests ---

func TestHandler_AddFriend_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}
	friendUserWithSettings := &domain.UserWithSettings{
		User:                     domain.User{ID: "usr-2", Name: "Bob"},
		AutoAcceptFriendRequests: false,
	}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("GetByEmailOrPhoneWithSettings", mock.Anything, "bob@example.com", "").Return(friendUserWithSettings, nil)
	mockRepo.On("GetFriendship", mock.Anything, "usr-1", "usr-2").Return(nil, nil)
	mockRepo.On("CreateFriendship", mock.Anything, "usr-1", "usr-2", domain.Pending, "usr-1").Return(nil)

	uc := domain.NewUseCase(mockRepo, nil)
	router := setupHandlerTestRouter(uc, identity)

	body, _ := json.Marshal(map[string]string{"friendEmail": "bob@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/friends", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp domain.FriendWithStatus
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "usr-2", resp.ID)
	assert.Equal(t, domain.Pending, resp.Status)
}

// --- PATCH /friends/{friendId} Tests ---

func TestHandler_UpdateFriendStatus_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("GetFriendship", mock.Anything, "usr-1", "usr-2").Return(&domain.Friendship{
		UserID:   "usr-2",
		FriendID: "usr-1",
		Status:   domain.Pending,
	}, nil)
	mockRepo.On("UpdateFriendshipStatus", mock.Anything, "usr-1", "usr-2", domain.Accepted, "usr-1").Return(nil)
	mockRepo.On("GetByID", mock.Anything, "usr-2").Return(&domain.User{ID: "usr-2", Name: "Bob"}, nil)

	uc := domain.NewUseCase(mockRepo, nil)
	router := setupHandlerTestRouter(uc, identity)

	body, _ := json.Marshal(map[string]string{"status": "ACCEPTED"})
	req := httptest.NewRequest(http.MethodPatch, "/friends/usr-2", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp domain.FriendWithStatus
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "usr-2", resp.ID)
	assert.Equal(t, domain.Accepted, resp.Status)
}

// --- DELETE /friends/{friendId} Tests ---

func TestHandler_RemoveFriend_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("GetFriendship", mock.Anything, "usr-1", "usr-2").Return(&domain.Friendship{
		UserID:   "usr-1",
		FriendID: "usr-2",
		Status:   domain.Accepted,
	}, nil)
	mockRepo.On("DeleteFriendship", mock.Anything, "usr-1", "usr-2").Return(nil)

	uc := domain.NewUseCase(mockRepo, nil)
	router := setupHandlerTestRouter(uc, identity)

	req := httptest.NewRequest(http.MethodDelete, "/friends/usr-2", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestHandler_SyncFriends_Success(t *testing.T) {
	mockRepo := new(mockUserRepository)
	identity := &auth.Identity{UserID: "fb-123"}
	currentUser := &domain.User{ID: "usr-1", FirebaseUID: "fb-123", Name: "Alice"}

	mockRepo.On("GetByFirebaseUID", mock.Anything, "fb-123").Return(currentUser, nil)
	mockRepo.On("SyncFriendsBySequence", mock.Anything, int64(10), currentUser.ID, int32(100)).Return([]domain.FriendshipSyncRecord{
		{UserID: "usr-1", FriendID: "usr-2", Status: domain.Accepted, SyncVersion: 11},
	}, nil)
	mockRepo.On("GetFriendTombstonesBySequence", mock.Anything, int64(10), currentUser.ID, int32(100)).Return([]domain.Tombstone{
		{EntityID: "usr-3", SyncVersion: 12},
	}, nil)

	uc := domain.NewUseCase(mockRepo, nil)
	router := setupHandlerTestRouter(uc, identity)

	req := httptest.NewRequest(http.MethodGet, "/friends/sync?lastVersion=10", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp domain.FriendSyncResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, int64(12), resp.NewVersion)
	assert.Len(t, resp.Friends, 1)
	assert.Equal(t, []string{"usr-3"}, resp.RemovedFriendIDs)
}

