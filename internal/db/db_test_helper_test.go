//go:build integration

package db_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupTestDB(ctx context.Context) (*db.DB, func(), error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("run postgres container: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("get connection string: %w", err)
	}

	database, err := db.Connect(ctx, connStr)
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("connect to database: %w", err)
	}

	// Apply migrations
	migrationsPath := filepath.Join("migrations")
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		database.Close()
		pgContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("read migrations dir: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			content, err := os.ReadFile(filepath.Join(migrationsPath, file.Name()))
			if err != nil {
				database.Close()
				pgContainer.Terminate(ctx)
				return nil, nil, fmt.Errorf("read migration file %s: %w", file.Name(), err)
			}
			
			// Only run Up migration part, ignore Down migration
			sqlContent := string(content)
			if parts := strings.Split(sqlContent, "-- +goose Down"); len(parts) > 0 {
				sqlContent = parts[0]
			}
			
			_, err = database.Pool.Exec(ctx, sqlContent)
			if err != nil {
				database.Close()
				pgContainer.Terminate(ctx)
				return nil, nil, fmt.Errorf("execute migration file %s: %w", file.Name(), err)
			}
		}
	}

	cleanup := func() {
		database.Close()
		pgContainer.Terminate(ctx)
	}

	return database, cleanup, nil
}
