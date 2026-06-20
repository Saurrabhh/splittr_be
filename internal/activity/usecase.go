package activity

import (
	"context"
	"fmt"

	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/google/uuid"
)

// Repository defines storage interface for activity domain.
type Repository interface {
	CreateActivity(ctx context.Context, act *Activity) error
	CreateActivityVisibility(ctx context.Context, activityID string, userID string) error
	ListUserActivities(ctx context.Context, userID string) ([]Activity, error)
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
func (u *Usecase) LogActivity(ctx context.Context, actorID string, groupID *string, actionType string, description string, visibleToUserIDs []string) (*Activity, error) {
	newAct := &Activity{
		ID:          uuid.New().String(),
		GroupID:     groupID,
		ActorID:     &actorID,
		ActionType:  actionType,
		Description: description,
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
