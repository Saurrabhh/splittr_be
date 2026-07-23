package group_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/group"
	"github.com/Saurrabhh/splittr_be/internal/notification"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockGroupRepository struct {
	mock.Mock
}

func (m *mockGroupRepository) GetByID(ctx context.Context, id string) (*group.Group, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*group.Group), args.Error(1)
}

func (m *mockGroupRepository) GetByInviteCode(ctx context.Context, inviteCode string) (*group.Group, error) {
	args := m.Called(ctx, inviteCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*group.Group), args.Error(1)
}

func (m *mockGroupRepository) GetPreviewByInviteCode(ctx context.Context, inviteCode string) (*group.Preview, error) {
	args := m.Called(ctx, inviteCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*group.Preview), args.Error(1)
}

func (m *mockGroupRepository) GetGroupMember(ctx context.Context, groupID, userID string) (*group.Member, error) {
	args := m.Called(ctx, groupID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*group.Member), args.Error(1)
}

func (m *mockGroupRepository) ListGroupMembers(ctx context.Context, groupID string) ([]group.Member, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]group.Member), args.Error(1)
}

func (m *mockGroupRepository) ListUserGroupsWithMembers(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]group.DetailsResponse, error) {
	args := m.Called(ctx, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]group.DetailsResponse), args.Error(1)
}

func (m *mockGroupRepository) CreateGroup(ctx context.Context, g *group.Group) error {
	return m.Called(ctx, g).Error(0)
}

func (m *mockGroupRepository) Update(ctx context.Context, g *group.Group) error {
	return m.Called(ctx, g).Error(0)
}

func (m *mockGroupRepository) Archive(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockGroupRepository) AddGroupMember(ctx context.Context, groupID, userID, role string) error {
	return m.Called(ctx, groupID, userID, role).Error(0)
}

func (m *mockGroupRepository) RemoveGroupMember(ctx context.Context, groupID, userID string) error {
	return m.Called(ctx, groupID, userID).Error(0)
}

func (m *mockGroupRepository) UpdateGroupMemberRole(ctx context.Context, groupID, userID, role string) error {
	return m.Called(ctx, groupID, userID, role).Error(0)
}

type mockActivityLogger struct {
	mock.Mock
}

func (m *mockActivityLogger) LogActivity(
	ctx context.Context,
	actorID string,
	groupID *string,
	actionType activity.ActionType,
	description string,
	visibleToUserIDs []string,
	entityType activity.EntityType,
	entityID string,
	metadata []byte,
) (*activity.Activity, error) {
	args := m.Called(ctx, actorID, groupID, actionType, description, visibleToUserIDs, entityType, entityID, metadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*activity.Activity), args.Error(1)
}

type mockNotificationSender struct {
	mock.Mock
}

func (m *mockNotificationSender) CreateAlert(ctx context.Context, userID string, actorID *string, activityID *string, title, content string) (*notification.Notification, error) {
	args := m.Called(ctx, userID, actorID, activityID, title, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*notification.Notification), args.Error(1)
}

type mockTransactor struct {
	fail bool
}

func (m *mockTransactor) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.fail {
		return errors.New("transaction error")
	}
	return fn(ctx)
}

// --- CreateGroup Tests ---

func TestCreateGroup_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	creatorID := "usr-creator"
	groupName := "Trip to Hawaii"
	groupDesc := "Fun vacation"

	mockRepo.On("CreateGroup", ctx, mock.AnythingOfType("*group.Group")).Return(nil)
	mockRepo.On("AddGroupMember", ctx, mock.AnythingOfType("string"), creatorID, "admin").Return(nil)
	mockRepo.On("ListGroupMembers", ctx, mock.AnythingOfType("string")).Return([]group.Member{
		{GroupID: "grp-1", UserID: creatorID, Role: "admin"},
	}, nil)

	mockAct.On("LogActivity",
		ctx, creatorID, mock.AnythingOfType("*string"), activity.ActionTypeGroupCreated, "created the group",
		([]string)(nil), activity.EntityTypeGroup, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"),
	).Return(&activity.Activity{ID: "act-1"}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	g, err := uc.CreateGroup(ctx, groupName, groupDesc, creatorID)

	require.NoError(t, err)
	assert.NotNil(t, g)
	assert.Equal(t, groupName, g.Name)
	assert.Equal(t, &groupDesc, g.Description)
	assert.Equal(t, &creatorID, g.CreatedBy)
	assert.NotNil(t, g.InviteCode)

	mockRepo.AssertExpectations(t)
	mockAct.AssertExpectations(t)
}

