//go:build integration

package data_test

import (
	"context"
	"testing"

	"github.com/Saurrabhh/splittr_be/internal/activity/data"
	"github.com/Saurrabhh/splittr_be/internal/activity/domain"
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db_test"
	"github.com/Saurrabhh/splittr_be/internal/group"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestActivityRepo(t *testing.T) (*data.DBRepository, *user.DBRepository, *group.DBRepository, func()) {
	ctx := context.Background()
	testDB, cleanup, err := db_test.SetupTestDB(ctx)
	require.NoError(t, err)

	tm := db.NewTransactionManager(testDB)
	actRepo := data.NewRepository(testDB, tm)
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
	groupAct := &domain.Activity{
		ID:          actID,
		GroupID:     &g.ID,
		Actor:       domain.ActorInfo{ID: actor.ID, Name: actor.Name},
		ActionType:  domain.ActionTypeExpenseCreated,
		Description: "Added flight expense",
		EntityType:  domain.EntityTypeExpense,
	}

	err := actRepo.CreateActivity(ctx, groupAct, nil)
	require.NoError(t, err)
	assert.False(t, groupAct.CreatedAt.IsZero())

	// Non-group activity & visibility creation
	nonGroupActID := uuid.New().String()
	nonGroupAct := &domain.Activity{
		ID:          nonGroupActID,
		Actor:       domain.ActorInfo{ID: actor.ID, Name: actor.Name},
		ActionType:  domain.ActionTypeSettlementCreated,
		Description: "Settled payment",
		EntityType:  domain.EntityTypeSettlement,
	}
	err = actRepo.CreateActivity(ctx, nonGroupAct, nil)
	require.NoError(t, err)

	err = actRepo.CreateActivityVisibility(ctx, nonGroupActID, u1.ID)
	require.NoError(t, err)

	// Invalid UUID errors
	invalidAct := &domain.Activity{ID: "invalid-uuid", ActionType: domain.ActionTypeGroupCreated, Description: "Desc", EntityType: domain.EntityTypeGroup}
	err = actRepo.CreateActivity(ctx, invalidAct, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid activity uuid")

	invalidGroup := "invalid-uuid"
	invalidGroupAct := &domain.Activity{ID: uuid.New().String(), GroupID: &invalidGroup, ActionType: domain.ActionTypeGroupCreated, Description: "Desc", EntityType: domain.EntityTypeGroup}
	err = actRepo.CreateActivity(ctx, invalidGroupAct, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid group uuid")

	invalidActor := "invalid-uuid"
	invalidActorAct := &domain.Activity{ID: uuid.New().String(), Actor: domain.ActorInfo{ID: invalidActor}, ActionType: domain.ActionTypeGroupCreated, Description: "Desc", EntityType: domain.EntityTypeGroup}
	err = actRepo.CreateActivity(ctx, invalidActorAct, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid actor uuid")

	invalidEntity := "invalid-uuid"
	invalidEntityAct := &domain.Activity{ID: uuid.New().String(), EntityID: &invalidEntity, ActionType: domain.ActionTypeGroupCreated, Description: "Desc", EntityType: domain.EntityTypeGroup}
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

	act1 := &domain.Activity{ID: uuid.New().String(), Actor: domain.ActorInfo{ID: u1.ID, Name: u1.Name}, ActionType: domain.ActionTypeSettlementCreated, Description: "Act 1", EntityType: domain.EntityTypeSettlement}
	act2 := &domain.Activity{ID: uuid.New().String(), Actor: domain.ActorInfo{ID: u2.ID, Name: u2.Name}, ActionType: domain.ActionTypeSettlementCreated, Description: "Act 2", EntityType: domain.EntityTypeSettlement}

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
	act1 := &domain.Activity{ID: uuid.New().String(), GroupID: &g.ID, Actor: domain.ActorInfo{ID: admin.ID, Name: admin.Name}, ActionType: domain.ActionTypeGroupCreated, Description: "Created group", EntityType: domain.EntityTypeGroup}
	act2 := &domain.Activity{ID: uuid.New().String(), GroupID: &g.ID, Actor: domain.ActorInfo{ID: member.ID, Name: member.Name}, ActionType: domain.ActionTypeMemberJoined, Description: "Joined group", EntityType: domain.EntityTypeMember}
	act3 := &domain.Activity{ID: uuid.New().String(), GroupID: &g.ID, Actor: domain.ActorInfo{ID: admin.ID, Name: admin.Name}, ActionType: domain.ActionTypeExpenseCreated, Description: "Added gas expense", EntityType: domain.EntityTypeExpense}

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
