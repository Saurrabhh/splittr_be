package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps the connection pool.
type DB struct {
	Pool     *pgxpool.Pool
	PingFunc func(ctx context.Context) error
}

// Connect initializes the PostgreSQL connection pool.
func Connect(ctx context.Context, connStr string) (*DB, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	// Use QueryExecModeCacheDescribe to support connection pooling/PgBouncer/Supavisor
	// (transaction mode) without prepared statement errors.
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	// Bound the connection and startup ping so an unreachable host fails fast
	// instead of hanging the process indefinitely.
	config.ConnConfig.ConnectTimeout = 10 * time.Second
	// Bound connection lifetime so stale connections are recycled behind poolers.
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the connection pool.
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Ping checks the database connection.
func (db *DB) Ping(ctx context.Context) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}
	if db.PingFunc != nil {
		return db.PingFunc(ctx)
	}
	if db.Pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}
	return db.Pool.Ping(ctx)
}
