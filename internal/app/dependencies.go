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
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/storage"
	"github.com/Saurrabhh/splittr_be/internal/storage/cloudinary"
	"github.com/Saurrabhh/splittr_be/internal/sync"
	"github.com/Saurrabhh/splittr_be/internal/user"
)

// activityLoggerAdapter adapts the activity UseCase to the group domain port.
type activityLoggerAdapter struct {
	uc *activity.UseCase
}

func (a activityLoggerAdapter) LogEvent(ctx context.Context, actorID string, groupID *string, visibleToUserIDs []string, event activity.Event) error {
	_, err := a.uc.LogEvent(ctx, actorID, groupID, visibleToUserIDs, event)
	return err
}

func (a activityLoggerAdapter) GetGroupFeed(ctx context.Context, userID, groupID string, p pagination.Params) (pagination.Response[activity.Activity], error) {
	return a.uc.GetGroupFeed(ctx, userID, groupID, p)
}

// notificationSenderAdapter adapts the notification UseCase to the group/expense domain ports.
type notificationSenderAdapter struct {
	uc *notification.UseCase
}

func (a notificationSenderAdapter) CreateAlert(ctx context.Context, userID string, actorID *string, activityID *string, alert notification.Alert) error {
	_, err := a.uc.CreateAlert(ctx, userID, actorID, activityID, alert)
	return err
}

// dependencies holds all wired repository, usecase, and handler instances.
type dependencies struct {
	authMiddleware      *auth.Middleware
	userHandler         *user.Handler
	groupHandler        *group.Handler
	expenseHandler      *expense.Handler
	activityHandler     *activity.Handler
	notificationHandler *notification.Handler
	appConfigHandler    *appconfig.Handler
	syncHandler         *sync.Handler
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

	// Storage service wiring
	var storageSvc storage.Service
	if app.Config.CloudinaryCloudName != "" && app.Config.CloudinaryAPIKey != "" && app.Config.CloudinaryAPISecret != "" {
		cldProvider, err := cloudinary.New(app.Config.CloudinaryCloudName, app.Config.CloudinaryAPIKey, app.Config.CloudinaryAPISecret)
		if err != nil {
			app.Logger.Warn(fmt.Sprintf("failed to initialize cloudinary storage provider: %v", err))
		} else {
			storageSvc = cldProvider
			app.Logger.Info("cloudinary storage provider initialized successfully")
		}
	} else {
		app.Logger.Warn("cloudinary credentials not configured; image uploads disabled")
	}

	// User domain wiring
	userRepo := user.NewRepository(app.DB, tm)
	userUseCase := user.NewUseCase(userRepo, storageSvc)
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
	groupUseCase := group.NewUseCase(groupRepo, tm, activityLoggerAdapter{activityUseCase}, notificationSenderAdapter{notificationUseCase}, storageSvc)
	groupHandler := group.NewHandler(groupUseCase)

	// Expense domain wiring
	expenseRepo := expense.NewRepository(app.DB, tm)
	expenseUseCase := expense.NewUseCase(expenseRepo, tm, groupUseCase, activityUseCase, notificationSenderAdapter{notificationUseCase})
	expenseHandler := expense.NewHandler(expenseUseCase)

	// Unified Sync wiring
	syncUseCase := sync.NewUseCase(userUseCase, groupUseCase, expenseUseCase)
	syncHandler := sync.NewHandler(syncUseCase)

	return &dependencies{
		authMiddleware:      authMiddleware,
		userHandler:         userHandler,
		groupHandler:        groupHandler,
		expenseHandler:      expenseHandler,
		activityHandler:     activityHandler,
		notificationHandler: notificationHandler,
		appConfigHandler:    appConfigHandler,
		syncHandler:         syncHandler,
	}, nil
}
