package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saurrabhh/splittr_be/internal/appconfig/domain"
	appconfighttp "github.com/Saurrabhh/splittr_be/internal/appconfig/presentation/http"
	"github.com/go-chi/chi/v5"
)

type mockRepository struct {
	versionHash string
	categories  []domain.Category
	currencies  []domain.Currency
	appVersion  domain.AppVersion
	maintenance domain.MaintenanceConfig
	limits      domain.LimitsConfig
	flags       map[string]bool
	legal       domain.LegalConfig
}

func (m *mockRepository) GetAppVersion(ctx context.Context) (domain.AppVersion, error) {
	return m.appVersion, nil
}

func (m *mockRepository) GetMaintenanceStatus(ctx context.Context) (domain.MaintenanceConfig, error) {
	return m.maintenance, nil
}

func (m *mockRepository) GetSystemLimits(ctx context.Context) (domain.LimitsConfig, error) {
	return m.limits, nil
}

func (m *mockRepository) GetFeatureFlags(ctx context.Context) (map[string]bool, error) {
	return m.flags, nil
}

func (m *mockRepository) GetLegalConfigs(ctx context.Context) (domain.LegalConfig, error) {
	return m.legal, nil
}

func (m *mockRepository) GetCategories(ctx context.Context) ([]domain.Category, error) {
	return m.categories, nil
}

func (m *mockRepository) GetCurrencies(ctx context.Context) ([]domain.Currency, error) {
	return m.currencies, nil
}

func (m *mockRepository) GetVersionHash(ctx context.Context) (string, error) {
	return m.versionHash, nil
}

func TestHandler_GetAppConfig_Public(t *testing.T) {
	mockRepo := &mockRepository{
		versionHash: "v1.0.0-test",
		categories:  []domain.Category{{ID: "cat_food", Name: "Food", IconURL: "https://assets.splittr.app/food.png"}},
		currencies:  []domain.Currency{{Code: "USD", Symbol: "$", Name: "Dollar", DecimalPlaces: 2, IsDefault: true}},
		appVersion:  domain.AppVersion{MinSupportedVersion: "1.0.0", LatestVersion: "1.0.0"},
		maintenance: domain.MaintenanceConfig{InMaintenance: false},
		limits:      domain.LimitsConfig{MaxExpenseAmount: 1000},
		flags:       map[string]bool{"enableOcr": true},
		legal:       domain.LegalConfig{SupportEmail: "test@splittr.app"},
	}
	uc := domain.NewUsecase(mockRepo)
	h := appconfighttp.NewHandler(uc)

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
	uc := domain.NewUsecase(mockRepo)
	h := appconfighttp.NewHandler(uc)

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
