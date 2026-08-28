package idempotency

import (
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/idempotency/data"
	"github.com/Saurrabhh/splittr_be/internal/idempotency/domain"
)

type (
	Repository = domain.Repository
	Record     = domain.IdempotencyRecord
)

func NewRepository(database *db.DB, tm *db.TransactionManager) domain.Repository {
	return data.NewRepository(database, tm)
}
