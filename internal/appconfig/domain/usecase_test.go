package domain

import (
	"context"
	"testing"
)

type mockRepository struct {
	versionHash string
	categories  []Category
	currencies  []Currency
	appVersion  AppVersion
	maintenance MaintenanceConfig
	limits      LimitsConfig
	flags       map[string]bool
	legal       LegalConfig
}

func (m *mockRepository) GetAppVersion(ctx context.Context) (AppVersion, error) {
	return m.appVersion, nil
}

func (m *mockRepository) GetMaintenanceStatus(ctx context.Context) (MaintenanceConfig, error) {
	return m.maintenance, nil
}

func (m *mockRepository) GetSystemLimits(ctx context.Context) (LimitsConfig, error) {
	return m.limits, nil
}

func (m *mockRepository) GetFeatureFlags(ctx context.Context) (map[string]bool, error) {
	return m.flags, nil
}

func (m *mockRepository) GetLegalConfigs(ctx context.Context) (LegalConfig, error) {
	return m.legal, nil
}

func (m *mockRepository) GetCategories(ctx context.Context) ([]Category, error) {
	return m.categories, nil
}

func (m *mockRepository) GetCurrencies(ctx context.Context) ([]Currency, error) {
	return m.currencies, nil
}

func (m *mockRepository) GetVersionHash(ctx context.Context) (string, error) {
	return m.versionHash, nil
}

func TestGetAppConfig_Unauthenticated_ETagMatch(t *testing.T) {
	mockRepo := &mockRepository{
		versionHash: "v1.0.0-test",
	}
	uc := NewUsecase(mockRepo)

	ctx := context.Background()
	resp, notModified, err := uc.GetAppConfig(ctx, "", "v1.0.0-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !notModified {
		t.Errorf("expected notModified=true for matching ETag")
	}

	if resp != nil {
		t.Errorf("expected nil response on ETag hit")
	}
}

func TestGetAppConfig_Unauthenticated_FreshFetch(t *testing.T) {
	mockRepo := &mockRepository{
		versionHash: "v1.0.0-test",
		categories: []Category{
			{ID: "cat_food", Name: "Food & Dining", IconURL: "https://assets.splittr.app/food.png"},
		},
		currencies: []Currency{
			{Code: "USD", Symbol: "$", Name: "US Dollar", DecimalPlaces: 2, IsDefault: true},
		},
		appVersion: AppVersion{
			MinSupportedVersion: "1.0.0",
			LatestVersion:       "1.2.0",
			ForceUpdate:         false,
			UpdateURL:           map[string]string{"ios": "https://apple.com", "android": "https://google.com"},
			UpdateMessage:       "Update available",
		},
		maintenance: MaintenanceConfig{InMaintenance: false, ReadOnlyMode: false, Message: "All good"},
		limits:      LimitsConfig{MaxExpenseAmount: 10000, MaxGroupMembers: 50, MaxSplitParticipants: 50, MaxReceiptSizeMB: 10, AllowedReceiptMIMEType: []string{"image/jpeg"}},
		flags:       map[string]bool{"enableOcrReceiptScan": true},
		legal:       LegalConfig{TermsOfServiceURL: "https://splittr.app/terms", PrivacyPolicyURL: "https://splittr.app/privacy", FAQURL: "https://splittr.app/faq", SupportEmail: "support@splittr.app"},
	}
	uc := NewUsecase(mockRepo)

	ctx := context.Background()
	resp, notModified, err := uc.GetAppConfig(ctx, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if notModified {
		t.Errorf("expected notModified=false")
	}

	if resp == nil {
		t.Fatalf("expected non-nil response")
	}

	if resp.Meta.ConfigVersion != "v1.0.0-test" {
		t.Errorf("expected ConfigVersion v1.0.0-test, got %s", resp.Meta.ConfigVersion)
	}

	if len(resp.Data.Domain.Categories) != 1 {
		t.Errorf("expected 1 category, got %d", len(resp.Data.Domain.Categories))
	}
}

func TestGetAppConfig_Authenticated_EnrichesUserContext(t *testing.T) {
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

	ctx := context.Background()
	resp, notModified, err := uc.GetAppConfig(ctx, "user_123", "v1.0.0-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if notModified {
		t.Errorf("authenticated request should not return 304 even if ETag matches")
	}

	if resp.Data.UserContext == nil || !resp.Data.UserContext.IsAuthenticated {
		t.Errorf("expected userContext with IsAuthenticated=true")
	}
}
