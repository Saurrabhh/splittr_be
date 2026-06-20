package app

import (
	"context"
	"fmt"

	"github.com/Saurrabhh/splittr_be/internal/auth"
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/expense"
	"github.com/Saurrabhh/splittr_be/internal/group"
	"github.com/Saurrabhh/splittr_be/internal/user"
)

// dependencies holds all wired repository, usecase, and handler instances.
type dependencies struct {
	authMiddleware *auth.Middleware
	userHandler    *user.Handler
	groupHandler   *group.Handler
	expenseHandler *expense.Handler
}

// initDependencies bootstraps and wires all application dependencies.
func initDependencies(ctx context.Context, app *Application) (*dependencies, error) {
	// Initialize transaction manager
	tm := db.NewTransactionManager(app.DB)

	// Initialize Firebase Auth verifier and middleware
	app.Logger.Info("initializing firebase admin sdk...")
	verifier, err := auth.NewFirebaseVerifier(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase: %w", err)
	}
	authMiddleware := auth.NewMiddleware(verifier)

	// User domain wiring
	userRepo := user.NewRepository(app.DB, tm)
	userUsecase := user.NewUsecase(userRepo)
	userHandler := user.NewHandler(userUsecase)

	// Group domain wiring
	groupRepo := group.NewRepository(app.DB, tm)
	groupUsecase := group.NewUsecase(groupRepo, tm)
	groupHandler := group.NewHandler(groupUsecase)

	// Expense domain wiring
	expenseRepo := expense.NewRepository(app.DB, tm)
	expenseUsecase := expense.NewUsecase(expenseRepo, tm, groupUsecase)
	expenseHandler := expense.NewHandler(expenseUsecase)

	return &dependencies{
		authMiddleware: authMiddleware,
		userHandler:    userHandler,
		groupHandler:   groupHandler,
		expenseHandler: expenseHandler,
	}, nil
}
