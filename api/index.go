package handler

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/Saurrabhh/splittr_be/internal/app"
	"github.com/Saurrabhh/splittr_be/internal/config"
	"github.com/Saurrabhh/splittr_be/internal/db"
)

var (
	httpHandler http.Handler
	initOnce    sync.Once
	initErr     error
)

// Handler is the serverless entrypoint for Vercel.
func Handler(w http.ResponseWriter, r *http.Request) {
	initOnce.Do(func() {
		// Create a root context for initialization
		ctx := context.Background()

		// Configure logger
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		logger.Info("initializing vercel serverless function...")

		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			logger.Error("failed to load configuration", "error", err)
			initErr = err
			return
		}

		// Connect to PostgreSQL database (Supabase)
		logger.Info("connecting to database...")
		database, err := db.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			logger.Error("failed to connect to database", "error", err)
			initErr = err
			return
		}
		logger.Info("database connection established")

		// Bootstrap application and mount routes
		application := app.NewApplication(cfg, logger, database)
		h, err := application.Mount(ctx)
		if err != nil {
			logger.Error("failed to mount application routes", "error", err)
			initErr = err
			return
		}

		httpHandler = h
		logger.Info("vercel serverless function initialization complete")
	})

	if initErr != nil {
		http.Error(w, "Internal Server Error: initialization failed", http.StatusInternalServerError)
		return
	}

	// Serve the HTTP request using our mounted chi router
	httpHandler.ServeHTTP(w, r)
}
