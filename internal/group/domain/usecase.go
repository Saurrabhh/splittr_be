package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/google/uuid"
)

const inviteTTL = 7 * 24 * time.Hour

// newInviteCode generates a fresh invite code for a group.
func newInviteCode() string {
	return "invite-" + uuid.New().String()[:8]
}

// ActivityLogger records group activity events.
type ActivityLogger interface {
	LogEvent(
		ctx context.Context,
		actorID string,
		groupID *string,
		visibleToUserIDs []string,
		event activity.Event,
	) error
}

// NotificationSender delivers notifications to users.
type NotificationSender interface {
	CreateAlert(ctx context.Context, userID string, actorID *string, activityID *string, title, content string) error
}

// DetailsResponse is the canonical shape for any endpoint or feed payload that returns group data.
type DetailsResponse struct {
	Group   Group    `json:"group"`
	Members []Member `json:"members"`
}

// JoinResponse is returned when joining a group, either as an active member or pending approval.
type JoinResponse struct {
	Status  MemberStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Group   *Group       `json:"group,omitempty"`
}

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
			Message: "group name is required",
		}
	}
	if creatorID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "creator id is required",
		}
	}

	expiresAt := time.Now().Add(inviteTTL)
	inviteCode := newInviteCode()
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

	err := u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.CreateGroup(txCtx, newGroup); err != nil {
			return err
		}
		if err := u.repo.AddGroupMember(txCtx, newGroup.ID, creatorID, MemberRoleAdmin, MemberStatusActive); err != nil {
			return err
		}
		members, err := u.repo.ListGroupMembers(txCtx, newGroup.ID, string(MemberStatusActive))
		if err != nil {
			return err
		}

		payload := activity.GroupPayload{
			Group:   *newGroup,
			Members: members,
		}
		err = u.activity.LogEvent(
			txCtx, creatorID, &newGroup.ID, nil,
			activity.NewGroupCreatedEvent(newGroup.ID, payload),
		)
		return err
	})
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to create group",
			Err:     err,
		}
	}

	return newGroup, nil
}

// GetGroupDetails retrieves a group and its members, verifying the requester is an ACTIVE member.
func (u *UseCase) GetGroupDetails(ctx context.Context, groupID, userID string) (*Group, []Member, error) {
	if groupID == "" || userID == "" {
		return nil, nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "group id and user id are required",
		}
	}

	member, err := u.repo.GetGroupMember(ctx, groupID, userID)
	if err != nil {
		return nil, nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to verify group membership status",
			Err:     err,
		}
	}
	if member == nil || member.Status != MemberStatusActive {
		return nil, nil, &response.AppError{
			Type:    response.TypeForbidden,
			Message: "access denied: not an active group member",
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve group details",
			Err:     err,
		}
	}
	if g == nil {
		return nil, nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: "group not found or archived",
		}
	}

	members, err := u.repo.ListGroupMembers(ctx, groupID, string(MemberStatusActive))
	if err != nil {
		return nil, nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve group members",
			Err:     err,
		}
	}

	return g, members, nil
}

// ListUserGroups returns a cursor-paginated list of groups the user actively belongs to.
func (u *UseCase) ListUserGroups(ctx context.Context, userID string, p pagination.Params) (pagination.Response[DetailsResponse], error) {
	if userID == "" {
		return pagination.Response[DetailsResponse]{}, &response.AppError{Type: response.TypeValidation, Message: "user id is required"}
	}
	cursor := pagination.ParseCursor(p.Cursor)
	rows, err := u.repo.ListUserGroupsWithMembers(ctx, userID, p.Limit+1, cursor.LastTime, cursor.LastID)
	if err != nil {
		return pagination.Response[DetailsResponse]{}, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve user groups",
			Err:     err,
		}
	}

	detailsList := make([]DetailsResponse, 0, len(rows))
	for _, row := range rows {
		detailsList = append(detailsList, DetailsResponse(row))
	}

	return pagination.BuildResponse(detailsList, p.Limit, func(g DetailsResponse) string {
		return pagination.EncodeCursor(g.Group.CreatedAt, g.Group.ID)
	}), nil
}

