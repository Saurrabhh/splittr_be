package activity_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockActivityRepository struct {
	mock.Mock
}

func (m *mockActivityRepository) CreateActivity(ctx context.Context, act *activity.Activity) error {
	return m.Called(ctx, act).Error(0)
}

func (m *mockActivityRepository) CreateActivityVisibility(ctx context.Context, activityID string, userID string) error {
	return m.Called(ctx, activityID, userID).Error(0)
}

func (m *mockActivityRepository) ListUserActivities(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]activity.Activity, error) {
	args := m.Called(ctx, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]activity.Activity), args.Error(1)
}

func (m *mockActivityRepository) ListGroupFeed(ctx context.Context, groupID string, userID string, limit int32, lastTime *time.Time, lastID *string) ([]activity.Activity, error) {
	args := m.Called(ctx, groupID, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]activity.Activity), args.Error(1)
}

// --- LogActivity Tests ---

func TestLogActivity_GroupActivity_Success(t *testing.T) {
	mockRepo := new(mockActivityRepository)
	ctx := context.Background()

	actorID := "usr-actor"
	groupID := "grp-1"
	actionType := activity.ActionTypeExpenseCreated
	desc := "Created expense for lunch"
	entityType := activity.EntityTypeExpense
	entityID := "exp-1"
	metadata := []byte(`{"amount": 50}`)

	mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*activity.Activity")).Return(nil)

	uc := activity.NewUseCase(mockRepo)
	act, err := uc.LogActivity(ctx, actorID, &groupID, actionType, desc, nil, entityType, entityID, metadata)
	require.NoError(t, err)
	require.NotNil(t, act)

	assert.NotEmpty(t, act.ID)
	assert.Equal(t, &groupID, act.GroupID)
	assert.Equal(t, &actorID, act.ActorID)
	assert.Equal(t, actionType, act.ActionType)
	assert.Equal(t, desc, act.Description)
	assert.Equal(t, entityType, act.EntityType)
	assert.Equal(t, &entityID, act.EntityID)
	assert.Equal(t, json.RawMessage(metadata), act.Metadata)

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "CreateActivityVisibility", mock.Anything, mock.Anything, mock.Anything)
}

func TestLogActivity_NonGroupActivity_WithVisibility_Success(t *testing.T) {
	mockRepo := new(mockActivityRepository)
	ctx := context.Background()

	actorID := "usr-actor"
	actionType := activity.ActionTypeSettlementCreated
	desc := "Settled up balance"
	visibleUsers := []string{"usr-1", "usr-2"}

	mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*activity.Activity")).Return(nil)
	mockRepo.On("CreateActivityVisibility", ctx, mock.AnythingOfType("string"), "usr-1").Return(nil)
	mockRepo.On("CreateActivityVisibility", ctx, mock.AnythingOfType("string"), "usr-2").Return(nil)

	uc := activity.NewUseCase(mockRepo)
	act, err := uc.LogActivity(ctx, actorID, nil, actionType, desc, visibleUsers, activity.EntityTypeSettlement, "settle-1", nil)
	require.NoError(t, err)
	require.NotNil(t, act)

	assert.Nil(t, act.GroupID)
	mockRepo.AssertExpectations(t)
}

