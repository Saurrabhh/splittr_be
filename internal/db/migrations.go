package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// embedMigrationFS is the embedded filesystem holding migrations.
//go:embed migrations/*.sql
var embedMigrationFS embed.FS

// RunMigrations executes pending SQL migrations.
func RunMigrations(ctx context.Context, connStr string) error {
	slog.Info("running database migrations...")

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("open migration db connection: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping migration db: %w", err)
	}

	goose.SetBaseFS(embedMigrationFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	slog.Info("database migrations completed successfully")
	return nil
}
