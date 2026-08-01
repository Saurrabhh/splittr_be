package domain

import "time"

type AppVersion struct {
	MinSupportedVersion string            `json:"minSupportedVersion"`
	LatestVersion       string            `json:"latestVersion"`
	ForceUpdate         bool              `json:"forceUpdate"`
	UpdateURL           map[string]string `json:"updateUrl"`
	UpdateMessage       string            `json:"updateMessage"`
}

type MaintenanceConfig struct {
	InMaintenance    bool       `json:"inMaintenance"`
	ReadOnlyMode     bool       `json:"readOnlyMode"`
	Message          string     `json:"message"`
	EstimatedEndTime *time.Time `json:"estimatedEndTime"`
}

type SystemConfig struct {
	AppVersion  AppVersion        `json:"appVersion"`
	Maintenance MaintenanceConfig `json:"maintenance"`
}

type Category struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IconURL string `json:"iconUrl"`
}

type Currency struct {
	Code          string `json:"code"`
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	DecimalPlaces int    `json:"decimalPlaces"`
	IsDefault     bool   `json:"isDefault"`
}

type SplitTypeConfig struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type LimitsConfig struct {
	MaxExpenseAmount       float64  `json:"maxExpenseAmount"`
	MaxGroupMembers        int      `json:"maxGroupMembers"`
	MaxSplitParticipants   int      `json:"maxSplitParticipants"`
	MaxReceiptSizeMB       int      `json:"maxReceiptSizeMb"`
	AllowedReceiptMIMEType []string `json:"allowedReceiptMimeTypes"`
}

type PaymentIntegration struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Enabled        bool   `json:"enabled"`
	DeepLinkScheme string `json:"deepLinkScheme"`
}

type DomainConfig struct {
	Categories          []Category           `json:"categories"`
	Currencies          []Currency           `json:"currencies"`
	SplitTypes          []SplitTypeConfig    `json:"splitTypes"`
	Limits              LimitsConfig         `json:"limits"`
	PaymentIntegrations []PaymentIntegration `json:"paymentIntegrations"`
}

type LegalConfig struct {
	TermsOfServiceURL string `json:"termsOfServiceUrl"`
	PrivacyPolicyURL  string `json:"privacyPolicyUrl"`
	FAQURL            string `json:"faqUrl"`
	SupportEmail      string `json:"supportEmail"`
}

type UserContext struct {
	IsAuthenticated       bool                   `json:"isAuthenticated"`
	UserPreferredCurrency string                 `json:"userPreferredCurrency,omitempty"`
	UserFeatureFlags      map[string]interface{} `json:"userFeatureFlags,omitempty"`
}

type AppConfigData struct {
	System       SystemConfig    `json:"system"`
	Domain       DomainConfig    `json:"domain"`
	FeatureFlags map[string]bool `json:"featureFlags"`
	Legal        LegalConfig     `json:"legal"`
	UserContext  *UserContext    `json:"userContext,omitempty"`
}

type AppConfigMeta struct {
	ConfigVersion string    `json:"configVersion"`
	ServerTime    time.Time `json:"serverTime"`
}

type AppConfigResponse struct {
	Data AppConfigData `json:"data"`
	Meta AppConfigMeta `json:"meta"`
}
