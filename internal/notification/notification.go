package notification

import (
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/notification/data"
	"github.com/Saurrabhh/splittr_be/internal/notification/domain"
	"github.com/Saurrabhh/splittr_be/internal/notification/presentation/http"
)

// Domain Type Aliases
type (
	Notification = domain.Notification
	Repository   = domain.Repository
	UseCase      = domain.UseCase
)

// Data Type Aliases
type DBRepository = data.DBRepository

// Presentation Type Aliases
type Handler = http.Handler

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
