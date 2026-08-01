package domain

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/google/uuid"
)

// UseCase manages business logic for activities.
type UseCase struct {
	repo Repository
}

// NewUseCase instantiates a new UseCase.
func NewUseCase(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

// LogEvent records a new activity in the system from a type-safe Event interface.
func (u *UseCase) LogEvent(
	ctx context.Context,
	actorID string,
	groupID *string,
	visibleToUserIDs []string,
	event Event,
) (*Activity, error) {
	var payloadBytes []byte
	if event.Payload() != nil {
		var err error
		payloadBytes, err = json.Marshal(event.Payload())
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
	}

	var entityIDPtr *string
	if event.EntityID() != "" {
		eid := event.EntityID()
		entityIDPtr = &eid
	}

	act := &Activity{
		ID:          uuid.New().String(),
		GroupID:     groupID,
		Actor:       ActorInfo{ID: actorID},
		ActionType:  event.ActionType(),
		EntityType:  event.EntityType(),
		EntityID:    entityIDPtr,
		Description: event.Description(),
		Payload:     event.Payload(),
	}

	if err := u.repo.CreateActivity(ctx, act, payloadBytes); err != nil {
		return nil, fmt.Errorf("create activity: %w", err)
	}

	if groupID == nil || *groupID == "" {
		for _, userID := range visibleToUserIDs {
			if err := u.repo.CreateActivityVisibility(ctx, act.ID, userID); err != nil {
				return nil, fmt.Errorf("create visibility maps: %w", err)
			}
		}
	}

	return act, nil
}

// ListActivities returns a cursor-paginated list of activities visible to the user.
func (u *UseCase) ListActivities(ctx context.Context, userID string, p pagination.Params) (pagination.Response[Activity], error) {
	cursor := pagination.ParseCursor(p.Cursor)
	activities, err := u.repo.ListUserActivities(ctx, userID, p.Limit+1, cursor.LastTime, cursor.LastID)
	if err != nil {
		return pagination.Response[Activity]{}, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogFetchActivityFeed,
			Err:     err,
		}
	}
	return pagination.BuildResponse(activities, p.Limit, func(a Activity) string {
		return pagination.EncodeCursor(a.CreatedAt, a.ID)
	}), nil
}

// GetGroupFeed retrieves group activities with cursor-based pagination.
func (u *UseCase) GetGroupFeed(ctx context.Context, userID, groupID string, p pagination.Params) (pagination.Response[Activity], error) {
	cursor := pagination.ParseCursor(p.Cursor)
	activities, err := u.repo.ListGroupFeed(ctx, groupID, userID, p.Limit+1, cursor.LastTime, cursor.LastID)
	if err != nil {
		return pagination.Response[Activity]{}, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogFetchActivityFeed,
			Err:     err,
		}
	}
	return pagination.BuildResponse(activities, p.Limit, func(a Activity) string {
		return pagination.EncodeCursor(a.CreatedAt, a.ID)
	}), nil
}
