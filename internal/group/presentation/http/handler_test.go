package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/group/domain"
	grouphttp "github.com/Saurrabhh/splittr_be/internal/group/presentation/http"
	"github.com/Saurrabhh/splittr_be/internal/notification"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
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


func (m *mockGroupRepository) AddGroupMembers(ctx context.Context, groupID string, userIDs []string, role domain.MemberRole, status domain.MemberStatus) ([]domain.Member, error) {
	args := m.Called(ctx, groupID, userIDs, role, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Member), args.Error(1)
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

func (m *mockGroupRepository) SyncGroupsBySequence(ctx context.Context, lastVersion int64, userID string, limit int32) ([]domain.Group, error) {
	args := m.Called(ctx, lastVersion, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Group), args.Error(1)
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

type mockActivityRepo struct {
	mock.Mock
}

func (m *mockActivityRepo) CreateActivity(ctx context.Context, act *activity.Activity, rawPayload []byte) error {
	return m.Called(ctx, act, rawPayload).Error(0)
}

func (m *mockActivityRepo) CreateActivityVisibility(ctx context.Context, activityID, userID string) error {
	return m.Called(ctx, activityID, userID).Error(0)
}

func (m *mockActivityRepo) ListUserActivities(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]activity.Activity, error) {
	args := m.Called(ctx, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]activity.Activity), args.Error(1)
}

func (m *mockActivityRepo) ListGroupFeed(ctx context.Context, groupID, userID string, limit int32, lastTime *time.Time, lastID *string) ([]activity.Activity, error) {
	args := m.Called(ctx, groupID, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]activity.Activity), args.Error(1)
}

func setupHandlerTestRouter(uc *domain.UseCase, currentUser *user.User) chi.Router {
	r := chi.NewRouter()
	h := grouphttp.NewHandler(uc)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if currentUser != nil {
				ctx := user.WithUser(r.Context(), currentUser)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
			} else {
				response.Unauthorized(w, "unauthorized")
			}
		})
	})

	h.RegisterRoutes(r)
	return r
}

// --- POST /groups Tests ---

func TestHandler_CreateGroup_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}

	currentUser := &user.User{ID: "usr-creator", Name: "Creator"}
	createdGroup := &domain.Group{ID: "grp-1", Name: "New Group", CreatedBy: &currentUser.ID}

	mockRepo.On("CreateGroup", mock.Anything, mock.AnythingOfType("*domain.Group")).Return(nil)
	mockRepo.On("AddGroupMembers", mock.Anything, mock.Anything, []string{currentUser.ID}, domain.MemberRoleAdmin, domain.MemberStatusActive).Return([]domain.Member{}, nil)
	mockRepo.On("ListGroupMembers", mock.Anything, mock.Anything, mock.Anything).Return([]domain.Member{
		{GroupID: "grp-1", UserID: currentUser.ID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive},
	}, nil)

	mockAct.On("LogEvent",
		mock.Anything, currentUser.ID, mock.Anything, mock.Anything, mock.Anything,
	).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]string{"name": "New Group", "description": "Desc"})
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var respGroup domain.Group
	err := json.Unmarshal(rr.Body.Bytes(), &respGroup)
	require.NoError(t, err)
	assert.Equal(t, createdGroup.Name, respGroup.Name)
}

func TestHandler_CreateGroup_BadRequest(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	currentUser := &user.User{ID: "usr-creator", Name: "Creator"}
	router := setupHandlerTestRouter(uc, currentUser)

	// Missing required name field
	body, _ := json.Marshal(map[string]string{"description": "No name"})
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_CreateGroup_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil) // no user context

	body, _ := json.Marshal(map[string]string{"name": "New Group"})
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- GET /groups/{id} Tests ---

func TestHandler_GetDetails_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-1", Name: "Alice"}
	groupID := "grp-1"

	member := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}
	g := &domain.Group{ID: groupID, Name: "Trip"}

	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(member, nil)
	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("ListGroupMembers", mock.Anything, groupID, mock.Anything).Return([]domain.Member{*member}, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+groupID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp domain.Group
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, groupID, resp.ID)
	assert.Equal(t, "Trip", resp.Name)
	assert.Len(t, resp.Members, 1)
}