func TestCreateGroup_ValidationErrors(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)

	// Test empty name
	_, err := uc.CreateGroup(ctx, "", "desc", "creator-1")
	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "group name is required")

	// Test empty creator ID
	_, err = uc.CreateGroup(ctx, "Hawaii", "desc", "")
	require.Error(t, err)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "creator ID is required")
}

func TestCreateGroup_RepoCreateError(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	creatorID := "usr-creator"
	mockRepo.On("CreateGroup", ctx, mock.AnythingOfType("*group.Group")).Return(errors.New("db insert error"))

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	_, err := uc.CreateGroup(ctx, "Trip", "", creatorID)
	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeInternal, appErr.Type)
	assert.Contains(t, appErr.Message, "failed to create group")

	mockRepo.AssertExpectations(t)
}

// --- GetGroupDetails Tests ---

func TestGetGroupDetails_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	userID := "usr-1"

	member := &group.Member{GroupID: groupID, UserID: userID, Role: "member"}
	expectedGroup := &group.Group{ID: groupID, Name: "Trip"}
	expectedMembers := []group.Member{*member}

	mockRepo.On("GetGroupMember", ctx, groupID, userID).Return(member, nil)
	mockRepo.On("GetByID", ctx, groupID).Return(expectedGroup, nil)
	mockRepo.On("ListGroupMembers", ctx, groupID).Return(expectedMembers, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	g, members, err := uc.GetGroupDetails(ctx, groupID, userID)

	require.NoError(t, err)
	assert.Equal(t, expectedGroup, g)
	assert.Equal(t, expectedMembers, members)

	mockRepo.AssertExpectations(t)
}

func TestGetGroupDetails_ValidationErrors(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)

	_, _, err := uc.GetGroupDetails(ctx, "", "usr-1")
	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)

	_, _, err = uc.GetGroupDetails(ctx, "grp-1", "")
	require.Error(t, err)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
}

func TestGetGroupDetails_NotAMember(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	userID := "usr-stranger"

	mockRepo.On("GetGroupMember", ctx, groupID, userID).Return(nil, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	_, _, err := uc.GetGroupDetails(ctx, groupID, userID)

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeForbidden, appErr.Type)
	assert.Contains(t, appErr.Message, "not a group member")

	mockRepo.AssertExpectations(t)
}

func TestGetGroupDetails_NotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-nonexistent"
	userID := "usr-1"

	member := &group.Member{GroupID: groupID, UserID: userID, Role: "member"}
	mockRepo.On("GetGroupMember", ctx, groupID, userID).Return(member, nil)
	mockRepo.On("GetByID", ctx, groupID).Return(nil, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	_, _, err := uc.GetGroupDetails(ctx, groupID, userID)

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeNotFound, appErr.Type)

	mockRepo.AssertExpectations(t)
}

// --- ListUserGroups Tests ---

func TestListUserGroups_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	userID := "usr-1"
	expectedGroups := []group.DetailsResponse{
		{Group: group.Group{ID: "grp-1", Name: "Group 1"}},
	}

	mockRepo.On("ListUserGroupsWithMembers", ctx, userID, int32(21), (*time.Time)(nil), (*string)(nil)).Return(expectedGroups, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	res, err := uc.ListUserGroups(ctx, userID, pagination.Params{Limit: 20})

	require.NoError(t, err)
	assert.Len(t, res.Data, 1)
	assert.Equal(t, "grp-1", res.Data[0].Group.ID)

	mockRepo.AssertExpectations(t)
}

