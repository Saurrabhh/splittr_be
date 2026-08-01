package appconfig

import (
	"github.com/Saurrabhh/splittr_be/internal/appconfig/data"
	"github.com/Saurrabhh/splittr_be/internal/appconfig/domain"
	"github.com/Saurrabhh/splittr_be/internal/appconfig/presentation/http"
	"github.com/Saurrabhh/splittr_be/internal/db"
)

// Domain Type Aliases
type (
	AppVersion         = domain.AppVersion
	MaintenanceConfig  = domain.MaintenanceConfig
	SystemConfig       = domain.SystemConfig
	Category           = domain.Category
	Currency           = domain.Currency
	SplitTypeConfig    = domain.SplitTypeConfig
	LimitsConfig       = domain.LimitsConfig
	PaymentIntegration = domain.PaymentIntegration
	DomainConfig       = domain.DomainConfig
	LegalConfig        = domain.LegalConfig
	UserContext        = domain.UserContext
	AppConfigData      = domain.AppConfigData
	AppConfigMeta      = domain.AppConfigMeta
	AppConfigResponse  = domain.AppConfigResponse
	Repository         = domain.Repository
	Usecase            = domain.Usecase
)

// Data Type Aliases
type DBRepository = data.DBRepository

// Presentation Type Aliases
type Handler = http.Handler

// Constructors
func NewRepository(database *db.DB) Repository {
	return data.NewRepository(database)
}

func NewUsecase(repo Repository) Usecase {
	return domain.NewUsecase(repo)
}

func NewHandler(usecase Usecase) *Handler {
	return http.NewHandler(usecase)
}
