//go:build integration

package db_test

import (
	"context"

	"github.com/Saurrabhh/splittr_be/internal/db"
)

// SetupTestDB delegates to db.SetupTestDB to initialize postgres testcontainer and migrations.
func SetupTestDB(ctx context.Context) (*db.DB, func(), error) {
	return db.SetupTestDB(ctx)
}
