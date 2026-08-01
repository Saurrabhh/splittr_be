package data

import (
	"context"
	"fmt"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity/domain"
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// DBRepository handles activity transactions in PostgreSQL.
type DBRepository struct {
	database *db.DB
	tm       *db.TransactionManager
}

// NewRepository instantiates a new DBRepository.
func NewRepository(database *db.DB, tm *db.TransactionManager) *DBRepository {
	return &DBRepository{
		database: database,
		tm:       tm,
	}
}

// CreateActivity logs a new activity.
func (r *DBRepository) CreateActivity(ctx context.Context, act *domain.Activity, rawPayload []byte) error {
	parsedID, err := uuid.Parse(act.ID)
	if err != nil {
		return fmt.Errorf("invalid activity uuid: %w", err)
	}

	var pgGroupID pgtype.UUID
	if act.GroupID != nil && *act.GroupID != "" {
		gUUID, err := uuid.Parse(*act.GroupID)
		if err != nil {
			return fmt.Errorf("invalid group uuid: %w", err)
		}
		pgGroupID = pgtype.UUID{Bytes: gUUID, Valid: true}
	}

	var pgActorID pgtype.UUID
	if act.Actor.ID != "" {
		aUUID, err := uuid.Parse(act.Actor.ID)
		if err != nil {
			return fmt.Errorf("invalid actor uuid: %w", err)
		}
		pgActorID = pgtype.UUID{Bytes: aUUID, Valid: true}
	}

	var pgEntityID pgtype.UUID
	if act.EntityID != nil && *act.EntityID != "" {
		eUUID, err := uuid.Parse(*act.EntityID)
		if err != nil {
			return fmt.Errorf("invalid entity uuid: %w", err)
		}
		pgEntityID = pgtype.UUID{Bytes: eUUID, Valid: true}
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbAct, err := q.CreateActivity(ctx, dbgen.CreateActivityParams{
		ID:          parsedID,
		GroupID:     pgGroupID,
		ActorID:     pgActorID,
		ActionType:  string(act.ActionType),
		Description: act.Description,
		EntityType:  string(act.EntityType),
		EntityID:    pgEntityID,
		Metadata:    rawPayload,
	})
	if err != nil {
		return fmt.Errorf("insert activity: %w", err)
	}

	act.CreatedAt = dbAct.CreatedAt.Time
	return nil
}

// CreateActivityVisibility adds permission mapping for non-group activities.
func (r *DBRepository) CreateActivityVisibility(ctx context.Context, activityID string, userID string) error {
	parsedAct, err := uuid.Parse(activityID)
	if err != nil {
		return fmt.Errorf("invalid activity uuid: %w", err)
	}
	parsedUser, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	return q.CreateActivityVisibility(ctx, dbgen.CreateActivityVisibilityParams{
		ActivityID: parsedAct,
		UserID:     parsedUser,
	})
}

// ListUserActivities lists activities visible to the current user with cursor-based pagination.
func (r *DBRepository) ListUserActivities(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.Activity, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	pgLastTime, lastIDUUID := db.ParseCursor(lastTime, lastID)

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListUserActivitiesPaginated(ctx, dbgen.ListUserActivitiesPaginatedParams{
		UserID:  parsedID,
		Limit:   limit,
		Column3: pgLastTime,
		Column4: lastIDUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list user activities: %w", err)
	}

	activities := make([]domain.Activity, 0, len(rows))
	for _, row := range rows {
		activities = append(activities, mapUserActivityRowToActivity(row))
	}

	return activities, nil
}

// ListGroupFeed queries a group activity feed chronologically with cursor pagination.
func (r *DBRepository) ListGroupFeed(ctx context.Context, groupID string, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.Activity, error) {
	gUUID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group uuid: %w", err)
	}
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user uuid: %w", err)
	}

	pgLastTime, lastIDUUID := db.ParseCursor(lastTime, lastID)

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListGroupFeedPaginated(ctx, dbgen.ListGroupFeedPaginatedParams{
		GroupID: pgtype.UUID{Bytes: gUUID, Valid: true},
		Limit:   limit,
		Column3: pgLastTime,
		Column4: lastIDUUID,
		UserID:  uUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("query list group feed: %w", err)
	}

	activities := make([]domain.Activity, 0, len(rows))
	for _, row := range rows {
		activities = append(activities, mapGroupFeedRowToActivity(row))
	}
	return activities, nil
}
