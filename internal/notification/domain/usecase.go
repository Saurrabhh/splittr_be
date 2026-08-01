package domain

import (
	"context"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/google/uuid"
)

// UseCase manages business logic for notifications.
type UseCase struct {
	repo Repository
}

// NewUseCase instantiates a new UseCase.
func NewUseCase(repo Repository) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

// CreateAlert stores a new notification for a specific recipient user.
func (u *UseCase) CreateAlert(ctx context.Context, userID string, actorID *string, activityID *string, title, content string) (*Notification, error) {
	newNotif := &Notification{
		ID:         uuid.New().String(),
		UserID:     userID,
		ActorID:    actorID,
		ActivityID: activityID,
		Title:      title,
		Content:    content,
	}

	if err := u.repo.CreateNotification(ctx, newNotif); err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to create notification",
			Err:     err,
		}
	}

	return newNotif, nil
}

// ListNotifications returns a cursor-paginated list of notifications for the user.
func (u *UseCase) ListNotifications(ctx context.Context, userID string, p pagination.Params) (pagination.Response[Notification], error) {
	cursor := pagination.ParseCursor(p.Cursor)
	notifs, err := u.repo.ListUserNotifications(ctx, userID, p.Limit+1, cursor.LastTime, cursor.LastID)
	if err != nil {
		return pagination.Response[Notification]{}, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve notifications",
			Err:     err,
		}
	}
	return pagination.BuildResponse(notifs, p.Limit, func(n Notification) string {
		return pagination.EncodeCursor(n.CreatedAt, n.ID)
	}), nil
}

// MarkAsRead marks a single notification as read.
func (u *UseCase) MarkAsRead(ctx context.Context, id, userID string) error {
	if id == "" {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: "notification id is required",
		}
	}
	found, err := u.repo.MarkNotificationAsRead(ctx, id, userID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to mark notification as read",
			Err:     err,
		}
	}
	if !found {
		return &response.AppError{
			Type:    response.TypeNotFound,
			Message: "notification not found",
		}
	}
	return nil
}

// MarkAllAsRead marks all notifications as read for a user.
func (u *UseCase) MarkAllAsRead(ctx context.Context, userID string) error {
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
