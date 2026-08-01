package data

import (
	"context"
	"fmt"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/Saurrabhh/splittr_be/internal/notification/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// DBRepository implements postgres database operations for notifications.
type DBRepository struct {
	db *db.DB
	tm *db.TransactionManager
}

// NewRepository creates a new DBRepository instance.
func NewRepository(database *db.DB, tm *db.TransactionManager) *DBRepository {
	return &DBRepository{
		db: database,
		tm: tm,
	}
}

// CreateNotification inserts a notification record.
func (r *DBRepository) CreateNotification(ctx context.Context, notif *domain.Notification) error {
	parsedID, err := uuid.Parse(notif.ID)
	if err != nil {
		return fmt.Errorf("invalid notification uuid: %w", err)
	}

	parsedUserID, err := uuid.Parse(notif.UserID)
	if err != nil {
		return fmt.Errorf("invalid recipient uuid: %w", err)
	}

	var pgActorID pgtype.UUID
	if notif.ActorID != nil && *notif.ActorID != "" {
		aUUID, err := uuid.Parse(*notif.ActorID)
		if err != nil {
			return fmt.Errorf("invalid actor uuid: %w", err)
		}
		pgActorID = pgtype.UUID{Bytes: aUUID, Valid: true}
	}

	var pgActivityID pgtype.UUID
	if notif.ActivityID != nil && *notif.ActivityID != "" {
		actUUID, err := uuid.Parse(*notif.ActivityID)
		if err != nil {
			return fmt.Errorf("invalid activity uuid: %w", err)
		}
		pgActivityID = pgtype.UUID{Bytes: actUUID, Valid: true}
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbNotif, err := q.CreateNotification(ctx, dbgen.CreateNotificationParams{
		ID:         parsedID,
		UserID:     parsedUserID,
		ActorID:    pgActorID,
		ActivityID: pgActivityID,
		Title:      notif.Title,
		Content:    notif.Content,
	})
	if err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}

	notif.CreatedAt = dbNotif.CreatedAt.Time
	notif.IsRead = dbNotif.IsRead
	return nil
}

// ListUserNotifications fetches notifications for a user with cursor-based pagination.
func (r *DBRepository) ListUserNotifications(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.Notification, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	pgLastTime, lastIDUUID := db.ParseCursor(lastTime, lastID)

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListUserNotificationsPaginated(ctx, dbgen.ListUserNotificationsPaginatedParams{
		UserID:  parsedID,
		Limit:   limit,
		Column3: pgLastTime,
		Column4: lastIDUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}

	notifications := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		var actorIDStr *string
		if row.ActorID.Valid {
			actorIDStr = new(uuid.UUID(row.ActorID.Bytes).String())
		}

		var activityIDStr *string
		if row.ActivityID.Valid {
			activityIDStr = new(uuid.UUID(row.ActivityID.Bytes).String())
		}

		var actorName *string
		if row.ActorName.Valid {
			actorName = &row.ActorName.String
		}

		notifications = append(notifications, domain.Notification{
			ID:         row.ID.String(),
			UserID:     row.UserID.String(),
			ActorID:    actorIDStr,
			ActorName:  actorName,
			ActivityID: activityIDStr,
			Title:      row.Title,
			Content:    row.Content,
			IsRead:     row.IsRead,
			CreatedAt:  row.CreatedAt.Time,
		})
	}

	return notifications, nil
}

// MarkNotificationAsRead marks one alert as read.
func (r *DBRepository) MarkNotificationAsRead(ctx context.Context, id, userID string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	return q.MarkNotificationAsRead(ctx, dbgen.MarkNotificationAsReadParams{
		ID:     parsedID,
		UserID: parsedUserID,
	})
}

// MarkAllNotificationsAsRead marks all alerts for a user as read.
func (r *DBRepository) MarkAllNotificationsAsRead(ctx context.Context, userID string) error {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	return q.MarkAllNotificationsAsRead(ctx, parsedUserID)
}
