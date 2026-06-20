package activity

import (
	"context"
	"fmt"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// DBRepository handles activity transactions in PostgreSQL.
type DBRepository struct {
	db *db.DB
	tm *db.TransactionManager
}

// NewRepository instantiates a new DBRepository.
func NewRepository(database *db.DB, tm *db.TransactionManager) *DBRepository {
	return &DBRepository{
		db: database,
		tm: tm,
	}
}

// CreateActivity logs a new activity.
func (r *DBRepository) CreateActivity(ctx context.Context, act *Activity) error {
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
	if act.ActorID != nil && *act.ActorID != "" {
		aUUID, err := uuid.Parse(*act.ActorID)
		if err != nil {
			return fmt.Errorf("invalid actor uuid: %w", err)
		}
		pgActorID = pgtype.UUID{Bytes: aUUID, Valid: true}
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbAct, err := q.CreateActivity(ctx, dbgen.CreateActivityParams{
		ID:          parsedID,
		GroupID:     pgGroupID,
		ActorID:     pgActorID,
		ActionType:  act.ActionType,
		Description: act.Description,
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

// ListUserActivities lists activities visible to the current user.
func (r *DBRepository) ListUserActivities(ctx context.Context, userID string) ([]Activity, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListUserActivities(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("list user activities: %w", err)
	}

	activities := make([]Activity, 0, len(rows))
	for _, row := range rows {
		var groupIDStr *string
		if row.GroupID.Valid {
			s := uuid.UUID(row.GroupID.Bytes).String()
			groupIDStr = &s
		}

		var actorIDStr *string
		if row.ActorID.Valid {
			s := uuid.UUID(row.ActorID.Bytes).String()
			actorIDStr = &s
		}

		var actorName *string
		if row.ActorName.Valid {
			actorName = &row.ActorName.String
		}

		activities = append(activities, Activity{
			ID:          row.ID.String(),
			GroupID:     groupIDStr,
			ActorID:     actorIDStr,
			ActorName:   actorName,
			ActionType:  row.ActionType,
			Description: row.Description,
			CreatedAt:   row.CreatedAt.Time,
		})
	}

	return activities, nil
}
