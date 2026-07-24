package appconfig

import (
	"context"
	"sync"
	"time"
)

type Usecase interface {
	GetAppConfig(ctx context.Context, userID string, clientETag string) (*AppConfigResponse, bool, error)
}

type usecase struct {
	repo       Repository
	mu         sync.RWMutex
	cachedResp *AppConfigResponse
	cachedHash string
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) GetAppConfig(ctx context.Context, userID string, clientETag string) (*AppConfigResponse, bool, error) {
	versionHash, err := u.repo.GetVersionHash(ctx)
	if err != nil {
		return nil, false, err
	}

	// Unauthenticated ETag check
	if userID == "" && clientETag != "" && clientETag == versionHash {
		return nil, true, nil // 304 Not Modified
	}

	u.mu.RLock()
	if u.cachedResp != nil && u.cachedHash == versionHash {
		cached := *u.cachedResp
		u.mu.RUnlock()
		cached.Meta.ServerTime = time.Now().UTC()
		if userID != "" {
			cached.Data.UserContext = &UserContext{
				IsAuthenticated:  true,
				UserFeatureFlags: map[string]interface{}{"betaFeaturesEnabled": true},
			}
		}
		return &cached, false, nil
	}
	u.mu.RUnlock()

	// Fetch strongly typed configs from relational tables
	appVersion, err := u.repo.GetAppVersion(ctx)
	if err != nil {
		return nil, false, err
	}

	maintenance, err := u.repo.GetMaintenanceStatus(ctx)
	if err != nil {
		return nil, false, err
	}

	limits, err := u.repo.GetSystemLimits(ctx)
	if err != nil {
		return nil, false, err
	}

	featureFlags, err := u.repo.GetFeatureFlags(ctx)
	if err != nil {
		return nil, false, err
	}

	legal, err := u.repo.GetLegalConfigs(ctx)
	if err != nil {
		return nil, false, err
	}

	categories, err := u.repo.GetCategories(ctx)
	if err != nil {
		return nil, false, err
	}

	currencies, err := u.repo.GetCurrencies(ctx)
	if err != nil {
		return nil, false, err
	}

	splitTypes := []SplitTypeConfig{
		{Code: "EQUAL", Label: "Equally", Description: "Split total amount equally among participants"},
		{Code: "EXACT", Label: "Exact Amounts", Description: "Specify exact amount for each participant"},
		{Code: "PERCENTAGE", Label: "By Percentage", Description: "Specify percentage share for each participant"},
	}

	paymentIntegrations := []PaymentIntegration{
		{ID: "upi", Name: "UPI", Enabled: true, DeepLinkScheme: "upi://"},
		{ID: "paypal", Name: "PayPal", Enabled: true, DeepLinkScheme: "https://paypal.me/"},
		{ID: "venmo", Name: "Venmo", Enabled: true, DeepLinkScheme: "venmo://"},
	}

	resp := &AppConfigResponse{
		Data: AppConfigData{
			System: SystemConfig{
				AppVersion:  appVersion,
				Maintenance: maintenance,
			},
			Domain: DomainConfig{
				Categories:          categories,
				Currencies:          currencies,
				SplitTypes:          splitTypes,
				Limits:              limits,
				PaymentIntegrations: paymentIntegrations,
			},
			FeatureFlags: featureFlags,
			Legal:        legal,
		},
		Meta: AppConfigMeta{
			ConfigVersion: versionHash,
			ServerTime:    time.Now().UTC(),
		},
	}

	u.mu.Lock()
	u.cachedResp = resp
	u.cachedHash = versionHash
	u.mu.Unlock()

	if userID != "" {
		respCopy := *resp
		respCopy.Data.UserContext = &UserContext{
			IsAuthenticated:  true,
			UserFeatureFlags: map[string]interface{}{"betaFeaturesEnabled": true},
		}
		return &respCopy, false, nil
	}

	return resp, false, nil
}
