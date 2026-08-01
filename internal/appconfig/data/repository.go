package data

import (
	"context"

	"github.com/Saurrabhh/splittr_be/internal/appconfig/domain"
	"github.com/Saurrabhh/splittr_be/internal/db"
)

type DBRepository struct {
	db *db.DB
}

func NewRepository(database *db.DB) domain.Repository {
	return &DBRepository{db: database}
}

func (r *DBRepository) GetAppVersion(ctx context.Context) (domain.AppVersion, error) {
	query := `SELECT min_supported_version, latest_version, force_update, ios_update_url, android_update_url, update_message FROM app_versions WHERE id = 1`
	var v domain.AppVersion
	var iosURL, androidURL string
	err := r.db.Pool.QueryRow(ctx, query).Scan(&v.MinSupportedVersion, &v.LatestVersion, &v.ForceUpdate, &iosURL, &androidURL, &v.UpdateMessage)
	if err != nil {
		return domain.AppVersion{}, err
	}
	v.UpdateURL = map[string]string{
		"ios":     iosURL,
		"android": androidURL,
	}
	return v, nil
}

func (r *DBRepository) GetMaintenanceStatus(ctx context.Context) (domain.MaintenanceConfig, error) {
	query := `SELECT in_maintenance, read_only_mode, message, estimated_end_time FROM maintenance_status WHERE id = 1`
	var m domain.MaintenanceConfig
	err := r.db.Pool.QueryRow(ctx, query).Scan(&m.InMaintenance, &m.ReadOnlyMode, &m.Message, &m.EstimatedEndTime)
	if err != nil {
		return domain.MaintenanceConfig{}, err
	}
	return m, nil
}

func (r *DBRepository) GetSystemLimits(ctx context.Context) (domain.LimitsConfig, error) {
	query := `SELECT max_expense_amount, max_group_members, max_split_participants, max_receipt_size_mb, allowed_receipt_mime_types FROM system_limits WHERE id = 1`
	var l domain.LimitsConfig
	var mimes []string
	err := r.db.Pool.QueryRow(ctx, query).Scan(&l.MaxExpenseAmount, &l.MaxGroupMembers, &l.MaxSplitParticipants, &l.MaxReceiptSizeMB, &mimes)
	if err != nil {
		return domain.LimitsConfig{}, err
	}
	l.AllowedReceiptMIMEType = mimes
	return l, nil
}

func (r *DBRepository) GetFeatureFlags(ctx context.Context) (map[string]bool, error) {
	query := `SELECT key, is_enabled FROM feature_flags`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flags := make(map[string]bool)
	for rows.Next() {
		var key string
		var isEnabled bool
		if err := rows.Scan(&key, &isEnabled); err != nil {
			return nil, err
		}
		flags[key] = isEnabled
	}
	return flags, nil
}

func (r *DBRepository) GetLegalConfigs(ctx context.Context) (domain.LegalConfig, error) {
	query := `SELECT terms_of_service_url, privacy_policy_url, faq_url, support_email FROM legal_configs WHERE id = 1`
	var l domain.LegalConfig
	err := r.db.Pool.QueryRow(ctx, query).Scan(&l.TermsOfServiceURL, &l.PrivacyPolicyURL, &l.FAQURL, &l.SupportEmail)
	if err != nil {
		return domain.LegalConfig{}, err
	}
	return l, nil
}

func (r *DBRepository) GetCategories(ctx context.Context) ([]domain.Category, error) {
	query := `SELECT id, name, icon_url FROM expense_categories WHERE is_active = true ORDER BY display_order ASC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.IconURL); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *DBRepository) GetCurrencies(ctx context.Context) ([]domain.Currency, error) {
	query := `SELECT code, symbol, name, decimal_places, is_default FROM currencies WHERE is_active = true ORDER BY code ASC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []domain.Currency
	for rows.Next() {
		var cur domain.Currency
		if err := rows.Scan(&cur.Code, &cur.Symbol, &cur.Name, &cur.DecimalPlaces, &cur.IsDefault); err != nil {
			return nil, err
		}
		currencies = append(currencies, cur)
	}
	return currencies, nil
}

func (r *DBRepository) GetVersionHash(ctx context.Context) (string, error) {
	query := `SELECT version_hash FROM config_versions WHERE id = 1`
	var versionHash string
	err := r.db.Pool.QueryRow(ctx, query).Scan(&versionHash)
	if err != nil {
		return "v1.0.0-fallback", nil
	}
	return versionHash, nil
}