func TestListUserGroups_EmptyUserID(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := group.NewUseCase(mockRepo, nil, nil, nil)

	_, err := uc.ListUserGroups(context.Background(), "", pagination.Params{Limit: 20})
	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
}

// --- AddMember Tests ---

func TestAddMember_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	targetUserID := "usr-target"
	actionByUserID := "usr-admin"

	adminMember := &group.Member{GroupID: groupID, UserID: actionByUserID, Role: "admin"}
	g := &group.Group{ID: groupID, Name: "Trip"}
	targetMember := group.Member{GroupID: groupID, UserID: targetUserID, Role: "member"}

	mockRepo.On("GetGroupMember", ctx, groupID, actionByUserID).Return(adminMember, nil)
	mockRepo.On("GetByID", ctx, groupID).Return(g, nil)
	mockRepo.On("AddGroupMember", ctx, groupID, targetUserID, "member").Return(nil)
	mockRepo.On("ListGroupMembers", ctx, groupID).Return([]group.Member{targetMember}, nil)

	act := &activity.Activity{ID: "act-add"}
	mockAct.On("LogActivity",
		ctx, actionByUserID, &groupID, activity.ActionTypeMemberAdded, mock.AnythingOfType("string"),
		([]string)(nil), activity.EntityTypeMember, targetUserID, mock.AnythingOfType("[]uint8"),
	).Return(act, nil)

	mockNotif.On("CreateAlert",
		ctx, targetUserID, &actionByUserID, &act.ID, "Added to Group", mock.AnythingOfType("string"),
	).Return(&notification.Notification{}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	err := uc.AddMember(ctx, groupID, targetUserID, actionByUserID)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAct.AssertExpectations(t)
	mockNotif.AssertExpectations(t)
}

func TestAddMember_NotAdmin(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	targetUserID := "usr-target"
	actionByUserID := "usr-regular"

	regularMember := &group.Member{GroupID: groupID, UserID: actionByUserID, Role: "member"}
	mockRepo.On("GetGroupMember", ctx, groupID, actionByUserID).Return(regularMember, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	err := uc.AddMember(ctx, groupID, targetUserID, actionByUserID)

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeForbidden, appErr.Type)
	assert.Contains(t, appErr.Message, "only admins can add members")

	mockRepo.AssertExpectations(t)
}

func TestAddMember_GroupNotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	targetUserID := "usr-target"
	actionByUserID := "usr-admin"

	adminMember := &group.Member{GroupID: groupID, UserID: actionByUserID, Role: "admin"}
	mockRepo.On("GetGroupMember", ctx, groupID, actionByUserID).Return(adminMember, nil)
	mockRepo.On("GetByID", ctx, groupID).Return(nil, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	err := uc.AddMember(ctx, groupID, targetUserID, actionByUserID)

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeNotFound, appErr.Type)

	mockRepo.AssertExpectations(t)
}

// --- RemoveMember Tests ---

func TestRemoveMember_SelfLeave_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	userID := "usr-member"

	g := &group.Group{ID: groupID, Name: "Trip"}
	member := group.Member{GroupID: groupID, UserID: userID, Role: "member"}
	otherAdmin := group.Member{GroupID: groupID, UserID: "usr-admin", Role: "admin"}
	members := []group.Member{member, otherAdmin}

	mockRepo.On("GetByID", ctx, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, userID).Return(&member, nil)
	mockRepo.On("ListGroupMembers", ctx, groupID).Return(members, nil)
	mockRepo.On("RemoveGroupMember", ctx, groupID, userID).Return(nil)

	mockAct.On("LogActivity",
		ctx, userID, &groupID, activity.ActionTypeMemberLeft, "left the group",
		([]string)(nil), activity.EntityTypeMember, userID, mock.AnythingOfType("[]uint8"),
	).Return(&activity.Activity{ID: "act-leave"}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	err := uc.RemoveMember(ctx, groupID, userID, userID)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAct.AssertExpectations(t)
}