func TestLogActivity_CreateActivityError(t *testing.T) {
	mockRepo := new(mockActivityRepository)
	ctx := context.Background()

	mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*activity.Activity")).Return(errors.New("db insert failure"))

	uc := activity.NewUseCase(mockRepo)
	_, err := uc.LogActivity(ctx, "usr-actor", nil, activity.ActionTypeGroupCreated, "Created group", nil, activity.EntityTypeGroup, "grp-1", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create activity: db insert failure")
	mockRepo.AssertExpectations(t)
}

func TestLogActivity_CreateVisibilityError(t *testing.T) {
	mockRepo := new(mockActivityRepository)
	ctx := context.Background()

	mockRepo.On("CreateActivity", ctx, mock.AnythingOfType("*activity.Activity")).Return(nil)
	mockRepo.On("CreateActivityVisibility", ctx, mock.AnythingOfType("string"), "usr-1").Return(errors.New("visibility map failure"))

	uc := activity.NewUseCase(mockRepo)
	_, err := uc.LogActivity(ctx, "usr-actor", nil, activity.ActionTypeSettlementCreated, "Desc", []string{"usr-1"}, activity.EntityTypeSettlement, "s-1", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create visibility maps: visibility map failure")
	mockRepo.AssertExpectations(t)
}

// --- ListActivities (User Feed) Tests ---

func TestListActivities_Success(t *testing.T) {
	mockRepo := new(mockActivityRepository)
	ctx := context.Background()

	now := time.Now()
	expectedActs := []activity.Activity{
		{ID: "act-1", Description: "Activity 1", CreatedAt: now},
		{ID: "act-2", Description: "Activity 2", CreatedAt: now.Add(-time.Minute)},
	}

	mockRepo.On("ListUserActivities", ctx, "usr-1", int32(11), (*time.Time)(nil), (*string)(nil)).Return(expectedActs, nil)

	uc := activity.NewUseCase(mockRepo)
	resp, err := uc.ListActivities(ctx, "usr-1", pagination.Params{Limit: 10})
	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "act-1", resp.Data[0].ID)
	assert.Equal(t, "act-2", resp.Data[1].ID)
	mockRepo.AssertExpectations(t)
}

func TestListActivities_RepoError(t *testing.T) {
	mockRepo := new(mockActivityRepository)
	ctx := context.Background()

	mockRepo.On("ListUserActivities", ctx, "usr-1", int32(11), (*time.Time)(nil), (*string)(nil)).Return(nil, errors.New("db read failure"))

	uc := activity.NewUseCase(mockRepo)
	_, err := uc.ListActivities(ctx, "usr-1", pagination.Params{Limit: 10})
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeInternal, appErr.Type)
	assert.Contains(t, appErr.Message, "failed to retrieve activities")
	mockRepo.AssertExpectations(t)
}

// --- GetGroupFeed Tests ---

func TestGetGroupFeed_Success(t *testing.T) {
	mockRepo := new(mockActivityRepository)
	ctx := context.Background()

	now := time.Now()
	actorName := "Alice"
	actorID := "usr-actor"
	expectedActs := []activity.Activity{
		{
			ID:          "act-1",
			ActorID:     &actorID,
			ActorName:   &actorName,
			ActionType:  activity.ActionTypeExpenseCreated,
			EntityType:  activity.EntityTypeExpense,
			Description: "Created expense",
			CreatedAt:   now,
		},
		{
			ID:          "act-2",
			ActorID:     nil, // System actor
			ActorName:   nil,
			ActionType:  activity.ActionTypeMemberJoined,
			EntityType:  activity.EntityTypeMember,
			Description: "Joined group",
			CreatedAt:   now.Add(-time.Minute),
		},
	}

	mockRepo.On("ListGroupFeed", ctx, "grp-1", "usr-1", int32(11), (*time.Time)(nil), (*string)(nil)).Return(expectedActs, nil)

	uc := activity.NewUseCase(mockRepo)
	resp, err := uc.GetGroupFeed(ctx, "usr-1", "grp-1", pagination.Params{Limit: 10})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Data, 2)

	assert.Equal(t, "usr-actor", resp.Data[0].Actor.ID)
	assert.Equal(t, "Alice", resp.Data[0].Actor.Name)

	assert.Equal(t, "", resp.Data[1].Actor.ID)
	assert.Equal(t, "System", resp.Data[1].Actor.Name)

	mockRepo.AssertExpectations(t)
}

func TestGetGroupFeed_CursorParsing(t *testing.T) {
	mockRepo := new(mockActivityRepository)
	ctx := context.Background()

	expectedTime, _ := time.Parse(time.RFC3339, "2026-07-18T18:00:00Z")
	lastID := "uuid-test"

	mockRepo.On("ListGroupFeed", ctx, "grp-1", "usr-1", int32(11), &expectedTime, &lastID).Return([]activity.Activity{}, nil)

	uc := activity.NewUseCase(mockRepo)
	_, err := uc.GetGroupFeed(ctx, "usr-1", "grp-1", pagination.Params{Limit: 10, Cursor: "2026-07-18T18:00:00Z_uuid-test"})
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetGroupFeed_RepoError(t *testing.T) {
	mockRepo := new(mockActivityRepository)
	ctx := context.Background()

	mockRepo.On("ListGroupFeed", ctx, "grp-1", "usr-1", int32(11), (*time.Time)(nil), (*string)(nil)).Return(nil, errors.New("db error"))

	uc := activity.NewUseCase(mockRepo)
	_, err := uc.GetGroupFeed(ctx, "usr-1", "grp-1", pagination.Params{Limit: 10})
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeInternal, appErr.Type)
	assert.Contains(t, appErr.Message, "failed to retrieve activity feed")
	mockRepo.AssertExpectations(t)
}