// ListMembers retrieves group members with status filter. Non-active queries require admin role.
func (u *UseCase) ListMembers(ctx context.Context, groupID, statusFilter, actionByUserID string) ([]Member, error) {
	if groupID == "" || actionByUserID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "group id and user id are required",
		}
	}

	statusFilter = strings.ToUpper(strings.TrimSpace(statusFilter))
	switch statusFilter {
	case "", "ALL", string(MemberStatusActive), string(MemberStatusPending), string(MemberStatusRejected):
	default:
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "status must be one of: ACTIVE, PENDING, REJECTED, ALL",
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
				Message: "only admins can query pending or non-active members",
			}
		}
	} else {
		member, err := u.repo.GetGroupMember(ctx, groupID, actionByUserID)
		if err != nil {
			return nil, &response.AppError{Type: response.TypeInternal, Message: "failed to verify member role", Err: err}
		}
		if member == nil || member.Status != MemberStatusActive {
			return nil, &response.AppError{Type: response.TypeForbidden, Message: "access denied: not an active group member"}
		}
	}

	dbFilter := statusFilter
	if dbFilter == "ALL" {
		dbFilter = ""
	}

	members, err := u.repo.ListGroupMembers(ctx, groupID, dbFilter)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve group members",
			Err:     err,
		}
	}
	return members, nil
}

// AddMember adds a new user to the group. Requires requester to be an admin.
func (u *UseCase) AddMember(ctx context.Context, groupID, targetUserID, actionByUserID string) error {
	if groupID == "" || targetUserID == "" || actionByUserID == "" {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: "group id, target user id, and action user id are required",
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve group details",
			Err:     err,
		}
	}
	if g == nil {
		return &response.AppError{
			Type:    response.TypeNotFound,
			Message: "group not found",
		}
	}

	isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return &response.AppError{
			Type:    response.TypeForbidden,
			Message: "only admins can add members to the group",
		}
	}

	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.AddGroupMember(txCtx, groupID, targetUserID, MemberRoleMember, MemberStatusActive); err != nil {
			return err
		}

		members, err := u.repo.ListGroupMembers(txCtx, groupID, string(MemberStatusActive))
		if err != nil {
			return err
		}
		var targetMember Member
		for _, m := range members {
			if m.UserID == targetUserID {
				targetMember = m
				break
			}
		}

		payload := activity.MemberPayload{
			Member: targetMember,
		}
		err = u.activity.LogEvent(
			txCtx, actionByUserID, &groupID, nil,
			activity.NewMemberAddedEvent(targetUserID, payload),
		)
		return err
	})
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to add member",
			Err:     err,
		}
	}

	return nil
}

// RemoveMember removes a member from the group.
func (u *UseCase) RemoveMember(ctx context.Context, groupID, targetUserID, actionByUserID string) error {
	if groupID == "" || targetUserID == "" || actionByUserID == "" {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: "group id, target user id, and action user id are required",
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve group details",
			Err:     err,
		}
	}
	if g == nil {
		return &response.AppError{
			Type:    response.TypeNotFound,
			Message: "group not found",
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
				Message: "only admins can remove other members from the group",
			}
		}
	}

	targetMember, err := u.repo.GetGroupMember(ctx, groupID, targetUserID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to verify member details",
			Err:     err,
		}
	}
	if targetMember == nil {
		return &response.AppError{
			Type:    response.TypeNotFound,
			Message: "member not found in group",
		}
	}

	if targetMember.Role == MemberRoleAdmin {
		members, err := u.repo.ListGroupMembers(ctx, groupID, string(MemberStatusActive))
		if err != nil {
			return &response.AppError{
				Type:    response.TypeInternal,
				Message: "failed to list group members",
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
				Message: "cannot remove the sole admin of a group",
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
			Message: "failed to remove member",
			Err:     err,
		}
	}

	return nil
}

