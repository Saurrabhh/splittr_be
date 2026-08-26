package user

import (
	"context"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/storage"
	"github.com/Saurrabhh/splittr_be/internal/user/data"
	"github.com/Saurrabhh/splittr_be/internal/user/domain"
	"github.com/Saurrabhh/splittr_be/internal/user/presentation/http"
)

// Domain Type Aliases
type (
	User       = domain.User
	Repository = domain.Repository
	UseCase    = domain.UseCase
)

// DBRepository Data Type Aliases
type DBRepository = data.DBRepository

// Handler Presentation Type Aliases
type Handler = http.Handler

// WithUser Middleware Helper Functions
func WithUser(ctx context.Context, u *User) context.Context {
	return http.WithUser(ctx, u)
}

func MustFrom(ctx context.Context) *User {
	return http.MustFrom(ctx)
}

// NewRepository Constructors
func NewRepository(database *db.DB, tm *db.TransactionManager) *DBRepository {
	return data.NewRepository(database, tm)
}

func NewUseCase(repo Repository, storageSvc storage.Service) *UseCase {
	return domain.NewUseCase(repo, storageSvc)
}

func NewHandler(uc *UseCase) *Handler {
	return http.NewHandler(uc)
}
