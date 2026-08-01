package data

import (
	"fmt"

	"github.com/Saurrabhh/splittr_be/internal/activity/domain"
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/google/uuid"
)

func mapGroupFeedRowToActivity(row dbgen.ListGroupFeedPaginatedRow) (domain.Activity, error) {
	var groupID *string
	if row.GroupID.Valid {
		s := uuid.UUID(row.GroupID.Bytes).String()
		groupID = &s
	}
	var actorID string
	if row.ActorID.Valid {
		actorID = uuid.UUID(row.ActorID.Bytes).String()
	}
	var entityID *string
	if row.EntityID.Valid {
		s := uuid.UUID(row.EntityID.Bytes).String()
		entityID = &s
	}
	payload, err := domain.UnmarshalPayload(domain.EntityType(row.EntityType), row.Metadata)
	if err != nil {
		return domain.Activity{}, fmt.Errorf("unmarshal activity payload: %w", err)
	}

	return domain.Activity{
		ID:          row.ID.String(),
		GroupID:     groupID,
		Actor:       domain.ActorInfo{ID: actorID, Name: row.ActorName},
		ActionType:  domain.ActionType(row.ActionType),
		EntityType:  domain.EntityType(row.EntityType),
		EntityID:    entityID,
		Description: row.Description,
		Payload:     payload,
		CreatedAt:   row.CreatedAt.Time,
	}, nil
}

func mapUserActivityRowToActivity(row dbgen.ListUserActivitiesPaginatedRow) (domain.Activity, error) {
	var groupID *string
	if row.GroupID.Valid {
		s := uuid.UUID(row.GroupID.Bytes).String()
		groupID = &s
	}
	var actorID string
	if row.ActorID.Valid {
		actorID = uuid.UUID(row.ActorID.Bytes).String()
	}
	var entityID *string
	if row.EntityID.Valid {
		s := uuid.UUID(row.EntityID.Bytes).String()
		entityID = &s
	}
	payload, err := domain.UnmarshalPayload(domain.EntityType(row.EntityType), row.Metadata)
	if err != nil {
		return domain.Activity{}, fmt.Errorf("unmarshal activity payload: %w", err)
	}

	return domain.Activity{
		ID:          row.ID.String(),
		GroupID:     groupID,
		Actor:       domain.ActorInfo{ID: actorID, Name: row.ActorName},
		ActionType:  domain.ActionType(row.ActionType),
		EntityType:  domain.EntityType(row.EntityType),
		EntityID:    entityID,
		Description: row.Description,
		Payload:     payload,
		CreatedAt:   row.CreatedAt.Time,
	}, nil
}