// UpdateMemberRole updates a member's role (admin or member).
func (u *UseCase) UpdateMemberRole(ctx context.Context, groupID, targetUserID string, newRole MemberRole, actionByUserID string) error {
	if groupID == "" || targetUserID == "" || newRole == "" || actionByUserID == "" {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: "group id, target user id, role, and action user id are required",
		}
	}

	if newRole != MemberRoleAdmin && newRole != MemberRoleMember {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: "role must be 'ADMIN' or 'MEMBER'",
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve group details",
			Err:     err,
		}
	}
	if g == nil {
		return &response.AppError{
			Type:    response.TypeNotFound,
			Message: "group not found",
		}
	}

	isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return &response.AppError{
			Type:    response.TypeForbidden,
			Message: "only admins can update member roles",
		}
	}
	targetMember, err := u.repo.GetGroupMember(ctx, groupID, targetUserID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to verify member details",
			Err:     err,
		}
	}
	if targetMember == nil {
		return &response.AppError{
			Type:    response.TypeNotFound,
			Message: "member not found in group",
		}
	}

	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.UpdateGroupMemberRole(txCtx, groupID, targetUserID, newRole); err != nil {
			return err
		}

		member, err := u.repo.GetGroupMember(txCtx, groupID, targetUserID)
		if err != nil {
			return err
		}
		payload := activity.MemberPayload{
			Member: member,
		}
		err = u.activity.LogEvent(
			txCtx, actionByUserID, &groupID, nil,
			activity.NewMemberRoleUpdatedEvent(targetUserID, string(newRole), payload),
		)
		return err
	})
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to update member role",
			Err:     err,
		}
	}

	return nil
}

// UpdateGroup updates group name, description, and admin approval requirement.
func (u *UseCase) UpdateGroup(ctx context.Context, groupID, name, description string, requireAdminApproval bool, actionByUserID string) (*Group, error) {
	if groupID == "" || name == "" || actionByUserID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "group id, name, and action user id are required",
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve group details",
			Err:     err,
		}
	}
	if g == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: "group not found",
		}
	}

	isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, &response.AppError{
			Type:    response.TypeForbidden,
			Message: "only admins can update group details",
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
			Message: "failed to update group",
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
			Message: "group id and action user id are required",
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve group details",
			Err:     err,
		}
	}
	if g == nil {
		return &response.AppError{
			Type:    response.TypeNotFound,
			Message: "group not found",
		}
	}

	isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return &response.AppError{
			Type:    response.TypeForbidden,
			Message: "only admins can archive the group",
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
			Message: "failed to archive group",
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
			Message: "invite code and user id are required",
		}
	}

	g, err := u.repo.GetByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to look up group by invite code",
			Err:     err,
		}
	}
	if g == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: "invalid or expired invite code",
		}
	}

	if g.InviteCodeExpiresAt != nil && g.InviteCodeExpiresAt.Before(time.Now()) {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "invite code has expired",
		}
	}

	existing, err := u.repo.GetGroupMember(ctx, g.ID, userID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to verify membership status",
			Err:     err,
		}
	}
	if existing != nil {
		if existing.Status == MemberStatusActive {
			return &JoinResponse{Status: MemberStatusActive, Group: g}, nil
		}
		if existing.Status == MemberStatusPending {
			return &JoinResponse{Status: MemberStatusPending, Message: "Join request submitted for admin approval"}, nil
		}
		if existing.Status == MemberStatusRejected {
			return nil, &response.AppError{
				Type:    response.TypeForbidden,
				Message: "your join request was rejected by an admin",
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
			members, err := u.repo.ListGroupMembers(txCtx, g.ID, string(MemberStatusActive))
			if err == nil {
				for _, m := range members {
					if m.Role == MemberRoleAdmin && u.notification != nil {
						_ = u.notification.CreateAlert(
							txCtx, m.UserID, &userID, nil,
							"Join Request Pending",
							fmt.Sprintf("A user requested to join group %s", g.Name),
						)
					}
				}
			}
			return nil
		}

		members, err := u.repo.ListGroupMembers(txCtx, g.ID, string(MemberStatusActive))
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
			Message: "failed to join group",
			Err:     err,
		}
	}

	if targetStatus == MemberStatusPending {
		return &JoinResponse{Status: MemberStatusPending, Message: "Join request submitted for admin approval"}, nil
	}
	return &JoinResponse{Status: MemberStatusActive, Group: g}, nil
}

