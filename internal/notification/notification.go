package notification

import (
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/notification/data"
	"github.com/Saurrabhh/splittr_be/internal/notification/domain"
	"github.com/Saurrabhh/splittr_be/internal/notification/presentation/http"
)

// Domain Type Aliases
type (
	Alert        = domain.Alert
	AlertType    = domain.AlertType
	Notification = domain.Notification
	Repository   = domain.Repository
	UseCase      = domain.UseCase
)

// Data Type Aliases
type DBRepository = data.DBRepository

// Presentation Type Aliases
type Handler = http.Handler

// AlertType constants
const (
	AlertTypeExpenseAdded        = domain.AlertTypeExpenseAdded
	AlertTypePaymentReceived     = domain.AlertTypePaymentReceived
	AlertTypeJoinRequestPending  = domain.AlertTypeJoinRequestPending
	AlertTypeJoinRequestApproved = domain.AlertTypeJoinRequestApproved
	AlertTypeJoinRequestRejected = domain.AlertTypeJoinRequestRejected
)

// Type-safe Alert factory functions
var (
	NewExpenseAddedAlert        = domain.NewExpenseAddedAlert
	NewPaymentReceivedAlert     = domain.NewPaymentReceivedAlert
	NewJoinRequestPendingAlert  = domain.NewJoinRequestPendingAlert
	NewJoinRequestApprovedAlert = domain.NewJoinRequestApprovedAlert
	NewJoinRequestRejectedAlert = domain.NewJoinRequestRejectedAlert
)

// Constructors
func NewRepository(database *db.DB, tm *db.TransactionManager) *DBRepository {
	return data.NewRepository(database, tm)
}

func NewUseCase(repo Repository) *UseCase {
	return domain.NewUseCase(repo)
}

func NewHandler(uc *UseCase) *Handler {
	return http.NewHandler(uc)
}
