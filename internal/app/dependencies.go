package app

import (
	"context"
	"fmt"
	"os"

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
func initDependencies(ctx context.Context, app *Application) (*dependencies, error) {
	// Initialize transaction manager
	tm := db.NewTransactionManager(app.DB)

	// Write Firebase credentials to file if supplied via environment variable
	firebaseKeyJSON := os.Getenv("FIREBASE_KEY_JSON")
	if firebaseKeyJSON != "" {
		credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if credentialsPath == "" || credentialsPath == "./firebase-key.json" {
			credentialsPath = "/tmp/firebase-key.json"
			os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsPath)
		}
		app.Logger.Info("writing firebase credentials from env var...", "path", credentialsPath)
		if err := os.WriteFile(credentialsPath, []byte(firebaseKeyJSON), 0600); err != nil {
			return nil, fmt.Errorf("failed to write firebase credentials: %w", err)
		}
	}

	// Initialize Firebase Auth verifier and middleware
	app.Logger.Info("initializing firebase admin sdk...")
	verifier, err := auth.NewFirebaseVerifier(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase: %w", err)
	}
	authMiddleware := auth.NewMiddleware(verifier)

	// User domain wiring
	userRepo := user.NewRepository(app.DB, tm)
	userUsecase := user.NewUsecase(userRepo, userRepo)
	userHandler := user.NewHandler(userUsecase)

	return &dependencies{
		authMiddleware: authMiddleware,
		userHandler:    userHandler,
	}, nil
}
