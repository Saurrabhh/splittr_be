package appconfig

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHandler_GetAppConfig_Public(t *testing.T) {
	mockRepo := &mockRepository{
		versionHash: "v1.0.0-test",
		categories:  []Category{{ID: "cat_food", Name: "Food", IconURL: "https://assets.splittr.app/food.png"}},
		currencies:  []Currency{{Code: "USD", Symbol: "$", Name: "Dollar", DecimalPlaces: 2, IsDefault: true}},
		appVersion:  AppVersion{MinSupportedVersion: "1.0.0", LatestVersion: "1.0.0"},
		maintenance: MaintenanceConfig{InMaintenance: false},
		limits:      LimitsConfig{MaxExpenseAmount: 1000},
		flags:       map[string]bool{"enableOcr": true},
		legal:       LegalConfig{SupportEmail: "test@splittr.app"},
	}
	uc := NewUsecase(mockRepo)
	h := NewHandler(uc)

	r := chi.NewRouter()
	noopAuth := func(next http.Handler) http.Handler {
		return next
	}
	h.RegisterRoutes(r, noopAuth)

	req := httptest.NewRequest("GET", "/app-config", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", w.Code)
	}

	etag := w.Header().Get("ETag")
	if etag != "v1.0.0-test" {
		t.Errorf("expected ETag v1.0.0-test, got %s", etag)
	}
}

func TestHandler_GetAppConfig_ETagNotModified(t *testing.T) {
	mockRepo := &mockRepository{
		versionHash: "v1.0.0-test",
	}
	uc := NewUsecase(mockRepo)
	h := NewHandler(uc)

	r := chi.NewRouter()
	noopAuth := func(next http.Handler) http.Handler {
		return next
	}
	h.RegisterRoutes(r, noopAuth)

	req := httptest.NewRequest("GET", "/app-config", nil)
	req.Header.Set("If-None-Match", "v1.0.0-test")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotModified {
		t.Errorf("expected status 304 Not Modified, got %d", w.Code)
	}
}
