package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// txKey is the context key for the active transaction.
type txKey struct{}

// DBTX defines the interface required to execute queries.
// It is implemented by both *pgxpool.Pool and pgx.Tx.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// TransactionManager handles database transaction boundaries.
type TransactionManager struct {
	db *DB
}

// NewTransactionManager creates a new TransactionManager.
func NewTransactionManager(db *DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// RunInTx runs the callback function inside a transaction.
// If the callback returns an error, the transaction is rolled back.
func (tm *TransactionManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	// If a transaction is already present, reuse it.
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := tm.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(context.Background())
			panic(p)
		}
	}()

	ctxWithTx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(ctxWithTx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback error: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetTxOrPool retrieves the active transaction from context, or returns the db connection pool.
func (tm *TransactionManager) GetTxOrPool(ctx context.Context) DBTX {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return tm.db.Pool
}
