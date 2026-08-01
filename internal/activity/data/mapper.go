package data

import (
	"github.com/Saurrabhh/splittr_be/internal/activity/domain"
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/google/uuid"
)

func mapGroupFeedRowToActivity(row dbgen.ListGroupFeedPaginatedRow) domain.Activity {
	var groupID *string
	if row.GroupID.Valid {
		groupID = new(uuid.UUID(row.GroupID.Bytes).String())
	}
	var actorID string
	if row.ActorID.Valid {
		actorID = uuid.UUID(row.ActorID.Bytes).String()
	}
	var entityID *string
	if row.EntityID.Valid {
		entityID = new(uuid.UUID(row.EntityID.Bytes).String())
	}
	payload, _ := domain.UnmarshalPayload(domain.EntityType(row.EntityType), row.Metadata)

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
	}
}

func mapUserActivityRowToActivity(row dbgen.ListUserActivitiesPaginatedRow) domain.Activity {
	var groupID *string
	if row.GroupID.Valid {
		groupID = new(uuid.UUID(row.GroupID.Bytes).String())
	}
	var actorID string
	if row.ActorID.Valid {
		actorID = uuid.UUID(row.ActorID.Bytes).String()
	}
	var entityID *string
	if row.EntityID.Valid {
		entityID = new(uuid.UUID(row.EntityID.Bytes).String())
	}
	payload, _ := domain.UnmarshalPayload(domain.EntityType(row.EntityType), row.Metadata)

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
	}
}