func TestHandler_GetDetails_Forbidden(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-stranger"}
	groupID := "grp-1"

	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(nil, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+groupID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_GetDetails_NotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-1"}
	groupID := "grp-nonexistent"

	member := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(member, nil)
	mockRepo.On("GetByID", mock.Anything, groupID).Return(nil, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+groupID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_GetDetails_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil)

	req := httptest.NewRequest(http.MethodGet, "/groups/grp-1", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- POST /groups/{id}/members Tests ---

func TestHandler_AddMembers_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}

	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-1"
	targetUserIDs := []string{"usr-target-1", "usr-target-2"}

	adminMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive}
	g := &domain.Group{ID: groupID, Name: "Trip"}

	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("AddGroupMembers", mock.Anything, groupID, targetUserIDs, domain.MemberRoleMember, domain.MemberStatusActive).Return([]domain.Member{
		{GroupID: groupID, UserID: "usr-target-1", Role: domain.MemberRoleMember, Status: domain.MemberStatusActive},
		{GroupID: groupID, UserID: "usr-target-2", Role: domain.MemberRoleMember, Status: domain.MemberStatusActive},
	}, nil)

	mockAct.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockNotif.On("CreateAlert", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]interface{}{"userIds": targetUserIDs})
	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID+"/members", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var resp []domain.Member
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp, 2)
}

func TestHandler_AddMembers_BadRequest_MissingUserIDs(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-1"

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]interface{}{"userIds": []string{}})
	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID+"/members", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_AddMembers_Forbidden(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-member"}
	groupID := "grp-1"

	regularMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}
	mockRepo.On("GetByID", mock.Anything, groupID).Return(&domain.Group{ID: groupID}, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(regularMember, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]interface{}{"userIds": []string{"usr-new"}})
	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID+"/members", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_AddMembers_NotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-nonexistent"

	adminMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive}
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetByID", mock.Anything, groupID).Return(nil, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]interface{}{"userIds": []string{"usr-new"}})
	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID+"/members", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_AddMembers_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil)

	body, _ := json.Marshal(map[string]interface{}{"userIds": []string{"usr-new"}})
	req := httptest.NewRequest(http.MethodPost, "/groups/grp-1/members", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- DELETE /groups/{id}/members/{userId} Tests ---

func TestHandler_RemoveMember_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}

	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-1"
	targetUserID := "usr-target"

	g := &domain.Group{ID: groupID, Name: "Trip"}
	adminMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive}
	targetMember := domain.Member{GroupID: groupID, UserID: targetUserID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}
	members := []domain.Member{*adminMember, targetMember}

	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, targetUserID).Return(&targetMember, nil)
	mockRepo.On("ListGroupMembers", mock.Anything, groupID, mock.Anything).Return(members, nil)
	mockRepo.On("RemoveGroupMember", mock.Anything, groupID, targetUserID).Return(nil)

	mockAct.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockNotif.On("CreateAlert", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID+"/members/"+targetUserID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestHandler_RemoveMember_BadRequest_SoleAdmin(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-sole-admin"}
	groupID := "grp-1"

	g := &domain.Group{ID: groupID, Name: "Trip"}
	adminMember := domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive}
	otherMember := domain.Member{GroupID: groupID, UserID: "usr-other", Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}

	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(&adminMember, nil)
	mockRepo.On("ListGroupMembers", mock.Anything, groupID, mock.Anything).Return([]domain.Member{adminMember, otherMember}, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID+"/members/"+currentUser.ID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_RemoveMember_Forbidden(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-regular"}
	groupID := "grp-1"
	targetUserID := "usr-other"

	g := &domain.Group{ID: groupID, Name: "Trip"}
	regularMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}

	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(regularMember, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID+"/members/"+targetUserID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_RemoveMember_NotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-nonexistent"

	adminMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive}
	mockRepo.On("GetByID", mock.Anything, groupID).Return(nil, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, "usr-target").Return(nil, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID+"/members/usr-target", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_RemoveMember_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil)

	req := httptest.NewRequest(http.MethodDelete, "/groups/grp-1/members/usr-2", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- PUT /groups/{id}/members/{userId}/role Tests ---

func TestHandler_UpdateMemberRole_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}

	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-1"
	targetUserID := "usr-target"

	g := &domain.Group{ID: groupID, Name: "Trip"}
	adminMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive}
	targetMember := domain.Member{GroupID: groupID, UserID: targetUserID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}
	members := []domain.Member{*adminMember, targetMember}

	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, targetUserID).Return(&targetMember, nil)
	mockRepo.On("ListGroupMembers", mock.Anything, groupID, mock.Anything).Return(members, nil)
	mockRepo.On("UpdateGroupMemberRole", mock.Anything, groupID, targetUserID, domain.MemberRoleAdmin).Return(nil)

	mockAct.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockNotif.On("CreateAlert", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]string{"role": "ADMIN"})
	req := httptest.NewRequest(http.MethodPut, "/groups/"+groupID+"/members/"+targetUserID+"/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp domain.Member
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, targetUserID, resp.UserID)
}

