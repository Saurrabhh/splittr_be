package domain

import (
	"context"
	"time"
)

// Repository defines storage interface for activity domain.
type Repository interface {
	CreateActivity(ctx context.Context, act *Activity, rawPayload []byte) error
	CreateActivityVisibility(ctx context.Context, activityID string, userID string) error
	ListUserActivities(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]Activity, error)
	ListGroupFeed(ctx context.Context, groupID string, userID string, limit int32, lastTime *time.Time, lastID *string) ([]Activity, error)
}
