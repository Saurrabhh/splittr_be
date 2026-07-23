//go:build integration

package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestDB initializes a Postgres container using testcontainers, applies migrations, and returns a DB client and cleanup function.
func SetupTestDB(ctx context.Context) (*DB, func(), error) {
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

	database, err := Connect(ctx, connStr)
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("connect to database: %w", err)
	}

	// Locate project root to find internal/db/migrations
	dir, err := os.Getwd()
	if err != nil {
		database.Close()
		pgContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("get wd: %w", err)
	}

	var projectRoot string
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			projectRoot = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			database.Close()
			pgContainer.Terminate(ctx)
			return nil, nil, fmt.Errorf("go.mod not found starting from %s", dir)
		}
		dir = parent
	}

	migrationsPath := filepath.Join(projectRoot, "internal", "db", "migrations")
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		database.Close()
		pgContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("read migrations dir %s: %w", migrationsPath, err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			content, err := os.ReadFile(filepath.Join(migrationsPath, file.Name()))
			if err != nil {
				database.Close()
				pgContainer.Terminate(ctx)
				return nil, nil, fmt.Errorf("read migration file %s: %w", file.Name(), err)
			}

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
