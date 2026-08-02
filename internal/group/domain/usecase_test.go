package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/group/domain"
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

func (m *mockGroupRepository) GetByID(ctx context.Context, id string) (*domain.Group, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Group), args.Error(1)
}

func (m *mockGroupRepository) GetByInviteCode(ctx context.Context, inviteCode string) (*domain.Group, error) {
	args := m.Called(ctx, inviteCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Group), args.Error(1)
}

func (m *mockGroupRepository) GetPreviewByInviteCode(ctx context.Context, inviteCode string) (*domain.Preview, error) {
	args := m.Called(ctx, inviteCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Preview), args.Error(1)
}

func (m *mockGroupRepository) GetGroupMember(ctx context.Context, groupID, userID string) (*domain.Member, error) {
	args := m.Called(ctx, groupID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Member), args.Error(1)
}

func (m *mockGroupRepository) ListGroupMembers(ctx context.Context, groupID string, status domain.MemberStatus) ([]domain.Member, error) {
	args := m.Called(ctx, groupID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Member), args.Error(1)
}

func (m *mockGroupRepository) ListUserGroupsWithMembers(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.GroupWithMembers, error) {
	args := m.Called(ctx, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.GroupWithMembers), args.Error(1)
}

func (m *mockGroupRepository) CreateGroup(ctx context.Context, g *domain.Group) error {
	return m.Called(ctx, g).Error(0)
}

func (m *mockGroupRepository) Update(ctx context.Context, g *domain.Group) error {
	return m.Called(ctx, g).Error(0)
}

func (m *mockGroupRepository) ResetInviteCode(ctx context.Context, groupID, newInviteCode string, expiresAt time.Time) (*domain.Group, error) {
	args := m.Called(ctx, groupID, newInviteCode, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Group), args.Error(1)
}

func (m *mockGroupRepository) Archive(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockGroupRepository) AddGroupMember(ctx context.Context, groupID, userID string, role domain.MemberRole, status domain.MemberStatus) error {
	return m.Called(ctx, groupID, userID, role, status).Error(0)
}

func (m *mockGroupRepository) UpdateMemberStatus(ctx context.Context, groupID, userID string, status domain.MemberStatus) error {
	return m.Called(ctx, groupID, userID, status).Error(0)
}

func (m *mockGroupRepository) RemoveGroupMember(ctx context.Context, groupID, userID string) error {
	return m.Called(ctx, groupID, userID).Error(0)
}

func (m *mockGroupRepository) UpdateGroupMemberRole(ctx context.Context, groupID, userID string, role domain.MemberRole) error {
	return m.Called(ctx, groupID, userID, role).Error(0)
}

type mockActivityLogger struct {
	mock.Mock
}

func (m *mockActivityLogger) LogEvent(
	ctx context.Context,
	actorID string,
	groupID *string,
	visibleToUserIDs []string,
	event activity.Event,
) error {
	args := m.Called(ctx, actorID, groupID, visibleToUserIDs, event)
	return args.Error(0)
}

func (m *mockActivityLogger) GetGroupFeed(ctx context.Context, userID, groupID string, p pagination.Params) (pagination.Response[activity.Activity], error) {
	args := m.Called(ctx, userID, groupID, p)
	if args.Get(0) == nil {
		return pagination.Response[activity.Activity]{}, args.Error(1)
	}
	return args.Get(0).(pagination.Response[activity.Activity]), args.Error(1)
}

type mockNotificationSender struct {
	mock.Mock
}

func (m *mockNotificationSender) CreateAlert(ctx context.Context, userID string, actorID *string, activityID *string, alert notification.Alert) error {
	args := m.Called(ctx, userID, actorID, activityID, alert)
	return args.Error(0)
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

	mockRepo.On("CreateGroup", ctx, mock.AnythingOfType("*domain.Group")).Return(nil)
	mockRepo.On("AddGroupMember", ctx, mock.AnythingOfType("string"), creatorID, domain.MemberRoleAdmin, domain.MemberStatusActive).Return(nil)
	mockRepo.On("ListGroupMembers", ctx, mock.AnythingOfType("string"), domain.MemberStatusActive).Return([]domain.Member{
		{GroupID: "grp-1", UserID: creatorID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive},
	}, nil)

	mockAct.On("LogEvent",
		mock.Anything, creatorID, mock.Anything, ([]string)(nil), mock.Anything,
	).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	g, err := uc.CreateGroup(ctx, groupName, groupDesc, false, creatorID)

	require.NoError(t, err)
	assert.NotNil(t, g)
	assert.Equal(t, groupName, g.Name)
	assert.Equal(t, &groupDesc, g.Description)

	time.Sleep(10 * time.Millisecond)

	mockRepo.AssertExpectations(t)
	mockAct.AssertExpectations(t)
}