func TestHandler_UpdateMemberRole_BadRequest_InvalidRole(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-1"

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]string{"role": "superadmin"})
	req := httptest.NewRequest(http.MethodPut, "/groups/"+groupID+"/members/usr-target/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_UpdateMemberRole_Forbidden(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-regular"}
	groupID := "grp-1"

	g := &domain.Group{ID: groupID, Name: "Trip"}
	regularMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}

	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(regularMember, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]string{"role": "ADMIN"})
	req := httptest.NewRequest(http.MethodPut, "/groups/"+groupID+"/members/usr-target/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_UpdateMemberRole_NotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-nonexistent"

	adminMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive}
	mockRepo.On("GetByID", mock.Anything, groupID).Return(nil, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, "usr-target").Return(nil, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]string{"role": "ADMIN"})
	req := httptest.NewRequest(http.MethodPut, "/groups/"+groupID+"/members/usr-target/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_UpdateMemberRole_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil)

	body, _ := json.Marshal(map[string]string{"role": "ADMIN"})
	req := httptest.NewRequest(http.MethodPut, "/groups/grp-1/members/usr-target/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- DELETE /groups/{id} (Archive) Tests ---

func TestHandler_Archive_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockTx := &mockTransactor{}

	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-1"

	g := &domain.Group{ID: groupID, Name: "Trip"}
	adminMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive}

	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("ListGroupMembers", mock.Anything, groupID, mock.Anything).Return([]domain.Member{*adminMember}, nil)
	mockRepo.On("Archive", mock.Anything, groupID).Return(nil)

	mockAct.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestHandler_Archive_Forbidden(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-regular"}
	groupID := "grp-1"

	regularMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}
	mockRepo.On("GetByID", mock.Anything, groupID).Return(&domain.Group{ID: groupID}, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(regularMember, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_Archive_NotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-nonexistent"

	adminMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive}
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetByID", mock.Anything, groupID).Return(nil, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_Archive_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil)

	req := httptest.NewRequest(http.MethodDelete, "/groups/grp-1", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- POST /groups/join Tests ---

func TestHandler_JoinGroup_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockTx := &mockTransactor{}

	currentUser := &user.User{ID: "usr-new"}
	inviteCode := "inv-123"
	groupID := "grp-1"
	futureExpires := time.Now().Add(1 * time.Hour)

	g := &domain.Group{ID: groupID, Name: "Trip", InviteCode: &inviteCode, InviteCodeExpiresAt: &futureExpires}
	newMember := domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}

	mockRepo.On("GetByInviteCode", mock.Anything, inviteCode).Return(g, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(nil, nil)
	mockRepo.On("AddGroupMembers", mock.Anything, groupID, []string{currentUser.ID}, domain.MemberRoleMember, mock.Anything).Return([]domain.Member{}, nil)
	mockRepo.On("ListGroupMembers", mock.Anything, groupID, mock.Anything).Return([]domain.Member{newMember}, nil)

	mockAct.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]string{"inviteCode": inviteCode})
	req := httptest.NewRequest(http.MethodPost, "/groups/join", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp domain.JoinResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, domain.MemberStatusActive, resp.Status)
	assert.Equal(t, g.ID, resp.Group.ID)
}

