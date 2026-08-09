package domain

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/notification"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/google/uuid"
)

const inviteTTL = 7 * 24 * time.Hour
const inviteCodeCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// newInviteCode generates a fresh, cryptographically secure invite code for a group.
func newInviteCode() (string, error) {
	b := make([]byte, 8)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(inviteCodeCharset))))
		if err != nil {
			return "", err
		}
		b[i] = inviteCodeCharset[num.Int64()]
	}
	return "INV-" + string(b), nil
}

// ActivityLogger records group activity events and fetches feed data.
type ActivityLogger interface {
	LogEvent(
		ctx context.Context,
		actorID string,
		groupID *string,
		visibleToUserIDs []string,
		event activity.Event,
	) error
	GetGroupFeed(
		ctx context.Context,
		userID, groupID string,
		p pagination.Params,
	) (pagination.Response[activity.Activity], error)
}

// NotificationSender delivers notifications to users.
type NotificationSender interface {
	CreateAlert(ctx context.Context, userID string, actorID *string, activityID *string, alert notification.Alert) error
}


// JoinResponse is returned when joining a group, either as an active member or pending approval.
type JoinResponse struct {
	Status  MemberStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Group   *Group       `json:"group,omitempty"`
} // @name JoinResponse

// UseCase manages business workflows for the group domain.
type UseCase struct {
	repo         Repository
	tx           db.Transactor
	activity     ActivityLogger
	notification NotificationSender
}

// NewUseCase instantiates a new UseCase.
func NewUseCase(repo Repository, tx db.Transactor, activitySvc ActivityLogger, notificationSvc NotificationSender) *UseCase {
	return &UseCase{
		repo:         repo,
		tx:           tx,
		activity:     activitySvc,
		notification: notificationSvc,
	}
}

// CreateGroup creates a new group, sets initial 7-day invite code expiration, and assigns creator as active admin.
func (u *UseCase) CreateGroup(ctx context.Context, name, description string, requireAdminApproval bool, creatorID string) (*Group, error) {
	if name == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgMissingGroupName,
		}
	}
	if creatorID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	expiresAt := time.Now().Add(inviteTTL)
	inviteCode, err := newInviteCode()
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to generate invite code",
			Err:     err,
		}
	}
	newGroup := &Group{
		ID:                   uuid.New().String(),
		Name:                 name,
		InviteCode:           &inviteCode,
		InviteCodeExpiresAt:  &expiresAt,
		RequireAdminApproval: requireAdminApproval,
		CreatedBy:            &creatorID,
	}
	if description != "" {
		newGroup.Description = &description
	}

	var members []Member
	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.CreateGroup(txCtx, newGroup); err != nil {
			return err
		}
		if err := u.repo.AddGroupMember(txCtx, newGroup.ID, creatorID, MemberRoleAdmin, MemberStatusActive); err != nil {
			return err
		}
		var err error
		members, err = u.repo.ListGroupMembers(txCtx, newGroup.ID, MemberStatusActive)
		return err
	})
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogCreateGroup,
			Err:     err,
		}
	}

	// Dispatch side effect post-transaction asynchronously
	go func() {
		payload := activity.GroupPayload{
			Group:   *newGroup,
			Members: members,
		}
		_ = u.activity.LogEvent(
			context.Background(), creatorID, &newGroup.ID, nil,
			activity.NewGroupCreatedEvent(newGroup.ID, payload),
		)
	}()

	return newGroup, nil
}

// GetGroupDetails retrieves a group and its members, verifying the requester is an ACTIVE member.
func (u *UseCase) GetGroupDetails(ctx context.Context, groupID, userID string) (*Group, error) {
	if groupID == "" || userID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	member, err := u.repo.GetGroupMember(ctx, groupID, userID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogVerifyMembership,
			Err:     err,
		}
	}
	if member == nil || member.Status != MemberStatusActive {
		return nil, &response.AppError{
			Type:    response.TypeForbidden,
			Message: response.MsgNotGroupMember,
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveGroup,
			Err:     err,
		}
	}
	if g == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgGroupNotFound,
		}
	}

	members, err := u.repo.ListGroupMembers(ctx, groupID, MemberStatusActive)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveMembers,
			Err:     err,
		}
	}

	g.Members = members
	return g, nil
}

