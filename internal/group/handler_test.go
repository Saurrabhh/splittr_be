package group_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/group"
	"github.com/Saurrabhh/splittr_be/internal/notification"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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

func setupHandlerTestRouter(uc *group.UseCase, activityUC *activity.UseCase, currentUser *user.User) chi.Router {
	r := chi.NewRouter()
	h := group.NewHandler(uc, activityUC)

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
	createdGroup := &group.Group{ID: "grp-1", Name: "New Group", CreatedBy: &currentUser.ID}

	mockRepo.On("CreateGroup", mock.Anything, mock.AnythingOfType("*domain.Group")).Return(nil)
	mockRepo.On("AddGroupMember", mock.Anything, mock.Anything, currentUser.ID, "admin", string(group.MemberStatusActive)).Return(nil)
	mockRepo.On("ListGroupMembers", mock.Anything, mock.Anything, mock.Anything).Return([]group.Member{
		{GroupID: "grp-1", UserID: currentUser.ID, Role: "admin", Status: string(group.MemberStatusActive)},
	}, nil)

	mockAct.On("LogEvent",
		mock.Anything, currentUser.ID, mock.Anything, mock.Anything, mock.Anything,
	).Return(&activity.Activity{ID: "act-1"}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	body, _ := json.Marshal(map[string]string{"name": "New Group", "description": "Desc"})
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var respGroup group.Group
	err := json.Unmarshal(rr.Body.Bytes(), &respGroup)
	require.NoError(t, err)
	assert.Equal(t, createdGroup.Name, respGroup.Name)
}

func TestHandler_CreateGroup_BadRequest(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	currentUser := &user.User{ID: "usr-creator", Name: "Creator"}
	router := setupHandlerTestRouter(uc, nil, currentUser)

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
	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, nil) // no user context

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

	member := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "member", Status: string(group.MemberStatusActive)}
	g := &group.Group{ID: groupID, Name: "Trip"}

	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(member, nil)
	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("ListGroupMembers", mock.Anything, groupID, mock.Anything).Return([]group.Member{*member}, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+groupID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp group.DetailsResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, groupID, resp.Group.ID)
	assert.Equal(t, "Trip", resp.Group.Name)
	assert.Len(t, resp.Members, 1)
}

func TestHandler_GetDetails_Forbidden(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-stranger"}
	groupID := "grp-1"

	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(nil, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+groupID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_GetDetails_NotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-1"}
	groupID := "grp-nonexistent"

	member := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "member", Status: string(group.MemberStatusActive)}
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(member, nil)
	mockRepo.On("GetByID", mock.Anything, groupID).Return(nil, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+groupID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_GetDetails_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/groups/grp-1", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- POST /groups/{id}/members Tests ---

func TestHandler_AddMember_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	mockAct := new(mockActivityLogger)
	mockNotif := new(mockNotificationSender)
	mockTx := &mockTransactor{}

	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-1"
	targetUserID := "usr-target"

	adminMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "admin", Status: string(group.MemberStatusActive)}
	g := &group.Group{ID: groupID, Name: "Trip"}

	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("AddGroupMember", mock.Anything, groupID, targetUserID, "member", mock.Anything).Return(nil)
	mockRepo.On("ListGroupMembers", mock.Anything, groupID, mock.Anything).Return([]group.Member{
		{GroupID: groupID, UserID: targetUserID, Role: "member", Status: string(group.MemberStatusActive)},
	}, nil)

	act := &activity.Activity{ID: "act-1"}
	mockAct.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(act, nil)
	mockNotif.On("CreateAlert", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&notification.Notification{}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	body, _ := json.Marshal(map[string]string{"userId": targetUserID})
	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID+"/members", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp response.MessageResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "member added successfully", resp.Message)
}

