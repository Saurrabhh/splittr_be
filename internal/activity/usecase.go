package activity

import (
	"context"
	"fmt"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/google/uuid"
)

// Repository defines storage interface for activity domain.
type Repository interface {
	CreateActivity(ctx context.Context, act *Activity) error
	CreateActivityVisibility(ctx context.Context, activityID string, userID string) error
	ListUserActivities(ctx context.Context, userID string) ([]Activity, error)
	ListGroupFeed(ctx context.Context, groupID string, userID string, limit int32, lastTime *time.Time, lastID *string) ([]Activity, error)
}

// Usecase manages business logic for activities.
type Usecase struct {
	repo Repository
}

// NewUsecase instantiates a new Usecase.
func NewUsecase(repo Repository) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

// LogActivity records a new activity in the system.
func (u *Usecase) LogActivity(
	ctx context.Context,
	actorID string,
	groupID *string,
	actionType string,
	description string,
	visibleToUserIDs []string,
	entityType string,
	entityID string,
	metadata []byte,
) (*Activity, error) {
	var entityIDPtr *string
	if entityID != "" {
		entityIDPtr = &entityID
	}

	newAct := &Activity{
		ID:          uuid.New().String(),
		GroupID:     groupID,
		ActorID:     &actorID,
		ActionType:  actionType,
		Description: description,
		EntityType:  entityType,
		EntityID:    entityIDPtr,
		Metadata:    metadata,
	}

	if err := u.repo.CreateActivity(ctx, newAct); err != nil {
		return nil, fmt.Errorf("create activity: %w", err)
	}

	// For non-group activities, restrict visibility to the specified users
	if groupID == nil || *groupID == "" {
		for _, userID := range visibleToUserIDs {
			if err := u.repo.CreateActivityVisibility(ctx, newAct.ID, userID); err != nil {
				return nil, fmt.Errorf("create visibility maps: %w", err)
			}
		}
	}

	return newAct, nil
}

// ListActivities returns all activities visible to a user.
func (u *Usecase) ListActivities(ctx context.Context, userID string) ([]Activity, error) {
	activities, err := u.repo.ListUserActivities(ctx, userID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve activities",
			Err:     err,
		}
	}
	return activities, nil
}

// GetGroupFeed retrieves group activities with cursor-based pagination.
func (u *Usecase) GetGroupFeed(ctx context.Context, userID, groupID string, p pagination.Params) (*FeedResponse, error) {
	// Delegate cursor parsing to the shared package
	cursor := pagination.ParseCursor(p.Cursor)

	// Fetch limit+1 items so BuildResponse can detect a next page
	activities, err := u.repo.ListGroupFeed(ctx, groupID, userID, p.Limit+1, cursor.LastTime, cursor.LastID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve activity feed",
			Err:     err,
		}
	}

	// Map raw Activity rows -> FeedItemResponse DTOs
	feedItems := make([]FeedItemResponse, 0, len(activities))
	for _, act := range activities {
		actorID := ""
		if act.ActorID != nil {
			actorID = *act.ActorID
		}
		actorName := "System"
		if act.ActorName != nil {
			actorName = *act.ActorName
		}
		feedItems = append(feedItems, FeedItemResponse{
			ID:          act.ID,
			EntityType:  act.EntityType,
			EntityID:    act.EntityID,
			ActionType:  act.ActionType,
			Actor:       ActorInfo{ID: actorID, Name: actorName},
			Description: act.Description,
			CreatedAt:   act.CreatedAt,
			Payload:     act.Metadata,
		})
	}

	// Delegate trimming, HasMore, and NextCursor to the shared package
	resp := pagination.BuildResponse(feedItems, p.Limit, func(item FeedItemResponse) string {
		return pagination.EncodeCursor(item.CreatedAt, item.ID)
	})
	return &FeedResponse{
		Data:       resp.Data,
		Pagination: resp.Pagination,
	}, nil
}