// GetGroupFeed retrieves group activity feed verified for the requesting user.
func (u *UseCase) GetGroupFeed(ctx context.Context, groupID, userID string, p pagination.Params) (pagination.Response[activity.Activity], error) {
	if groupID == "" || userID == "" {
		return pagination.Response[activity.Activity]{}, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}
	member, err := u.repo.GetGroupMember(ctx, groupID, userID)
	if err != nil {
		return pagination.Response[activity.Activity]{}, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogVerifyMembership,
			Err:     err,
		}
	}
	if member == nil || member.Status != MemberStatusActive {
		return pagination.Response[activity.Activity]{}, &response.AppError{
			Type:    response.TypeForbidden,
			Message: response.MsgNotGroupMember,
		}
	}
	return u.activity.GetGroupFeed(ctx, userID, groupID, p)
}

// ListUserGroups returns a cursor-paginated list of groups the user actively belongs to.
func (u *UseCase) ListUserGroups(ctx context.Context, userID string, p pagination.Params) (pagination.Response[Group], error) {
	if userID == "" {
		return pagination.Response[Group]{}, &response.AppError{Type: response.TypeValidation, Message: response.MsgInvalidParam}
	}
	cursor := pagination.ParseCursor(p.Cursor)
	rows, err := u.repo.ListUserGroupsWithMembers(ctx, userID, p.Limit+1, cursor.LastTime, cursor.LastID)
	if err != nil {
		return pagination.Response[Group]{}, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveUserGroups,
			Err:     err,
		}
	}

	groups := make([]Group, 0, len(rows))
	for _, row := range rows {
		g := row.Group
		g.Members = row.Members
		groups = append(groups, g)
	}

	return pagination.BuildResponse(groups, p.Limit, func(g Group) string {
		return pagination.EncodeCursor(g.CreatedAt, g.ID)
	}), nil
}

// ListMembers retrieves group members with status filter. Non-active queries require admin role.
func (u *UseCase) ListMembers(ctx context.Context, groupID, statusFilter, actionByUserID string) ([]Member, error) {
	if groupID == "" || actionByUserID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	statusFilter = strings.ToUpper(strings.TrimSpace(statusFilter))
	switch statusFilter {
	case "", "ALL", string(MemberStatusActive), string(MemberStatusPending), string(MemberStatusRejected):
	default:
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidStatusFilter,
		}
	}

	if statusFilter != "" && statusFilter != string(MemberStatusActive) {
		isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
		if err != nil {
			return nil, err
		}
		if !isAdmin {
			return nil, &response.AppError{
				Type:    response.TypeForbidden,
				Message: response.MsgAdminRequiredNonActive,
			}
		}
	} else {
		member, err := u.repo.GetGroupMember(ctx, groupID, actionByUserID)
		if err != nil {
			return nil, &response.AppError{Type: response.TypeInternal, Message: response.ErrLogVerifyMemberRole, Err: err}
		}
		if member == nil || member.Status != MemberStatusActive {
			return nil, &response.AppError{Type: response.TypeForbidden, Message: response.MsgNotGroupMember}
		}
	}

	dbFilter := statusFilter
	if dbFilter == "ALL" {
		dbFilter = ""
	}

	members, err := u.repo.ListGroupMembers(ctx, groupID, MemberStatus(dbFilter))
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveMembers,
			Err:     err,
		}
	}
	return members, nil
}

// AddMembers adds new users to the group in bulk. Requires requester to be an admin.
func (u *UseCase) AddMembers(ctx context.Context, groupID string, targetUserIDs []string, actionByUserID string) ([]Member, error) {
	if groupID == "" || len(targetUserIDs) == 0 || actionByUserID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveGroup,
			Err:     err,
		}
	}
	if g == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgGroupNotFound,
		}
	}

	isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, &response.AppError{
			Type:    response.TypeForbidden,
			Message: response.MsgOnlyAdminAddMembers,
		}
	}

	var addedMembers []Member
	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		var err error
		addedMembers, err = u.repo.AddGroupMembers(txCtx, groupID, targetUserIDs, MemberRoleMember, MemberStatusActive)
		if err != nil {
			return err
		}

		for _, m := range addedMembers {
			payload := activity.MemberPayload{
				Member: m,
			}
			_ = u.activity.LogEvent(
				txCtx, actionByUserID, &groupID, nil,
				activity.NewMemberAddedEvent(m.UserID, payload),
			)
		}
		return nil
	})
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogAddMember,
			Err:     err,
		}
	}

	return addedMembers, nil
}

