package user

import (
	"context"

	"github.com/Saurrabhh/splittr_be/internal/db"
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

// Data Type Aliases
type DBRepository = data.DBRepository

// Presentation Type Aliases
type (
	Handler             = http.Handler
	ListFriendsResponse = http.ListFriendsResponse
)

// Middleware Helper Functions
func WithUser(ctx context.Context, u *User) context.Context {
	return http.WithUser(ctx, u)
}

func From(ctx context.Context) *User {
	return http.From(ctx)
}

func MustFrom(ctx context.Context) *User {
	return http.MustFrom(ctx)
}

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
