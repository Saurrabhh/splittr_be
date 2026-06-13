package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/config"
	"github.com/Saurrabhh/splittr_be/internal/db"
)

type application struct {
	config *config.Config
	logger *slog.Logger
	db     *db.DB
}

func (app *application) mount(ctx context.Context) (http.Handler, error) {
	// Initialize dependencies
	deps, err := initDependencies(ctx, app)
	if err != nil {
		return nil, err
	}

	// Setup and return routing handler
	return app.routes(deps), nil
}

func (app *application) run(ctx context.Context) error {
	handler, err := app.mount(ctx)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:    ":" + app.config.Port,
		Handler: handler,
	}

	// Channel to listen for errors from ListenAndServe
	serverErrorChan := make(chan error, 1)
	go func() {
		app.logger.Info("starting server", "port", app.config.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrorChan <- err
		}
	}()

	// Graceful shutdown listener
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrorChan:
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigChan:
		app.logger.Info("shutting down server...", "signal", sig.String())
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("forced shutdown: %w", err)
		}
	}

	app.logger.Info("server stopped gracefully")
	return nil
}