func TestHandler_JoinGroup_BadRequest(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-new"}

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]string{"inviteCode": ""})
	req := httptest.NewRequest(http.MethodPost, "/groups/join", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_JoinGroup_NotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-new"}

	mockRepo.On("GetByInviteCode", mock.Anything, "invalid-code").Return(nil, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]string{"inviteCode": "invalid-code"})
	req := httptest.NewRequest(http.MethodPost, "/groups/join", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_JoinGroup_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil)

	body, _ := json.Marshal(map[string]string{"inviteCode": "inv-123"})
	req := httptest.NewRequest(http.MethodPost, "/groups/join", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- GET /groups/preview Tests ---

func TestHandler_Preview_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-1"}
	inviteCode := "inv-123"

	preview := &domain.Preview{
		Name:        "Trip",
		MemberCount: 3,
		CreatorName: "Bob",
	}

	mockRepo.On("GetPreviewByInviteCode", mock.Anything, inviteCode).Return(preview, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups/preview?inviteCode="+inviteCode, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandler_Preview_BadRequest(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-1"}

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups/preview", nil) // missing query param
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_Preview_NotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-1"}
	inviteCode := "bad-code"

	mockRepo.On("GetPreviewByInviteCode", mock.Anything, inviteCode).Return(nil, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups/preview?inviteCode="+inviteCode, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- GET /groups Tests ---

func TestHandler_List_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-1"}

	groupsList := []domain.GroupWithMembers{
		{Group: domain.Group{ID: "grp-1", Name: "Group 1"}},
	}
	mockRepo.On("ListUserGroupsWithMembers", mock.Anything, currentUser.ID, int32(21), mock.Anything, mock.Anything).Return(groupsList, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp pagination.Response[domain.Group]
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "grp-1", resp.Data[0].ID)
}

func TestHandler_List_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil)

	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- GET /groups/{id}/feed Tests ---

func TestHandler_GetFeed_Success(t *testing.T) {
	mockGroupRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)

	currentUser := &user.User{ID: "usr-1"}
	groupID := "grp-1"

	member := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}
	mockGroupRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(member, nil)
	mockAct.On("GetGroupFeed", mock.Anything, currentUser.ID, groupID, mock.Anything).Return(pagination.Response[activity.Activity]{}, nil)

	uc := domain.NewUseCase(mockGroupRepo, &mockTransactor{}, mockAct, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+groupID+"/feed", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// --- POST /groups/{id}/members/{userId}/decision & ResetInviteCode Tests ---

func TestHandler_DecideJoinRequest_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}

	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-1"
	targetUserID := "usr-pending"

	adminMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive}
	approvedMember := &domain.Member{GroupID: groupID, UserID: targetUserID, Role: domain.MemberRoleMember, Status: domain.MemberStatusActive}

	mockRepo.On("GetByID", mock.Anything, groupID).Return(&domain.Group{ID: groupID}, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("UpdateMemberStatus", mock.Anything, groupID, targetUserID, domain.MemberStatusActive).Return(nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, targetUserID).Return(approvedMember, nil)
	mockAct.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockNotif.On("CreateAlert", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := domain.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	router := setupHandlerTestRouter(uc, currentUser)

	body, _ := json.Marshal(map[string]string{"action": "APPROVE"})
	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID+"/members/"+targetUserID+"/decision", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp domain.Member
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, targetUserID, resp.UserID)
}

func TestHandler_ResetInviteCode_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-1"

	adminMember := &domain.Member{GroupID: groupID, UserID: currentUser.ID, Role: domain.MemberRoleAdmin, Status: domain.MemberStatusActive}
	updatedGroup := &domain.Group{ID: groupID, Name: "Trip"}

	mockRepo.On("GetByID", mock.Anything, groupID).Return(updatedGroup, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("ResetInviteCode", mock.Anything, groupID, mock.Anything, mock.Anything).Return(updatedGroup, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID+"/invite-code/reset", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandler_SyncGroups_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-1"}

	now := time.Now()
	mockRepo.On("SyncGroupsBySequence", mock.Anything, int64(10), currentUser.ID, int32(100)).Return([]domain.Group{
		{ID: "grp-1", Name: "Trip", SyncVersion: 11},
		{ID: "grp-2", Name: "Flat", SyncVersion: 12, ArchivedAt: &now},
	}, nil)

	uc := domain.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups/sync?lastVersion=10", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp domain.GroupSyncResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, int64(12), resp.NewVersion)
	assert.Len(t, resp.Updated, 1)
	assert.Equal(t, "grp-1", resp.Updated[0].ID)
	assert.Equal(t, []string{"grp-2"}, resp.RemovedGroupIDs)
}