func TestCreateGroup_EmptyName(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	_, err := uc.CreateGroup(ctx, "", "desc", false, "usr-1")

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
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

	activeMember := &domain.Member{GroupID: groupID, UserID: userID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}
	expectedGroup := &domain.Group{ID: groupID, Name: "Flatmates"}
	expectedMembers := []domain.Member{*activeMember}

	mockRepo.On("GetGroupMember", ctx, groupID, userID).Return(activeMember, nil)
	mockRepo.On("GetByID", ctx, groupID).Return(expectedGroup, nil)
	mockRepo.On("ListGroupMembers", ctx, groupID, domain.MemberStatusActive).Return(expectedMembers, nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	g, err := uc.GetGroupDetails(ctx, groupID, userID)

	require.NoError(t, err)
	assert.Equal(t, expectedGroup, g)
	assert.Equal(t, expectedMembers, g.Members)

	mockRepo.AssertExpectations(t)
}

func TestGetGroupDetails_NotActiveMember_Forbidden(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	userID := "usr-1"

	pendingMember := &domain.Member{GroupID: groupID, UserID: userID, Role: domain.MemberRoleMember, Status: domain.MemberStatusPending}
	mockRepo.On("GetGroupMember", ctx, groupID, userID).Return(pendingMember, nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	_, err := uc.GetGroupDetails(ctx, groupID, userID)

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeForbidden, appErr.Type)
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
	futureExpires := time.Now().Add(1 * time.Hour)

	g := &domain.Group{ID: groupID, Name: "Trip", InviteCode: &inviteCode, InviteCodeExpiresAt: &futureExpires, RequireAdminApproval: false}
	newMember := domain.Member{GroupID: groupID, UserID: userID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}

	mockRepo.On("GetByInviteCode", ctx, inviteCode).Return(g, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, userID).Return(nil, nil)
	mockRepo.On("AddGroupMember", ctx, groupID, userID, domain.MemberRoleMember, domain.MemberStatusActive).Return(nil)
	mockRepo.On("ListGroupMembers", ctx, groupID, domain.MemberStatusActive).Return([]domain.Member{newMember}, nil)

	mockAct.On("LogEvent",
		ctx, userID, &groupID, ([]string)(nil), mock.Anything,
	).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	resp, err := uc.JoinGroup(ctx, inviteCode, userID)

	require.NoError(t, err)
	assert.Equal(t, domain.MemberStatusActive, resp.Status)
	assert.Equal(t, g, resp.Group)

	mockRepo.AssertExpectations(t)
	mockAct.AssertExpectations(t)
}

func TestJoinGroup_ExpiredInviteCode(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	inviteCode := "inv-expired"
	pastExpires := time.Now().Add(-1 * time.Hour)
	g := &domain.Group{ID: "grp-1", Name: "Trip", InviteCode: &inviteCode, InviteCodeExpiresAt: &pastExpires}

	mockRepo.On("GetByInviteCode", ctx, inviteCode).Return(g, nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	_, err := uc.JoinGroup(ctx, inviteCode, "usr-1")

	require.Error(t, err)
	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "expired")
}

func TestJoinGroup_RequireAdminApproval_Pending(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	inviteCode := "inv-approval"
	userID := "usr-pending"
	groupID := "grp-approval"
	futureExpires := time.Now().Add(1 * time.Hour)
	adminID := "usr-admin"

	g := &domain.Group{ID: groupID, Name: "Strict Group", InviteCode: &inviteCode, InviteCodeExpiresAt: &futureExpires, RequireAdminApproval: true}

	mockRepo.On("GetByInviteCode", ctx, inviteCode).Return(g, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, userID).Return(nil, nil)
	mockRepo.On("AddGroupMember", ctx, groupID, userID, domain.MemberRoleMember, domain.MemberStatusPending).Return(nil)
	mockRepo.On("ListGroupMembers", ctx, groupID, domain.MemberStatusActive).Return([]domain.Member{
		{GroupID: groupID, UserID: adminID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive},
	}, nil)

	mockNotif.On("CreateAlert", ctx, adminID, &userID, (*string)(nil), mock.Anything).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	resp, err := uc.JoinGroup(ctx, inviteCode, userID)

	require.NoError(t, err)
	assert.Equal(t, domain.MemberStatusPending, resp.Status)
	assert.Contains(t, resp.Message, "admin approval")
}

// --- DecideJoinRequest Tests ---

func TestDecideJoinRequest_Approve(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	targetUserID := "usr-pending"
	adminUserID := "usr-admin"

	adminMember := &domain.Member{GroupID: groupID, UserID: adminUserID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive}
	approvedMember := &domain.Member{GroupID: groupID, UserID: targetUserID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}

	mockRepo.On("GetByID", ctx, groupID).Return(&domain.Group{ID: groupID}, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, adminUserID).Return(adminMember, nil)
	mockRepo.On("UpdateMemberStatus", ctx, groupID, targetUserID, domain.MemberStatusActive).Return(nil)
	mockRepo.On("GetGroupMember", ctx, groupID, targetUserID).Return(approvedMember, nil)

	mockAct.On("LogEvent",
		ctx, adminUserID, &groupID, ([]string)(nil), mock.Anything,
	).Return(nil)

	mockNotif.On("CreateAlert", ctx, targetUserID, &adminUserID, (*string)(nil), mock.Anything).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	res, err := uc.DecideJoinRequest(ctx, groupID, targetUserID, "APPROVE", adminUserID)

	require.NoError(t, err)
	assert.NotNil(t, res)
	mockRepo.AssertExpectations(t)
}

// --- ResetInviteCode Tests ---

func TestResetInviteCode_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}
	ctx := context.Background()

	groupID := "grp-1"
	adminUserID := "usr-admin"

	adminMember := &domain.Member{GroupID: groupID, UserID: adminUserID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive}
	updatedGroup := &domain.Group{ID: groupID, Name: "Trip"}

	mockRepo.On("GetByID", ctx, groupID).Return(updatedGroup, nil)
	mockRepo.On("GetGroupMember", ctx, groupID, adminUserID).Return(adminMember, nil)
	mockRepo.On("ResetInviteCode", ctx, groupID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(updatedGroup, nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	res, err := uc.ResetInviteCode(ctx, groupID, adminUserID)

	require.NoError(t, err)
	assert.Equal(t, updatedGroup, res)
}
