//go:build integration

package group_test

import (
	"context"
	"testing"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db_test"
	"github.com/Saurrabhh/splittr_be/internal/group"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) (*group.DBRepository, *user.DBRepository, func()) {
	ctx := context.Background()
	testDB, cleanup, err := db_test.SetupTestDB(ctx)
	require.NoError(t, err)

	tm := db.NewTransactionManager(testDB)
	groupRepo := group.NewRepository(testDB, tm)
	userRepo := user.NewRepository(testDB, tm)
	return groupRepo, userRepo, cleanup
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

func TestRepository_CreateGroupAndGetByID(t *testing.T) {
	groupRepo, userRepo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	creator := createTestUser(t, userRepo, "Group Creator")
	groupID := uuid.New().String()
	desc := "Test group description"
	inviteCode := "inv-test-1"

	g := &group.Group{
		ID:          groupID,
		Name:        "Test Group",
		Description: &desc,
		InviteCode:  &inviteCode,
		CreatedBy:   &creator.ID,
	}

	err := groupRepo.CreateGroup(ctx, g)
	require.NoError(t, err)
	assert.False(t, g.CreatedAt.IsZero())
	assert.False(t, g.UpdatedAt.IsZero())
	assert.Nil(t, g.ArchivedAt)

	// GetByID Success
	fetched, err := groupRepo.GetByID(ctx, groupID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, groupID, fetched.ID)
	assert.Equal(t, "Test Group", fetched.Name)
	assert.Equal(t, &desc, fetched.Description)
	assert.Equal(t, &inviteCode, fetched.InviteCode)
	assert.Equal(t, &creator.ID, fetched.CreatedBy)

	// GetByID Invalid UUID
	_, err = groupRepo.GetByID(ctx, "invalid-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid uuid")

	// GetByID Not Found
	notFound, err := groupRepo.GetByID(ctx, uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestRepository_GetByInviteCode_And_Preview(t *testing.T) {
	groupRepo, userRepo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	creator := createTestUser(t, userRepo, "Creator Bob")
	groupID := uuid.New().String()
	desc := "Vacation group"
	inviteCode := "inv-code-999"

	g := &group.Group{
		ID:          groupID,
		Name:        "Vacation",
		Description: &desc,
		InviteCode:  &inviteCode,
		CreatedBy:   &creator.ID,
	}
	require.NoError(t, groupRepo.CreateGroup(ctx, g))
	require.NoError(t, groupRepo.AddGroupMember(ctx, groupID, creator.ID, "admin"))

	// GetByInviteCode Success
	fetched, err := groupRepo.GetByInviteCode(ctx, inviteCode)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, groupID, fetched.ID)
	assert.Equal(t, "Vacation", fetched.Name)

	// GetByInviteCode Empty Code Error
	_, err = groupRepo.GetByInviteCode(ctx, "")
	require.Error(t, err)

	// GetByInviteCode Not Found
	notFound, err := groupRepo.GetByInviteCode(ctx, "non-existent-code")
	require.NoError(t, err)
	assert.Nil(t, notFound)

	// GetPreviewByInviteCode Success
	preview, err := groupRepo.GetPreviewByInviteCode(ctx, inviteCode)
	require.NoError(t, err)
	require.NotNil(t, preview)
	assert.Equal(t, "Vacation", preview.Name)
	assert.Equal(t, &desc, preview.Description)
	assert.Equal(t, int64(1), preview.MemberCount)
	assert.Equal(t, "Creator Bob", preview.CreatorName)

	// GetPreviewByInviteCode Empty Code Error
	_, err = groupRepo.GetPreviewByInviteCode(ctx, "")
	require.Error(t, err)

	// GetPreviewByInviteCode Not Found
	noPreview, err := groupRepo.GetPreviewByInviteCode(ctx, "non-existent-code")
	require.NoError(t, err)
	assert.Nil(t, noPreview)
}

func TestRepository_UpdateGroup(t *testing.T) {
	groupRepo, userRepo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	creator := createTestUser(t, userRepo, "Creator Charlie")
	groupID := uuid.New().String()
	g := &group.Group{
		ID:        groupID,
		Name:      "Original Name",
		CreatedBy: &creator.ID,
	}
	require.NoError(t, groupRepo.CreateGroup(ctx, g))

	newDesc := "Updated Description"
	g.Name = "Updated Name"
	g.Description = &newDesc

	err := groupRepo.Update(ctx, g)
	require.NoError(t, err)

	fetched, err := groupRepo.GetByID(ctx, groupID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", fetched.Name)
	assert.Equal(t, &newDesc, fetched.Description)

	// Update Invalid UUID
	invalidG := &group.Group{ID: "invalid-uuid", Name: "Bad"}
	err = groupRepo.Update(ctx, invalidG)
	require.Error(t, err)
}

func TestRepository_ArchiveGroup(t *testing.T) {
	groupRepo, userRepo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	creator := createTestUser(t, userRepo, "Creator Dave")
	groupID := uuid.New().String()
	g := &group.Group{
		ID:        groupID,
		Name:      "Group to Archive",
		CreatedBy: &creator.ID,
	}
	require.NoError(t, groupRepo.CreateGroup(ctx, g))

	err := groupRepo.Archive(ctx, groupID)
	require.NoError(t, err)

	// GetByID filters out archived groups (archived_at IS NULL)
	fetched, err := groupRepo.GetByID(ctx, groupID)
	require.NoError(t, err)
	assert.Nil(t, fetched)

	// Archive Invalid UUID
	err = groupRepo.Archive(ctx, "invalid-uuid")
	require.Error(t, err)
}

func TestRepository_GroupMembers_CRUD(t *testing.T) {
	groupRepo, userRepo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	u1 := createTestUser(t, userRepo, "Alice")
	u2 := createTestUser(t, userRepo, "Bob")
	groupID := uuid.New().String()
	g := &group.Group{
		ID:        groupID,
		Name:      "Members Test Group",
		CreatedBy: &u1.ID,
	}
	require.NoError(t, groupRepo.CreateGroup(ctx, g))

	// Add Members
	require.NoError(t, groupRepo.AddGroupMember(ctx, groupID, u1.ID, "admin"))
	require.NoError(t, groupRepo.AddGroupMember(ctx, groupID, u2.ID, "member"))

	// GetGroupMember Success
	m1, err := groupRepo.GetGroupMember(ctx, groupID, u1.ID)
	require.NoError(t, err)
	require.NotNil(t, m1)
	assert.Equal(t, "admin", m1.Role)
	assert.Equal(t, groupID, m1.GroupID)
	assert.Equal(t, u1.ID, m1.UserID)

	m2, err := groupRepo.GetGroupMember(ctx, groupID, u2.ID)
	require.NoError(t, err)
	require.NotNil(t, m2)
	assert.Equal(t, "member", m2.Role)

	// GetGroupMember Not Found
	nonMember, err := groupRepo.GetGroupMember(ctx, groupID, uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, nonMember)

	// GetGroupMember Invalid UUID
	_, err = groupRepo.GetGroupMember(ctx, "invalid-uuid", u1.ID)
	require.Error(t, err)

	// ListGroupMembers Success
	members, err := groupRepo.ListGroupMembers(ctx, groupID)
	require.NoError(t, err)
	assert.Len(t, members, 2)

	// UpdateGroupMemberRole Success
	err = groupRepo.UpdateGroupMemberRole(ctx, groupID, u2.ID, "admin")
	require.NoError(t, err)

	updatedM2, err := groupRepo.GetGroupMember(ctx, groupID, u2.ID)
	require.NoError(t, err)
	assert.Equal(t, "admin", updatedM2.Role)

	// UpdateGroupMemberRole Invalid UUID
	err = groupRepo.UpdateGroupMemberRole(ctx, "invalid-uuid", u2.ID, "member")
	require.Error(t, err)

	// RemoveGroupMember Success
	err = groupRepo.RemoveGroupMember(ctx, groupID, u2.ID)
	require.NoError(t, err)

	removedM2, err := groupRepo.GetGroupMember(ctx, groupID, u2.ID)
	require.NoError(t, err)
	assert.Nil(t, removedM2)

	// RemoveGroupMember Invalid UUID
	err = groupRepo.RemoveGroupMember(ctx, "invalid-uuid", u2.ID)
	require.Error(t, err)
}

func TestRepository_ListUserGroupsWithMembers(t *testing.T) {
	groupRepo, userRepo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	u1 := createTestUser(t, userRepo, "User Multi1")
	u2 := createTestUser(t, userRepo, "User Multi2")

	g1ID := uuid.New().String()
	g2ID := uuid.New().String()
	g1 := &group.Group{ID: g1ID, Name: "Group 1", CreatedBy: &u1.ID}
	g2 := &group.Group{ID: g2ID, Name: "Group 2", CreatedBy: &u1.ID}

	require.NoError(t, groupRepo.CreateGroup(ctx, g1))
	require.NoError(t, groupRepo.CreateGroup(ctx, g2))

	require.NoError(t, groupRepo.AddGroupMember(ctx, g1ID, u1.ID, "admin"))
	require.NoError(t, groupRepo.AddGroupMember(ctx, g1ID, u2.ID, "member"))
	require.NoError(t, groupRepo.AddGroupMember(ctx, g2ID, u1.ID, "admin"))

	// List groups for u1
	groupsList, err := groupRepo.ListUserGroupsWithMembers(ctx, u1.ID, 10, nil, nil)
	require.NoError(t, err)
	assert.Len(t, groupsList, 2)

	for _, resp := range groupsList {
		if resp.Group.ID == g1ID {
			assert.Len(t, resp.Members, 2)
		} else if resp.Group.ID == g2ID {
			assert.Len(t, resp.Members, 1)
		}
	}

	// List groups for u2
	u2Groups, err := groupRepo.ListUserGroupsWithMembers(ctx, u2.ID, 10, nil, nil)
	require.NoError(t, err)
	assert.Len(t, u2Groups, 1)
	assert.Equal(t, g1ID, u2Groups[0].Group.ID)

	// Invalid User UUID
	_, err = groupRepo.ListUserGroupsWithMembers(ctx, "invalid-uuid", 10, nil, nil)
	require.Error(t, err)
}

func TestRepository_ForeignKeys_And_ConstraintEdgeCases(t *testing.T) {
	groupRepo, userRepo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	u1 := createTestUser(t, userRepo, "FK User")
	g1ID := uuid.New().String()
	g1 := &group.Group{ID: g1ID, Name: "FK Group", CreatedBy: &u1.ID}
	require.NoError(t, groupRepo.CreateGroup(ctx, g1))

	// 1. Non-existent User ID in AddGroupMember (Foreign Key Constraint)
	nonExistentUserID := uuid.New().String()
	err := groupRepo.AddGroupMember(ctx, g1ID, nonExistentUserID, "member")
	require.Error(t, err, "expected foreign key violation for non-existent user")

	// 2. Non-existent Group ID in AddGroupMember (Foreign Key Constraint)
	nonExistentGroupID := uuid.New().String()
	err = groupRepo.AddGroupMember(ctx, nonExistentGroupID, u1.ID, "member")
	require.Error(t, err, "expected foreign key violation for non-existent group")

	// 3. Duplicate Group Member Insertion is handled via ON CONFLICT DO NOTHING
	require.NoError(t, groupRepo.AddGroupMember(ctx, g1ID, u1.ID, "admin"))
	err = groupRepo.AddGroupMember(ctx, g1ID, u1.ID, "member")
	require.NoError(t, err, "ON CONFLICT DO NOTHING handles duplicate insertion silently")

	members, err := groupRepo.ListGroupMembers(ctx, g1ID)
	require.NoError(t, err)
	assert.Len(t, members, 1)
}
