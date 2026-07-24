package app

import (
	"context"
	"fmt"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/appconfig"
	"github.com/Saurrabhh/splittr_be/internal/auth"
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/expense"
	"github.com/Saurrabhh/splittr_be/internal/group"
	"github.com/Saurrabhh/splittr_be/internal/notification"
	"github.com/Saurrabhh/splittr_be/internal/user"
)

// dependencies holds all wired repository, usecase, and handler instances.
type dependencies struct {
	authMiddleware      *auth.Middleware
	userHandler         *user.Handler
	groupHandler        *group.Handler
	expenseHandler      *expense.Handler
	activityHandler     *activity.Handler
	notificationHandler *notification.Handler
	appConfigHandler    *appconfig.Handler
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

	// AppConfig domain wiring
	appConfigRepo := appconfig.NewRepository(app.DB)
	appConfigUseCase := appconfig.NewUsecase(appConfigRepo)
	appConfigHandler := appconfig.NewHandler(appConfigUseCase)

	// User domain wiring
	userRepo := user.NewRepository(app.DB, tm)
	userUseCase := user.NewUseCase(userRepo)
	userHandler := user.NewHandler(userUseCase)

	// Activity domain wiring
	activityRepo := activity.NewRepository(app.DB, tm)
	activityUseCase := activity.NewUseCase(activityRepo)
	activityHandler := activity.NewHandler(activityUseCase)

	// Notification domain wiring
	notificationRepo := notification.NewRepository(app.DB, tm)
	notificationUseCase := notification.NewUseCase(notificationRepo)
	notificationHandler := notification.NewHandler(notificationUseCase)

	// Group domain wiring
	groupRepo := group.NewRepository(app.DB, tm)
	groupUseCase := group.NewUseCase(groupRepo, tm, activityUseCase, notificationUseCase)
	groupHandler := group.NewHandler(groupUseCase, activityUseCase)

	// Expense domain wiring
	expenseRepo := expense.NewRepository(app.DB, tm)
	expenseUseCase := expense.NewUseCase(expenseRepo, tm, groupUseCase, activityUseCase, notificationUseCase)
	expenseHandler := expense.NewHandler(expenseUseCase)

	return &dependencies{
		authMiddleware:      authMiddleware,
		userHandler:         userHandler,
		groupHandler:        groupHandler,
		expenseHandler:      expenseHandler,
		activityHandler:     activityHandler,
		notificationHandler: notificationHandler,
		appConfigHandler:    appConfigHandler,
	}, nil
}