// RemoveMember removes a member from the group.
func (u *UseCase) RemoveMember(ctx context.Context, groupID, targetUserID, actionByUserID string) error {
	if groupID == "" || targetUserID == "" || actionByUserID == "" {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveGroup,
			Err:     err,
		}
	}
	if g == nil {
		return &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgGroupNotFound,
		}
	}

	if targetUserID != actionByUserID {
		isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
		if err != nil {
			return err
		}
		if !isAdmin {
			return &response.AppError{
				Type:    response.TypeForbidden,
				Message: response.MsgOnlyAdminRemoveMembers,
			}
		}
	}

	targetMember, err := u.repo.GetGroupMember(ctx, groupID, targetUserID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogVerifyMemberDetails,
			Err:     err,
		}
	}
	if targetMember == nil {
		return &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgMemberNotFound,
		}
	}

	if targetMember.Role == MemberRoleAdmin {
		members, err := u.repo.ListGroupMembers(ctx, groupID, MemberStatusActive)
		if err != nil {
			return &response.AppError{
				Type:    response.TypeInternal,
				Message: response.ErrLogListGroupMembers,
				Err:     err,
			}
		}
		adminCount := 0
		for _, m := range members {
			if m.Role == MemberRoleAdmin {
				adminCount++
			}
		}
		if adminCount <= 1 {
			return &response.AppError{
				Type:    response.TypeValidation,
				Message: response.MsgSoleAdminRemovalError,
			}
		}
	}

	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.RemoveGroupMember(txCtx, groupID, targetUserID); err != nil {
			return err
		}

		payload := activity.MemberPayload{
			Member: targetMember,
		}
		var evt activity.Event
		if targetUserID == actionByUserID {
			evt = activity.NewMemberLeftEvent(targetUserID, payload)
		} else {
			evt = activity.NewMemberKickedEvent(targetUserID, payload)
		}

		err = u.activity.LogEvent(
			txCtx, actionByUserID, &groupID, nil, evt,
		)
		return err
	})
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRemoveMember,
			Err:     err,
		}
	}

	return nil
}

// UpdateMemberRole updates a member's role (admin or member).
func (u *UseCase) UpdateMemberRole(ctx context.Context, groupID, targetUserID string, newRole MemberRole, actionByUserID string) (*Member, error) {
	if groupID == "" || targetUserID == "" || newRole == "" || actionByUserID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	if newRole != MemberRoleAdmin && newRole != MemberRoleMember {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidMemberRole,
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveGroup,
			Err:     err,
		}
	}
	if g == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgGroupNotFound,
		}
	}

	isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, &response.AppError{
			Type:    response.TypeForbidden,
			Message: response.MsgOnlyAdminRoleUpdate,
		}
	}
	targetMember, err := u.repo.GetGroupMember(ctx, groupID, targetUserID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogVerifyMemberDetails,
			Err:     err,
		}
	}
	if targetMember == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgMemberNotFound,
		}
	}

	var updatedMember Member
	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.UpdateGroupMemberRole(txCtx, groupID, targetUserID, newRole); err != nil {
			return err
		}

		member, err := u.repo.GetGroupMember(txCtx, groupID, targetUserID)
		if err != nil {
			return err
		}
		if member != nil {
			updatedMember = *member
		}

		payload := activity.MemberPayload{
			Member: member,
		}
		return u.activity.LogEvent(
			txCtx, actionByUserID, &groupID, nil,
			activity.NewMemberRoleUpdatedEvent(targetUserID, string(newRole), payload),
		)
	})
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogUpdateMemberRole,
			Err:     err,
		}
	}

	return &updatedMember, nil
}

