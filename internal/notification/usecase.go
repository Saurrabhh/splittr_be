package notification

import (
	"context"
	"fmt"

	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/google/uuid"
)

// Repository defines storage contract for notifications.
type Repository interface {
	CreateNotification(ctx context.Context, notif *Notification) error
	ListUserNotifications(ctx context.Context, userID string) ([]Notification, error)
	MarkNotificationAsRead(ctx context.Context, id, userID string) error
	MarkAllNotificationsAsRead(ctx context.Context, userID string) error
}

// Usecase manages business logic for notifications.
type Usecase struct {
	repo Repository
}

// NewUsecase instantiates a new Usecase.
func NewUsecase(repo Repository) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

// CreateAlert stores a new notification for a specific recipient user.
func (u *Usecase) CreateAlert(ctx context.Context, userID string, actorID *string, activityID *string, title, content string) (*Notification, error) {
	newNotif := &Notification{
		ID:         uuid.New().String(),
		UserID:     userID,
		ActorID:    actorID,
		ActivityID: activityID,
		Title:      title,
		Content:    content,
	}

	if err := u.repo.CreateNotification(ctx, newNotif); err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	return newNotif, nil
}

// ListNotifications lists all notifications for a user.
func (u *Usecase) ListNotifications(ctx context.Context, userID string) ([]Notification, error) {
	notifs, err := u.repo.ListUserNotifications(ctx, userID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve notifications",
			Err:     err,
		}
	}
	return notifs, nil
}

// MarkAsRead marks a single notification as read.
func (u *Usecase) MarkAsRead(ctx context.Context, id, userID string) error {
	if id == "" {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: "notification id is required",
		}
	}
	err := u.repo.MarkNotificationAsRead(ctx, id, userID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to mark notification as read",
			Err:     err,
		}
	}
	return nil
}

// MarkAllAsRead marks all notifications as read for a user.
func (u *Usecase) MarkAllAsRead(ctx context.Context, userID string) error {
	err := u.repo.MarkAllNotificationsAsRead(ctx, userID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to mark all notifications as read",
			Err:     err,
		}
	}
	return nil
}