// DecideJoinRequest approves or rejects a pending join request. Admin only.
func (u *UseCase) DecideJoinRequest(ctx context.Context, groupID, targetUserID string, action JoinRequestAction, adminUserID string) error {
	if groupID == "" || targetUserID == "" || action == "" || adminUserID == "" {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: "group id, target user id, action, and admin user id are required",
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve group details",
			Err:     err,
		}
	}
	if g == nil {
		return &response.AppError{
			Type:    response.TypeNotFound,
			Message: "group not found",
		}
	}

	isAdmin, err := u.checkIsAdmin(ctx, groupID, adminUserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return &response.AppError{
			Type:    response.TypeForbidden,
			Message: "only admins can decide join requests",
		}
	}

	var newStatus MemberStatus
	switch action {
	case JoinRequestActionApprove:
		newStatus = MemberStatusActive
	case JoinRequestActionReject:
		newStatus = MemberStatusRejected
	default:
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: "action must be 'APPROVE' or 'REJECT'",
		}
	}

	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.UpdateMemberStatus(txCtx, groupID, targetUserID, newStatus); err != nil {
			return err
		}

		if newStatus == MemberStatusActive {
			member, err := u.repo.GetGroupMember(txCtx, groupID, targetUserID)
			if err == nil && member != nil {
				payload := activity.MemberPayload{
					Member: *member,
				}
				_ = u.activity.LogEvent(
					txCtx, adminUserID, &groupID, nil,
					activity.NewMemberJoinedEvent(targetUserID, payload),
				)
			}
		}
		if u.notification != nil {
			title := "Join Request Approved"
			content := "Your request to join the group was approved."
			if newStatus == MemberStatusRejected {
				title = "Join Request Rejected"
				content = "Your request to join the group was declined."
			}
			_ = u.notification.CreateAlert(txCtx, targetUserID, &adminUserID, nil, title, content)
		}
		return nil
	})
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to decide join request",
			Err:     err,
		}
	}

	return nil
}

// ResetInviteCode generates a new invite code with a 7-day expiration. Admin only.
func (u *UseCase) ResetInviteCode(ctx context.Context, groupID, adminUserID string) (*Group, error) {
	if groupID == "" || adminUserID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "group id and admin user id are required",
		}
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve group details",
			Err:     err,
		}
	}
	if g == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: "group not found",
		}
	}

	isAdmin, err := u.checkIsAdmin(ctx, groupID, adminUserID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, &response.AppError{
			Type:    response.TypeForbidden,
			Message: "only admins can reset group invite code",
		}
	}

	code := newInviteCode()
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
			Message: "failed to reset invite code",
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
			Message: "invite code is required",
		}
	}

	preview, err := u.repo.GetPreviewByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve group preview",
			Err:     err,
		}
	}
	if preview == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: "invalid or expired invite code",
		}
	}

	return preview, nil
}

func (u *UseCase) checkIsAdmin(ctx context.Context, groupID, userID string) (bool, error) {
	member, err := u.repo.GetGroupMember(ctx, groupID, userID)
	if err != nil {
		return false, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to verify member role",
			Err:     err,
		}
	}
	if member == nil || member.Status != MemberStatusActive {
		return false, nil
	}
	return member.Role == MemberRoleAdmin, nil
}