func TestHandler_AddMember_BadRequest_MissingUserID(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-1"

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	body, _ := json.Marshal(map[string]string{"userId": ""})
	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID+"/members", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_AddMember_Forbidden(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-member"}
	groupID := "grp-1"

	regularMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "member", Status: string(group.MemberStatusActive)}
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(regularMember, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	body, _ := json.Marshal(map[string]string{"userId": "usr-new"})
	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID+"/members", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_AddMember_NotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-nonexistent"

	adminMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "admin", Status: string(group.MemberStatusActive)}
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetByID", mock.Anything, groupID).Return(nil, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	body, _ := json.Marshal(map[string]string{"userId": "usr-new"})
	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID+"/members", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_AddMember_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, nil)

	body, _ := json.Marshal(map[string]string{"userId": "usr-new"})
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

	g := &group.Group{ID: groupID, Name: "Trip"}
	adminMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "admin", Status: string(group.MemberStatusActive)}
	targetMember := group.Member{GroupID: groupID, UserID: targetUserID, Role: "member", Status: string(group.MemberStatusActive)}
	members := []group.Member{*adminMember, targetMember}

	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, targetUserID).Return(&targetMember, nil)
	mockRepo.On("ListGroupMembers", mock.Anything, groupID, mock.Anything).Return(members, nil)
	mockRepo.On("RemoveGroupMember", mock.Anything, groupID, targetUserID).Return(nil)

	act := &activity.Activity{ID: "act-1"}
	mockAct.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(act, nil)
	mockNotif.On("CreateAlert", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&notification.Notification{}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID+"/members/"+targetUserID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestHandler_RemoveMember_BadRequest_SoleAdmin(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-sole-admin"}
	groupID := "grp-1"

	g := &group.Group{ID: groupID, Name: "Trip"}
	adminMember := group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "admin", Status: string(group.MemberStatusActive)}
	otherMember := group.Member{GroupID: groupID, UserID: "usr-other", Role: "member", Status: string(group.MemberStatusActive)}

	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(&adminMember, nil)
	mockRepo.On("ListGroupMembers", mock.Anything, groupID, mock.Anything).Return([]group.Member{adminMember, otherMember}, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

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

	g := &group.Group{ID: groupID, Name: "Trip"}
	regularMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "member", Status: string(group.MemberStatusActive)}

	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(regularMember, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID+"/members/"+targetUserID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_RemoveMember_NotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-nonexistent"

	adminMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "admin", Status: string(group.MemberStatusActive)}
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, "usr-target").Return(nil, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID+"/members/usr-target", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_RemoveMember_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, nil)

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

	g := &group.Group{ID: groupID, Name: "Trip"}
	adminMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "admin", Status: string(group.MemberStatusActive)}
	targetMember := group.Member{GroupID: groupID, UserID: targetUserID, Role: "member", Status: string(group.MemberStatusActive)}
	members := []group.Member{*adminMember, targetMember}

	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, targetUserID).Return(&targetMember, nil)
	mockRepo.On("ListGroupMembers", mock.Anything, groupID, mock.Anything).Return(members, nil)
	mockRepo.On("UpdateGroupMemberRole", mock.Anything, groupID, targetUserID, "admin").Return(nil)

	act := &activity.Activity{ID: "act-1"}
	mockAct.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(act, nil)
	mockNotif.On("CreateAlert", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&notification.Notification{}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	body, _ := json.Marshal(map[string]string{"role": "admin"})
	req := httptest.NewRequest(http.MethodPut, "/groups/"+groupID+"/members/"+targetUserID+"/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp response.MessageResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "role updated successfully", resp.Message)
}

func TestHandler_UpdateMemberRole_BadRequest_InvalidRole(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-1"

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

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

	g := &group.Group{ID: groupID, Name: "Trip"}
	regularMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "member", Status: string(group.MemberStatusActive)}

	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(regularMember, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	body, _ := json.Marshal(map[string]string{"role": "admin"})
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

	adminMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "admin", Status: string(group.MemberStatusActive)}
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, "usr-target").Return(nil, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	body, _ := json.Marshal(map[string]string{"role": "admin"})
	req := httptest.NewRequest(http.MethodPut, "/groups/"+groupID+"/members/usr-target/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_UpdateMemberRole_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, nil)

	body, _ := json.Marshal(map[string]string{"role": "admin"})
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

	g := &group.Group{ID: groupID, Name: "Trip"}
	adminMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "admin", Status: string(group.MemberStatusActive)}

	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetByID", mock.Anything, groupID).Return(g, nil)
	mockRepo.On("ListGroupMembers", mock.Anything, groupID, mock.Anything).Return([]group.Member{*adminMember}, nil)
	mockRepo.On("Archive", mock.Anything, groupID).Return(nil)

	mockAct.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&activity.Activity{ID: "act-1"}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp response.MessageResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "group archived successfully", resp.Message)
}