func TestRemoveMember_AdminKicksMember_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	targetUserID := "usr-member"
	actionByUserID := "usr-admin"

	g := &group.Group{ID: groupID, Name: "Trip"}
	adminMember := &group.Member{GroupID: groupID, UserID: actionByUserID, Role: "admin"}
	targetMember := group.Member{GroupID: groupID, UserID: targetUserID, Role: "member"}
	members := []group.Member{*adminMember, targetMember}

	mockRepo.On("GetByID", ctx, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, actionByUserID).Return(adminMember, nil)
	mockRepo.On("ListGroupMembers", ctx, groupID).Return(members, nil)
	mockRepo.On("RemoveGroupMember", ctx, groupID, targetUserID).Return(nil)

	act := &activity.Activity{ID: "act-kick"}
	mockAct.On("LogActivity",
		ctx, actionByUserID, &groupID, activity.ActionTypeMemberKicked, mock.AnythingOfType("string"),
		([]string)(nil), activity.EntityTypeMember, targetUserID, mock.AnythingOfType("[]uint8"),
	).Return(act, nil)

	mockNotif.On("CreateAlert",
		ctx, targetUserID, &actionByUserID, &act.ID, "Removed from Group", mock.AnythingOfType("string"),
	).Return(&notification.Notification{}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	err := uc.RemoveMember(ctx, groupID, targetUserID, actionByUserID)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAct.AssertExpectations(t)
	mockNotif.AssertExpectations(t)
}

func TestRemoveMember_SoleAdminWithOtherMembers(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	adminID := "usr-sole-admin"

	g := &group.Group{ID: groupID, Name: "Trip"}
	adminMember := group.Member{GroupID: groupID, UserID: adminID, Role: "admin"}
	otherMember := group.Member{GroupID: groupID, UserID: "usr-other", Role: "member"}
	members := []group.Member{adminMember, otherMember}

	mockRepo.On("GetByID", ctx, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, adminID).Return(&adminMember, nil)
	mockRepo.On("ListGroupMembers", ctx, groupID).Return(members, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	err := uc.RemoveMember(ctx, groupID, adminID, adminID)

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "cannot remove the sole admin of a group containing other members")

	mockRepo.AssertExpectations(t)
}

