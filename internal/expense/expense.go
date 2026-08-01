package expense

import (
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/expense/data"
	"github.com/Saurrabhh/splittr_be/internal/expense/domain"
	"github.com/Saurrabhh/splittr_be/internal/expense/presentation/http"
)

// Domain Type Aliases
type (
	Expense            = domain.Expense
	Split              = domain.Split
	InputSplit         = domain.InputSplit
	UserBalance        = domain.UserBalance
	Settlement         = domain.Settlement
	PairwiseDebt       = domain.PairwiseDebt
	SplitType          = domain.SplitType
	Repository         = domain.Repository
	GroupService       = domain.GroupService
	ActivityLogger     = domain.ActivityLogger
	NotificationSender = domain.NotificationSender
	UseCase            = domain.UseCase
	BalanceResponse    = domain.BalanceResponse
)

// Data Type Aliases
type DBRepository = data.DBRepository

// Presentation DTO & Handler Aliases
type (
	Handler                   = http.Handler
	CreateExpenseResponse     = http.CreateExpenseResponse
	SettleExpenseResponse     = http.SettleExpenseResponse
	GetExpenseDetailsResponse = http.GetExpenseDetailsResponse
	ListExpensesResponse      = http.ListExpensesResponse
)

// Constants
const (
	SplitTypeEqual      = domain.SplitTypeEqual
	SplitTypeExact      = domain.SplitTypeExact
	SplitTypePercentage = domain.SplitTypePercentage
)

// Constructors
func NewRepository(database *db.DB, tm *db.TransactionManager) *DBRepository {
	return data.NewRepository(database, tm)
}

func NewUseCase(
	repo Repository,
	tx db.Transactor,
	groupSvc GroupService,
	activitySvc ActivityLogger,
	notificationSvc NotificationSender,
) *UseCase {
	return domain.NewUseCase(repo, tx, groupSvc, activitySvc, notificationSvc)
}

func NewHandler(uc *UseCase) *Handler {
	return http.NewHandler(uc)
}