// UpdateGroup updates group name, description, and admin approval requirement.
func (u *UseCase) UpdateGroup(ctx context.Context, groupID, name, description string, requireAdminApproval bool, actionByUserID string) (*Group, error) {
	if groupID == "" || name == "" || actionByUserID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveGroup,
			Err:     err,
		}
	}
	if g == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgGroupNotFound,
		}
	}

	isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, &response.AppError{
			Type:    response.TypeForbidden,
			Message: response.MsgOnlyAdminGroupUpdate,
		}
	}

	g.Name = name
	if description != "" {
		g.Description = &description
	} else {
		g.Description = nil
	}
	g.RequireAdminApproval = requireAdminApproval

	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.Update(txCtx, g); err != nil {
			return err
		}

		payload := activity.GroupPayload{
			Group: g,
		}
		err = u.activity.LogEvent(
			txCtx, actionByUserID, &groupID, nil,
			activity.NewGroupUpdatedEvent(groupID, payload),
		)
		return err
	})
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogUpdateGroup,
			Err:     err,
		}
	}

	return g, nil
}

// ArchiveGroup soft-deletes a group.
func (u *UseCase) ArchiveGroup(ctx context.Context, groupID, actionByUserID string) error {
	if groupID == "" || actionByUserID == "" {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveGroup,
			Err:     err,
		}
	}
	if g == nil {
		return &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgGroupNotFound,
		}
	}

	isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return &response.AppError{
			Type:    response.TypeForbidden,
			Message: response.MsgOnlyAdminArchiveGroup,
		}
	}

	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.Archive(txCtx, groupID); err != nil {
			return err
		}

		payload := activity.GroupPayload{
			Group: g,
		}
		err = u.activity.LogEvent(
			txCtx, actionByUserID, &groupID, nil,
			activity.NewGroupArchivedEvent(groupID, payload),
		)
		return err
	})
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogArchiveGroup,
			Err:     err,
		}
	}
	return nil
}

// JoinGroup handles joining a group via invite code with expiration & admin approval check.
func (u *UseCase) JoinGroup(ctx context.Context, inviteCode, userID string) (*JoinResponse, error) {
	if inviteCode == "" || userID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	g, err := u.repo.GetByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogLookupInviteCode,
			Err:     err,
		}
	}
	if g == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgInvalidInviteCode,
		}
	}

	if g.InviteCodeExpiresAt != nil && g.InviteCodeExpiresAt.Before(time.Now()) {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidInviteCode,
		}
	}

	existing, err := u.repo.GetGroupMember(ctx, g.ID, userID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogVerifyMembership,
			Err:     err,
		}
	}
	if existing != nil {
		if existing.Status == MemberStatusActive {
			return &JoinResponse{Status: MemberStatusActive, Group: g}, nil
		}
		if existing.Status == MemberStatusPending {
			return &JoinResponse{Status: MemberStatusPending, Message: response.MsgJoinRequestSubmitted}, nil
		}
		if existing.Status == MemberStatusRejected {
			return nil, &response.AppError{
				Type:    response.TypeForbidden,
				Message: response.MsgJoinRequestRejected,
			}
		}
	}

	targetStatus := MemberStatusActive
	if g.RequireAdminApproval {
		targetStatus = MemberStatusPending
	}

	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.AddGroupMember(txCtx, g.ID, userID, MemberRoleMember, targetStatus); err != nil {
			return err
		}

		if targetStatus == MemberStatusPending {
			members, err := u.repo.ListGroupMembers(txCtx, g.ID, MemberStatusActive)
			if err == nil {
				for _, m := range members {
					if m.Role == MemberRoleAdmin && u.notification != nil {
						_ = u.notification.CreateAlert(
							txCtx, m.UserID, &userID, nil,
							notification.NewJoinRequestPendingAlert(g.Name),
						)
					}
				}
			}
			return nil
		}

		members, err := u.repo.ListGroupMembers(txCtx, g.ID, MemberStatusActive)
		if err != nil {
			return err
		}
		var targetMember Member
		for _, m := range members {
			if m.UserID == userID {
				targetMember = m
				break
			}
		}

		payload := activity.MemberPayload{
			Member: targetMember,
		}
		err = u.activity.LogEvent(
			txCtx, userID, &g.ID, nil,
			activity.NewMemberJoinedEvent(userID, payload),
		)
		return err
	})
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogJoinGroup,
			Err:     err,
		}
	}

	if targetStatus == MemberStatusPending {
		return &JoinResponse{Status: MemberStatusPending, Message: response.MsgJoinRequestSubmitted}, nil
	}
	return &JoinResponse{Status: MemberStatusActive, Group: g}, nil
}

