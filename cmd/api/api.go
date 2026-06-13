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

	"github.com/Saurrabhh/splittr_be/internal/auth"
	"github.com/Saurrabhh/splittr_be/internal/config"
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type application struct {
	config *config.Config
	logger *slog.Logger
	db     *db.DB
}

func (app *application) mount(ctx context.Context) (http.Handler, error) {
	// Initialize transaction manager
	tm := db.NewTransactionManager(app.db)

	// Initialize Firebase Auth
	app.logger.Info("initializing firebase admin sdk...")
	verifier, err := auth.NewFirebaseVerifier(ctx, app.config.FirebaseProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase: %w", err)
	}
	authMiddleware := auth.NewMiddleware(verifier)

	// Wire dependencies manually
	userRepo := user.NewRepository(app.db, tm)
	userUsecase := user.NewUsecase(userRepo, userRepo)
	userHandler := user.NewHandler(userUsecase)

	// Setup routing
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Register routes
	userHandler.RegisterRoutes(r, authMiddleware.Authenticate)

	return r, nil
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