func TestRemoveMember_SoleAdminLastMemberArchivesGroup(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	adminID := "usr-sole-admin"

	g := &group.Group{ID: groupID, Name: "Trip"}
	adminMember := group.Member{GroupID: groupID, UserID: adminID, Role: "admin"}
	members := []group.Member{adminMember}

	mockRepo.On("GetByID", ctx, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, adminID).Return(&adminMember, nil)
	mockRepo.On("ListGroupMembers", ctx, groupID).Return(members, nil)
	mockRepo.On("RemoveGroupMember", ctx, groupID, adminID).Return(nil)
	mockRepo.On("Archive", ctx, groupID).Return(nil)

	mockAct.On("LogActivity",
		ctx, adminID, &groupID, activity.ActionTypeMemberLeft, "left the group",
		([]string)(nil), activity.EntityTypeMember, adminID, mock.AnythingOfType("[]uint8"),
	).Return(&activity.Activity{ID: "act-last-leave"}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	err := uc.RemoveMember(ctx, groupID, adminID, adminID)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAct.AssertExpectations(t)
}

func TestRemoveMember_UnauthorizedNonAdmin(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	targetUserID := "usr-target"
	actionByUserID := "usr-regular"

	g := &group.Group{ID: groupID, Name: "Trip"}
	regularMember := &group.Member{GroupID: groupID, UserID: actionByUserID, Role: "member"}

	mockRepo.On("GetByID", ctx, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, actionByUserID).Return(regularMember, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	err := uc.RemoveMember(ctx, groupID, targetUserID, actionByUserID)

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeForbidden, appErr.Type)
	assert.Contains(t, appErr.Message, "only admins can remove other members")

	mockRepo.AssertExpectations(t)
}

func TestRemoveMember_TargetNotMember(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	targetUserID := "usr-stranger"
	actionByUserID := "usr-admin"

	g := &group.Group{ID: groupID, Name: "Trip"}
	adminMember := &group.Member{GroupID: groupID, UserID: actionByUserID, Role: "admin"}

	mockRepo.On("GetByID", ctx, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, actionByUserID).Return(adminMember, nil)
	mockRepo.On("ListGroupMembers", ctx, groupID).Return([]group.Member{*adminMember}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	err := uc.RemoveMember(ctx, groupID, targetUserID, actionByUserID)

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "user is not a member of the group")

	mockRepo.AssertExpectations(t)
}

// --- UpdateMemberRole Tests ---

func TestUpdateMemberRole_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	targetUserID := "usr-target"
	actionByUserID := "usr-admin"

	g := &group.Group{ID: groupID, Name: "Trip"}
	adminMember := &group.Member{GroupID: groupID, UserID: actionByUserID, Role: "admin"}
	targetMember := group.Member{GroupID: groupID, UserID: targetUserID, Role: "member"}
	members := []group.Member{*adminMember, targetMember}

	mockRepo.On("GetByID", ctx, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, actionByUserID).Return(adminMember, nil)
	mockRepo.On("ListGroupMembers", ctx, groupID).Return(members, nil)
	mockRepo.On("UpdateGroupMemberRole", ctx, groupID, targetUserID, "admin").Return(nil)

	act := &activity.Activity{ID: "act-role"}
	mockAct.On("LogActivity",
		ctx, actionByUserID, &groupID, activity.ActionTypeMemberRoleUpdated, mock.AnythingOfType("string"),
		([]string)(nil), activity.EntityTypeMember, targetUserID, mock.AnythingOfType("[]uint8"),
	).Return(act, nil)

	mockNotif.On("CreateAlert",
		ctx, targetUserID, &actionByUserID, &act.ID, "Role Updated", mock.AnythingOfType("string"),
	).Return(&notification.Notification{}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	err := uc.UpdateMemberRole(ctx, groupID, targetUserID, "admin", actionByUserID)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAct.AssertExpectations(t)
	mockNotif.AssertExpectations(t)
}

func TestUpdateMemberRole_InvalidRole(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := group.NewUseCase(mockRepo, nil, nil, nil)

	err := uc.UpdateMemberRole(context.Background(), "grp-1", "usr-2", "superadmin", "usr-1")
	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "invalid role: must be admin or member")
}

func TestUpdateMemberRole_DemoteSoleAdminError(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	adminID := "usr-sole-admin"

	g := &group.Group{ID: groupID, Name: "Trip"}
	adminMember := group.Member{GroupID: groupID, UserID: adminID, Role: "admin"}
	otherMember := group.Member{GroupID: groupID, UserID: "usr-other", Role: "member"}
	members := []group.Member{adminMember, otherMember}

	mockRepo.On("GetByID", ctx, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, adminID).Return(&adminMember, nil)
	mockRepo.On("ListGroupMembers", ctx, groupID).Return(members, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	err := uc.UpdateMemberRole(ctx, groupID, adminID, "member", adminID)

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "cannot demote the sole admin")

	mockRepo.AssertExpectations(t)
}

// --- ArchiveGroup Tests ---

func TestArchiveGroup_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	adminID := "usr-admin"

	g := &group.Group{ID: groupID, Name: "Trip"}
	adminMember := &group.Member{GroupID: groupID, UserID: adminID, Role: "admin"}
	members := []group.Member{*adminMember}

	mockRepo.On("GetGroupMember", ctx, groupID, adminID).Return(adminMember, nil)
	mockRepo.On("GetByID", ctx, groupID).Return(g, nil)
	mockRepo.On("ListGroupMembers", ctx, groupID).Return(members, nil)
	mockRepo.On("Archive", ctx, groupID).Return(nil)

	mockAct.On("LogActivity",
		ctx, adminID, &groupID, activity.ActionTypeGroupArchived, "archived the group",
		([]string)(nil), activity.EntityTypeGroup, groupID, mock.AnythingOfType("[]uint8"),
	).Return(&activity.Activity{ID: "act-archive"}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	err := uc.ArchiveGroup(ctx, groupID, adminID)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAct.AssertExpectations(t)
}