// DecideJoinRequest approves or rejects a pending join request. Admin only.
func (u *UseCase) DecideJoinRequest(ctx context.Context, groupID, targetUserID string, action JoinRequestAction, adminUserID string) (*Member, error) {
	if groupID == "" || targetUserID == "" || action == "" || adminUserID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveGroup,
			Err:     err,
		}
	}
	if g == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgGroupNotFound,
		}
	}

	isAdmin, err := u.checkIsAdmin(ctx, groupID, adminUserID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, &response.AppError{
			Type:    response.TypeForbidden,
			Message: response.MsgOnlyAdminDecideJoin,
		}
	}

	var newStatus MemberStatus
	switch action {
	case JoinRequestActionApprove:
		newStatus = MemberStatusActive
	case JoinRequestActionReject:
		newStatus = MemberStatusRejected
	default:
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidAction,
		}
	}

	var updatedMember Member
	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.UpdateMemberStatus(txCtx, groupID, targetUserID, newStatus); err != nil {
			return err
		}

		member, err := u.repo.GetGroupMember(txCtx, groupID, targetUserID)
		if err != nil {
			return err
		}
		if member != nil {
			updatedMember = *member
		}

		if newStatus == MemberStatusActive && member != nil {
			payload := activity.MemberPayload{
				Member: *member,
			}
			_ = u.activity.LogEvent(
				txCtx, adminUserID, &groupID, nil,
				activity.NewMemberJoinedEvent(targetUserID, payload),
			)
		}
		if u.notification != nil {
			var alert notification.Alert
			if newStatus == MemberStatusRejected {
				alert = notification.NewJoinRequestRejectedAlert()
			} else {
				alert = notification.NewJoinRequestApprovedAlert()
			}
			_ = u.notification.CreateAlert(txCtx, targetUserID, &adminUserID, nil, alert)
		}
		return nil
	})
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogDecideJoinRequest,
			Err:     err,
		}
	}

	return &updatedMember, nil
}

// ResetInviteCode generates a new invite code with a 7-day expiration. Admin only.
func (u *UseCase) ResetInviteCode(ctx context.Context, groupID, adminUserID string) (*Group, error) {
	if groupID == "" || adminUserID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveGroup,
			Err:     err,
		}
	}
	if g == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgGroupNotFound,
		}
	}

	isAdmin, err := u.checkIsAdmin(ctx, groupID, adminUserID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, &response.AppError{
			Type:    response.TypeForbidden,
			Message: response.MsgOnlyAdminResetCode,
		}
	}

	code, err := newInviteCode()
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to generate invite code",
			Err:     err,
		}
	}
	expiresAt := time.Now().Add(inviteTTL)

	var updatedGroup *Group
	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		var err error
		updatedGroup, err = u.repo.ResetInviteCode(txCtx, groupID, code, expiresAt)
		return err
	})
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogResetInviteCode,
			Err:     err,
		}
	}

	return updatedGroup, nil
}

// GetGroupPreview looks up group details for user preview before joining.
func (u *UseCase) GetGroupPreview(ctx context.Context, inviteCode string) (*Preview, error) {
	if inviteCode == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	preview, err := u.repo.GetPreviewByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrievePreview,
			Err:     err,
		}
	}
	if preview == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgInvalidInviteCode,
		}
	}

	return preview, nil
}

func (u *UseCase) checkIsAdmin(ctx context.Context, groupID, userID string) (bool, error) {
	member, err := u.repo.GetGroupMember(ctx, groupID, userID)
	if err != nil {
		return false, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogVerifyMemberRole,
			Err:     err,
		}
	}
	if member == nil || member.Status != MemberStatusActive {
		return false, nil
	}
	return member.Role == MemberRoleAdmin, nil
}
