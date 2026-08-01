package domain

import "context"

type Repository interface {
	GetAppVersion(ctx context.Context) (AppVersion, error)
	GetMaintenanceStatus(ctx context.Context) (MaintenanceConfig, error)
	GetSystemLimits(ctx context.Context) (LimitsConfig, error)
	GetFeatureFlags(ctx context.Context) (map[string]bool, error)
	GetLegalConfigs(ctx context.Context) (LegalConfig, error)
	GetCategories(ctx context.Context) ([]Category, error)
	GetCurrencies(ctx context.Context) ([]Currency, error)
	GetVersionHash(ctx context.Context) (string, error)
}
