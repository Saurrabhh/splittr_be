package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Saurrabhh/splittr_be/internal/app"
	"github.com/Saurrabhh/splittr_be/internal/config"
	"github.com/Saurrabhh/splittr_be/internal/db"
)

func main() {
	// Configure logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to database
	logger.Info("connecting to database...")
	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()
	logger.Info("database connection established")

	// Initialize application
	application := app.NewApplication(cfg, logger, database)

	// Run application
	if err := application.Run(ctx); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}
}
