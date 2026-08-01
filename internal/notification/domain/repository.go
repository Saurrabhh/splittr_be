package domain

import (
	"context"
	"time"
)

// Repository defines storage contract for notifications.
type Repository interface {
	CreateNotification(ctx context.Context, notif *Notification) error
	ListUserNotifications(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]Notification, error)
	MarkNotificationAsRead(ctx context.Context, id, userID string) (bool, error)
	MarkAllNotificationsAsRead(ctx context.Context, userID string) error
}
