//go:build integration

package activity_test

import (
	"context"
	"testing"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db_test"
	"github.com/Saurrabhh/splittr_be/internal/group"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestActivityRepo(t *testing.T) (*activity.DBRepository, *user.DBRepository, *group.DBRepository, func()) {
	ctx := context.Background()
	testDB, cleanup, err := db_test.SetupTestDB(ctx)
	require.NoError(t, err)

	tm := db.NewTransactionManager(testDB)
	actRepo := activity.NewRepository(testDB, tm)
	userRepo := user.NewRepository(testDB, tm)
	groupRepo := group.NewRepository(testDB, tm)

	return actRepo, userRepo, groupRepo, cleanup
}

func createTestUser(t *testing.T, userRepo *user.DBRepository, name string) *user.User {
	ctx := context.Background()
	email := uuid.New().String() + "@example.com"
	u := &user.User{
		ID:          uuid.New().String(),
		FirebaseUID: "fb-" + uuid.New().String(),
		Email:       &email,
		Name:        name,
	}
	err := userRepo.Create(ctx, u)
	require.NoError(t, err)
	return u
}

func createTestGroup(t *testing.T, groupRepo *group.DBRepository, creatorID string, name string) *group.Group {
	ctx := context.Background()
	g := &group.Group{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedBy: &creatorID,
	}
	err := groupRepo.CreateGroup(ctx, g)
	require.NoError(t, err)
	err = groupRepo.AddGroupMember(ctx, g.ID, creatorID, group.MemberRoleAdmin, group.MemberStatusActive)
	require.NoError(t, err)
	return g
}

func TestRepository_CreateActivity_And_Visibility(t *testing.T) {
	actRepo, userRepo, groupRepo, cleanup := setupTestActivityRepo(t)
	defer cleanup()
	ctx := context.Background()

	actor := createTestUser(t, userRepo, "Actor Alice")
	u1 := createTestUser(t, userRepo, "User Bob")
	g := createTestGroup(t, groupRepo, actor.ID, "Vacation Group")

	// Group activity creation
	actID := uuid.New().String()
	groupAct := &activity.Activity{
		ID:          actID,
		GroupID:     &g.ID,
		Actor:       activity.ActorInfo{ID: actor.ID, Name: actor.Name},
		ActionType:  activity.ActionTypeExpenseCreated,
		Description: "Added flight expense",
		EntityType:  activity.EntityTypeExpense,
	}

	err := actRepo.CreateActivity(ctx, groupAct, nil)
	require.NoError(t, err)
	assert.False(t, groupAct.CreatedAt.IsZero())

	// Non-group activity & visibility creation
	nonGroupActID := uuid.New().String()
	nonGroupAct := &activity.Activity{
		ID:          nonGroupActID,
		Actor:       activity.ActorInfo{ID: actor.ID, Name: actor.Name},
		ActionType:  activity.ActionTypeSettlementCreated,
		Description: "Settled payment",
		EntityType:  activity.EntityTypeSettlement,
	}
	err = actRepo.CreateActivity(ctx, nonGroupAct, nil)
	require.NoError(t, err)

	err = actRepo.CreateActivityVisibility(ctx, nonGroupActID, u1.ID)
	require.NoError(t, err)

	// Invalid UUID errors
	invalidAct := &activity.Activity{ID: "invalid-uuid", ActionType: activity.ActionTypeGroupCreated, Description: "Desc", EntityType: activity.EntityTypeGroup}
	err = actRepo.CreateActivity(ctx, invalidAct, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid activity uuid")

	invalidGroup := "invalid-uuid"
	invalidGroupAct := &activity.Activity{ID: uuid.New().String(), GroupID: &invalidGroup, ActionType: activity.ActionTypeGroupCreated, Description: "Desc", EntityType: activity.EntityTypeGroup}
	err = actRepo.CreateActivity(ctx, invalidGroupAct, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid group uuid")

	invalidActor := "invalid-uuid"
	invalidActorAct := &activity.Activity{ID: uuid.New().String(), Actor: activity.ActorInfo{ID: invalidActor}, ActionType: activity.ActionTypeGroupCreated, Description: "Desc", EntityType: activity.EntityTypeGroup}
	err = actRepo.CreateActivity(ctx, invalidActorAct, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid actor uuid")

	invalidEntity := "invalid-uuid"
	invalidEntityAct := &activity.Activity{ID: uuid.New().String(), EntityID: &invalidEntity, ActionType: activity.ActionTypeGroupCreated, Description: "Desc", EntityType: activity.EntityTypeGroup}
	err = actRepo.CreateActivity(ctx, invalidEntityAct, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid entity uuid")

	// Invalid visibility UUIDs
	err = actRepo.CreateActivityVisibility(ctx, "invalid-uuid", u1.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid activity uuid")

	err = actRepo.CreateActivityVisibility(ctx, nonGroupActID, "invalid-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user uuid")
}

func TestRepository_ListUserActivities(t *testing.T) {
	actRepo, userRepo, _, cleanup := setupTestActivityRepo(t)
	defer cleanup()
	ctx := context.Background()

	u1 := createTestUser(t, userRepo, "User One")
	u2 := createTestUser(t, userRepo, "User Two")

	act1 := &activity.Activity{ID: uuid.New().String(), Actor: activity.ActorInfo{ID: u1.ID, Name: u1.Name}, ActionType: activity.ActionTypeSettlementCreated, Description: "Act 1", EntityType: activity.EntityTypeSettlement}
	act2 := &activity.Activity{ID: uuid.New().String(), Actor: activity.ActorInfo{ID: u2.ID, Name: u2.Name}, ActionType: activity.ActionTypeSettlementCreated, Description: "Act 2", EntityType: activity.EntityTypeSettlement}

	require.NoError(t, actRepo.CreateActivity(ctx, act1, nil))
	require.NoError(t, actRepo.CreateActivity(ctx, act2, nil))

	require.NoError(t, actRepo.CreateActivityVisibility(ctx, act1.ID, u1.ID))
	require.NoError(t, actRepo.CreateActivityVisibility(ctx, act2.ID, u2.ID))

	// List for User 1
	list1, err := actRepo.ListUserActivities(ctx, u1.ID, 10, nil, nil)
	require.NoError(t, err)
	require.Len(t, list1, 1)
	assert.Equal(t, act1.ID, list1[0].ID)

	// Test invalid User UUID
	_, err = actRepo.ListUserActivities(ctx, "invalid-uuid", 10, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid uuid")
}

func TestRepository_ListGroupFeed_And_Pagination(t *testing.T) {
	actRepo, userRepo, groupRepo, cleanup := setupTestActivityRepo(t)
	defer cleanup()
	ctx := context.Background()

	admin := createTestUser(t, userRepo, "Admin Alice")
	member := createTestUser(t, userRepo, "Member Bob")
	g := createTestGroup(t, groupRepo, admin.ID, "Road Trip")

	require.NoError(t, groupRepo.AddGroupMember(ctx, g.ID, member.ID, group.MemberRoleMember, group.MemberStatusActive))

	// Create 3 group activities
	act1 := &activity.Activity{ID: uuid.New().String(), GroupID: &g.ID, Actor: activity.ActorInfo{ID: admin.ID, Name: admin.Name}, ActionType: activity.ActionTypeGroupCreated, Description: "Created group", EntityType: activity.EntityTypeGroup}
	act2 := &activity.Activity{ID: uuid.New().String(), GroupID: &g.ID, Actor: activity.ActorInfo{ID: member.ID, Name: member.Name}, ActionType: activity.ActionTypeMemberJoined, Description: "Joined group", EntityType: activity.EntityTypeMember}
	act3 := &activity.Activity{ID: uuid.New().String(), GroupID: &g.ID, Actor: activity.ActorInfo{ID: admin.ID, Name: admin.Name}, ActionType: activity.ActionTypeExpenseCreated, Description: "Added gas expense", EntityType: activity.EntityTypeExpense}

	require.NoError(t, actRepo.CreateActivity(ctx, act1, nil))
	require.NoError(t, actRepo.CreateActivity(ctx, act2, nil))
	require.NoError(t, actRepo.CreateActivity(ctx, act3, nil))

	// Fetch page 1 (limit 2)
	feed1, err := actRepo.ListGroupFeed(ctx, g.ID, member.ID, 2, nil, nil)
	require.NoError(t, err)
	require.Len(t, feed1, 2)

	// Fetch page 2 using cursor from last item in feed1
	lastItem := feed1[1]
	feed2, err := actRepo.ListGroupFeed(ctx, g.ID, member.ID, 2, &lastItem.CreatedAt, &lastItem.ID)
	require.NoError(t, err)
	require.Len(t, feed2, 1)

	// Test invalid UUIDs
	_, err = actRepo.ListGroupFeed(ctx, "invalid-uuid", member.ID, 10, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid group uuid")

	_, err = actRepo.ListGroupFeed(ctx, g.ID, "invalid-uuid", 10, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user uuid")
}
