package main

import (
	"context"
	"fmt"

	"github.com/Saurrabhh/splittr_be/internal/auth"
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/user"
)

// dependencies holds all wired repository, usecase, and handler instances.
type dependencies struct {
	authMiddleware *auth.Middleware
	userHandler    *user.Handler
}

// initDependencies bootstraps and wires all application dependencies.
func initDependencies(ctx context.Context, app *application) (*dependencies, error) {
	// Initialize transaction manager
	tm := db.NewTransactionManager(app.db)

	// Initialize Firebase Auth verifier and middleware
	app.logger.Info("initializing firebase admin sdk...")
	verifier, err := auth.NewFirebaseVerifier(ctx, app.config.FirebaseProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase: %w", err)
	}
	authMiddleware := auth.NewMiddleware(verifier)

	// User domain wiring
	userRepo := user.NewRepository(app.db, tm)
	userUsecase := user.NewUsecase(userRepo, userRepo)
	userHandler := user.NewHandler(userUsecase)

	return &dependencies{
		authMiddleware: authMiddleware,
		userHandler:    userHandler,
	}, nil
}
