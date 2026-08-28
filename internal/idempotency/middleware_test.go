package idempotency_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/idempotency"
	"github.com/Saurrabhh/splittr_be/internal/idempotency/domain"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockIdempotencyRepo struct {
	mock.Mock
}

func (m *mockIdempotencyRepo) Get(ctx context.Context, userID, key string) (*domain.IdempotencyRecord, error) {
	args := m.Called(ctx, userID, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IdempotencyRecord), args.Error(1)
}

func (m *mockIdempotencyRepo) Create(ctx context.Context, rec *domain.IdempotencyRecord) (*domain.IdempotencyRecord, error) {
	args := m.Called(ctx, rec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IdempotencyRecord), args.Error(1)
}

func (m *mockIdempotencyRepo) UpdateResponse(ctx context.Context, userID, key string, code int, headers map[string]string, body []byte) error {
	return m.Called(ctx, userID, key, code, headers, body).Error(0)
}

func TestIdempotencyMiddleware_NoHeader(t *testing.T) {
	mockRepo := new(mockIdempotencyRepo)
	mw := idempotency.NewMiddleware(mockRepo)

	handlerCalled := false
	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/expenses", bytes.NewBufferString(`{"amount":10}`))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestIdempotencyMiddleware_KeyTooLong(t *testing.T) {
	mockRepo := new(mockIdempotencyRepo)
	mw := idempotency.NewMiddleware(mockRepo)

	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	longKey := strings.Repeat("a", 256)
	req := httptest.NewRequest(http.MethodPost, "/v1/expenses", bytes.NewBufferString(`{}`))
	req.Header.Set("X-Idempotency-Key", longKey)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIdempotencyMiddleware_FreshKey_Success(t *testing.T) {
	mockRepo := new(mockIdempotencyRepo)
	mw := idempotency.NewMiddleware(mockRepo)

	testUser := &user.User{ID: "usr-123", Name: "Test User"}

	mockRepo.On("Get", mock.Anything, "usr-123", "idem-uuid-1").Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *domain.IdempotencyRecord) bool {
		return r.Key == "idem-uuid-1" && r.UserID == "usr-123" && r.RequestPath == "/v1/expenses"
	})).Return(&domain.IdempotencyRecord{Key: "idem-uuid-1"}, nil)

	mockRepo.On("UpdateResponse", mock.Anything, "usr-123", "idem-uuid-1", http.StatusCreated, mock.Anything, []byte(`{"id":"exp-1"}`)).Return(nil)

	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"exp-1"}`))
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/expenses", bytes.NewBufferString(`{"amount":100}`))
	req.Header.Set("X-Idempotency-Key", "idem-uuid-1")
	ctx := user.WithUser(req.Context(), testUser)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, `{"id":"exp-1"}`, w.Body.String())
	mockRepo.AssertExpectations(t)
}

func testHash(method, path, body string) string {
	h := sha256.New()
	h.Write([]byte(method + " " + path + "\n"))
	h.Write([]byte(body))
	return hex.EncodeToString(h.Sum(nil))
}

func TestIdempotencyMiddleware_Replay_CacheHit(t *testing.T) {
	mockRepo := new(mockIdempotencyRepo)
	mw := idempotency.NewMiddleware(mockRepo)

	testUser := &user.User{ID: "usr-123", Name: "Test User"}

	code := http.StatusCreated
	cachedRecord := &domain.IdempotencyRecord{
		Key:           "idem-uuid-1",
		UserID:        "usr-123",
		RequestPath:   "/v1/expenses",
		RequestMethod: http.MethodPost,
		RequestHash:   testHash(http.MethodPost, "/v1/expenses", `{"amount":100}`),
		ResponseCode:  &code,
		ResponseHeaders: map[string]string{
			"Content-Type": "application/json",
		},
		ResponseBody: []byte(`{"id":"exp-cached"}`),
	}

	mockRepo.On("Get", mock.Anything, "usr-123", "idem-uuid-1").Return(cachedRecord, nil)

	handlerExecuted := false
	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerExecuted = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/expenses", bytes.NewBufferString(`{"amount":100}`))
	req.Header.Set("X-Idempotency-Key", "idem-uuid-1")
	ctx := user.WithUser(req.Context(), testUser)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.False(t, handlerExecuted)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "true", w.Header().Get("X-Idempotency-Hit"))
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, `{"id":"exp-cached"}`, w.Body.String())
	mockRepo.AssertExpectations(t)
}

func TestIdempotencyMiddleware_PayloadMismatch_422(t *testing.T) {
	mockRepo := new(mockIdempotencyRepo)
	mw := idempotency.NewMiddleware(mockRepo)

	testUser := &user.User{ID: "usr-123", Name: "Test User"}

	code := http.StatusCreated
	cachedRecord := &domain.IdempotencyRecord{
		Key:           "idem-uuid-1",
		UserID:        "usr-123",
		RequestPath:   "/v1/expenses",
		RequestMethod: http.MethodPost,
		RequestHash:   "different-hash",
		ResponseCode:  &code,
	}

	mockRepo.On("Get", mock.Anything, "usr-123", "idem-uuid-1").Return(cachedRecord, nil)

	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodPost, "/v1/expenses", bytes.NewBufferString(`{"amount":100}`))
	req.Header.Set("X-Idempotency-Key", "idem-uuid-1")
	ctx := user.WithUser(req.Context(), testUser)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestIdempotencyMiddleware_ConcurrentInFlight_409(t *testing.T) {
	mockRepo := new(mockIdempotencyRepo)
	mw := idempotency.NewMiddleware(mockRepo)

	testUser := &user.User{ID: "usr-123", Name: "Test User"}

	inFlightRecord := &domain.IdempotencyRecord{
		Key:           "idem-uuid-1",
		UserID:        "usr-123",
		RequestPath:   "/v1/expenses",
		RequestMethod: http.MethodPost,
		RequestHash:   testHash(http.MethodPost, "/v1/expenses", `{"amount":100}`),
		ResponseCode:  nil,
		LockedAt:      time.Now(),
	}

	mockRepo.On("Get", mock.Anything, "usr-123", "idem-uuid-1").Return(inFlightRecord, nil)

	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodPost, "/v1/expenses", bytes.NewBufferString(`{"amount":100}`))
	req.Header.Set("X-Idempotency-Key", "idem-uuid-1")
	ctx := user.WithUser(req.Context(), testUser)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockRepo.AssertExpectations(t)
}