func TestArchiveGroup_NotAdmin(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	userID := "usr-regular"

	regularMember := &group.Member{GroupID: groupID, UserID: userID, Role: "member"}
	mockRepo.On("GetGroupMember", ctx, groupID, userID).Return(regularMember, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	err := uc.ArchiveGroup(ctx, groupID, userID)

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeForbidden, appErr.Type)
	assert.Contains(t, appErr.Message, "only admins can archive")

	mockRepo.AssertExpectations(t)
}

// --- JoinGroup Tests ---

func TestJoinGroup_NewMember_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	inviteCode := "inv-12345"
	userID := "usr-new"
	groupID := "grp-1"

	g := &group.Group{ID: groupID, Name: "Trip", InviteCode: &inviteCode}
	newMember := group.Member{GroupID: groupID, UserID: userID, Role: "member"}

	mockRepo.On("GetByInviteCode", ctx, inviteCode).Return(g, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, userID).Return(nil, nil)
	mockRepo.On("AddGroupMember", ctx, groupID, userID, "member").Return(nil)
	mockRepo.On("ListGroupMembers", ctx, groupID).Return([]group.Member{newMember}, nil)

	mockAct.On("LogActivity",
		ctx, userID, &groupID, activity.ActionTypeMemberJoined, mock.AnythingOfType("string"),
		([]string)(nil), activity.EntityTypeMember, userID, mock.AnythingOfType("[]uint8"),
	).Return(&activity.Activity{ID: "act-join"}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	resGroup, err := uc.JoinGroup(ctx, inviteCode, userID)

	require.NoError(t, err)
	assert.Equal(t, g, resGroup)

	mockRepo.AssertExpectations(t)
	mockAct.AssertExpectations(t)
}

func TestJoinGroup_AlreadyMember(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	inviteCode := "inv-12345"
	userID := "usr-existing"
	groupID := "grp-1"

	g := &group.Group{ID: groupID, Name: "Trip", InviteCode: &inviteCode}
	existingMember := &group.Member{GroupID: groupID, UserID: userID, Role: "member"}

	mockRepo.On("GetByInviteCode", ctx, inviteCode).Return(g, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, userID).Return(existingMember, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	resGroup, err := uc.JoinGroup(ctx, inviteCode, userID)

	require.NoError(t, err)
	assert.Equal(t, g, resGroup)

	mockRepo.AssertExpectations(t)
}

func TestJoinGroup_InvalidInviteCode(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	mockRepo.On("GetByInviteCode", ctx, "invalid-code").Return(nil, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	_, err := uc.JoinGroup(ctx, "invalid-code", "usr-1")

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeNotFound, appErr.Type)

	mockRepo.AssertExpectations(t)
}

// --- GetGroupPreview Tests ---

func TestGetGroupPreview_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	ctx := context.Background()

	inviteCode := "inv-12345"
	expectedPreview := &group.Preview{
		Name:        "Trip",
		MemberCount: 5,
		CreatorName: "Alice",
	}

	mockRepo.On("GetPreviewByInviteCode", ctx, inviteCode).Return(expectedPreview, nil)

	uc := group.NewUseCase(mockRepo, nil, nil, nil)
	preview, err := uc.GetGroupPreview(ctx, inviteCode)

	require.NoError(t, err)
	assert.Equal(t, expectedPreview, preview)

	mockRepo.AssertExpectations(t)
}

func TestGetGroupPreview_EmptyInviteCode(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := group.NewUseCase(mockRepo, nil, nil, nil)

	_, err := uc.GetGroupPreview(context.Background(), "")
	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
}

func TestGetGroupPreview_NotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	ctx := context.Background()

	mockRepo.On("GetPreviewByInviteCode", ctx, "bad-code").Return(nil, nil)

	uc := group.NewUseCase(mockRepo, nil, nil, nil)
	_, err := uc.GetGroupPreview(ctx, "bad-code")

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeNotFound, appErr.Type)

	mockRepo.AssertExpectations(t)
}