func TestHandler_Archive_Forbidden(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-regular"}
	groupID := "grp-1"

	regularMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "member", Status: string(group.MemberStatusActive)}
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(regularMember, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_Archive_NotFound(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-nonexistent"

	adminMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "admin", Status: string(group.MemberStatusActive)}
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("GetByID", mock.Anything, groupID).Return(nil, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_Archive_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, nil)

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

	g := &group.Group{ID: groupID, Name: "Trip", InviteCode: &inviteCode, InviteCodeExpiresAt: &futureExpires}
	newMember := group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "member", Status: string(group.MemberStatusActive)}

	mockRepo.On("GetByInviteCode", mock.Anything, inviteCode).Return(g, nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(nil, nil)
	mockRepo.On("AddGroupMember", mock.Anything, groupID, currentUser.ID, "member", mock.Anything).Return(nil)
	mockRepo.On("ListGroupMembers", mock.Anything, groupID, mock.Anything).Return([]group.Member{newMember}, nil)

	mockAct.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&activity.Activity{ID: "act-1"}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	body, _ := json.Marshal(map[string]string{"inviteCode": inviteCode})
	req := httptest.NewRequest(http.MethodPost, "/groups/join", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp group.JoinResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, string(group.MemberStatusActive), resp.Status)
	assert.Equal(t, g.ID, resp.Group.ID)
}

func TestHandler_JoinGroup_BadRequest(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-new"}

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

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

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	body, _ := json.Marshal(map[string]string{"inviteCode": "invalid-code"})
	req := httptest.NewRequest(http.MethodPost, "/groups/join", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_JoinGroup_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, nil)

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

	preview := &group.Preview{
		Name:        "Trip",
		MemberCount: 3,
		CreatorName: "Bob",
	}

	mockRepo.On("GetPreviewByInviteCode", mock.Anything, inviteCode).Return(preview, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups/preview?inviteCode="+inviteCode, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandler_Preview_BadRequest(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-1"}

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

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

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups/preview?inviteCode="+inviteCode, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- GET /groups Tests ---

func TestHandler_List_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-1"}

	groupsList := []group.GroupWithMembers{
		{Group: group.Group{ID: "grp-1", Name: "Group 1"}},
	}
	mockRepo.On("ListUserGroupsWithMembers", mock.Anything, currentUser.ID, int32(21), mock.Anything, mock.Anything).Return(groupsList, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp group.ListGroupsResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "grp-1", resp.Data[0].Group.ID)
}

func TestHandler_List_Unauthorized(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- GET /groups/{id}/feed Tests ---

func TestHandler_GetFeed_Success(t *testing.T) {
	mockGroupRepo := new(mockGroupRepository)
	mockActRepo := new(mockActivityRepo)

	currentUser := &user.User{ID: "usr-1"}
	groupID := "grp-1"

	mockActRepo.On("ListGroupFeed", mock.Anything, groupID, currentUser.ID, int32(21), mock.Anything, mock.Anything).Return([]activity.Activity{}, nil)

	uc := group.NewUseCase(mockGroupRepo, &mockTransactor{}, nil, nil)
	activityUC := activity.NewUseCase(mockActRepo)
	router := setupHandlerTestRouter(uc, activityUC, currentUser)

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

	adminMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "admin", Status: string(group.MemberStatusActive)}
	approvedMember := &group.Member{GroupID: groupID, UserID: targetUserID, Role: "member", Status: string(group.MemberStatusActive)}

	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("UpdateMemberStatus", mock.Anything, groupID, targetUserID, string(group.MemberStatusActive)).Return(nil)
	mockRepo.On("GetGroupMember", mock.Anything, groupID, targetUserID).Return(approvedMember, nil)
	mockAct.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&activity.Activity{ID: "act-1"}, nil)
	mockNotif.On("CreateAlert", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&notification.Notification{}, nil)

	uc := group.NewUseCase(mockRepo, mockTx, mockAct, mockNotif)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	body, _ := json.Marshal(map[string]string{"action": "APPROVE"})
	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID+"/members/"+targetUserID+"/decision", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandler_ResetInviteCode_Success(t *testing.T) {
	mockRepo := new(mockGroupRepository)
	currentUser := &user.User{ID: "usr-admin"}
	groupID := "grp-1"

	adminMember := &group.Member{GroupID: groupID, UserID: currentUser.ID, Role: "admin", Status: string(group.MemberStatusActive)}
	updatedGroup := &group.Group{ID: groupID, Name: "Trip"}

	mockRepo.On("GetGroupMember", mock.Anything, groupID, currentUser.ID).Return(adminMember, nil)
	mockRepo.On("ResetInviteCode", mock.Anything, groupID, mock.Anything, mock.Anything).Return(updatedGroup, nil)

	uc := group.NewUseCase(mockRepo, &mockTransactor{}, nil, nil)
	router := setupHandlerTestRouter(uc, nil, currentUser)

	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID+"/invite-code/reset", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
