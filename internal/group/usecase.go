package group

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/notification"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/google/uuid"
)

// Repository defines the storage contract for groups and memberships.
type Repository interface {
	GetByID(ctx context.Context, id string) (*Group, error)
	GetByInviteCode(ctx context.Context, inviteCode string) (*Group, error)
	GetPreviewByInviteCode(ctx context.Context, inviteCode string) (*Preview, error)
	GetGroupMember(ctx context.Context, groupID, userID string) (*Member, error)
	ListGroupMembers(ctx context.Context, groupID string, status string) ([]Member, error)
	ListUserGroupsWithMembers(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]DetailsResponse, error)
	CreateGroup(ctx context.Context, g *Group) error
	Update(ctx context.Context, g *Group) error
	ResetInviteCode(ctx context.Context, groupID, newInviteCode string, expiresAt time.Time) (*Group, error)
	Archive(ctx context.Context, id string) error
	AddGroupMember(ctx context.Context, groupID, userID, role, status string) error
	UpdateMemberStatus(ctx context.Context, groupID, userID, status string) error
	RemoveGroupMember(ctx context.Context, groupID, userID string) error
	UpdateGroupMemberRole(ctx context.Context, groupID, userID, role string) error
}

type ActivityLogger interface {
	LogEvent(
		ctx context.Context,
		actorID string,
		groupID *string,
		visibleToUserIDs []string,
		event activity.Event,
	) (*activity.Activity, error)
}

type NotificationSender interface {
	CreateAlert(ctx context.Context, userID string, actorID *string, activityID *string, title, content string) (*notification.Notification, error)
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
			Message: "creator ID is required",
		}
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	newGroup := &Group{
		ID:                   uuid.New().String(),
		Name:                 name,
		InviteCodeExpiresAt:  &expiresAt,
		RequireAdminApproval: requireAdminApproval,
		CreatedBy:            &creatorID,
	}
	if description != "" {
		newGroup.Description = &description
	}

	// Generate invite code
	newGroup.InviteCode = new("invite-" + uuid.New().String()[:8])

	err := u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.CreateGroup(txCtx, newGroup); err != nil {
			return err
		}
		if err := u.repo.AddGroupMember(txCtx, newGroup.ID, creatorID, "admin", string(MemberStatusActive)); err != nil {
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
		_, err = u.activity.LogEvent(
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
			Message: "group ID and user ID are required",
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
	if member == nil || member.Status != string(MemberStatusActive) {
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
		return pagination.Response[DetailsResponse]{}, &response.AppError{Type: response.TypeValidation, Message: "user ID is required"}
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
	return pagination.BuildResponse(rows, p.Limit, func(g DetailsResponse) string {
		return pagination.EncodeCursor(g.Group.CreatedAt, g.Group.ID)
	}), nil
}

// ListMembers retrieves group members with status filter. Non-active queries require admin role.
func (u *UseCase) ListMembers(ctx context.Context, groupID, statusFilter, actionByUserID string) ([]Member, error) {
	if groupID == "" || actionByUserID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "group ID and user ID are required",
		}
	}

	statusFilter = strings.ToUpper(strings.TrimSpace(statusFilter))
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
		// Active check for normal member
		member, err := u.repo.GetGroupMember(ctx, groupID, actionByUserID)
		if err != nil {
			return nil, &response.AppError{Type: response.TypeInternal, Message: "failed to verify member role", Err: err}
		}
		if member == nil || member.Status != string(MemberStatusActive) {
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
			Message: "missing required fields",
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

	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.AddGroupMember(txCtx, groupID, targetUserID, "member", string(MemberStatusActive)); err != nil {
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
		_, err = u.activity.LogEvent(
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
			Message: "missing required fields",
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

	if targetMember.Role == "admin" {
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
			if m.Role == "admin" {
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

		_, err = u.activity.LogEvent(
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
func (u *UseCase) UpdateMemberRole(ctx context.Context, groupID, targetUserID, newRole, actionByUserID string) error {
	if groupID == "" || targetUserID == "" || newRole == "" || actionByUserID == "" {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: "missing required fields",
		}
	}

	if newRole != "admin" && newRole != "member" {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: "role must be 'admin' or 'member'",
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
		_, err = u.activity.LogEvent(
			txCtx, actionByUserID, &groupID, nil,
			activity.NewMemberRoleUpdatedEvent(targetUserID, newRole, payload),
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
			Message: "group ID, name, and action user ID are required",
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
		_, err = u.activity.LogEvent(
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
			Message: "group ID and user ID are required",
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

	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.Archive(txCtx, groupID); err != nil {
			return err
		}

		payload := activity.GroupPayload{
			Group: g,
		}
		_, err = u.activity.LogEvent(
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
			Message: "invite code and user ID are required",
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
		if existing.Status == string(MemberStatusActive) {
			return &JoinResponse{Status: string(MemberStatusActive), Group: g}, nil
		}
		if existing.Status == string(MemberStatusPending) {
			return &JoinResponse{Status: string(MemberStatusPending), Message: "Join request submitted for admin approval"}, nil
		}
		if existing.Status == string(MemberStatusRejected) {
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
		if err := u.repo.AddGroupMember(txCtx, g.ID, userID, "member", string(targetStatus)); err != nil {
			return err
		}

		if targetStatus == MemberStatusPending {
			// Notify group admins
			members, err := u.repo.ListGroupMembers(txCtx, g.ID, string(MemberStatusActive))
			if err == nil {
				for _, m := range members {
					if m.Role == "admin" && u.notification != nil {
						_, _ = u.notification.CreateAlert(
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
		_, err = u.activity.LogEvent(
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
		return &JoinResponse{Status: string(MemberStatusPending), Message: "Join request submitted for admin approval"}, nil
	}
	return &JoinResponse{Status: string(MemberStatusActive), Group: g}, nil
}

// DecideJoinRequest approves or rejects a pending join request. Admin only.
func (u *UseCase) DecideJoinRequest(ctx context.Context, groupID, targetUserID, action, adminUserID string) error {
	if groupID == "" || targetUserID == "" || action == "" || adminUserID == "" {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: "group ID, target user ID, action, and admin user ID are required",
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

	actionUpper := strings.ToUpper(action)
	var newStatus MemberStatus
	if actionUpper == "APPROVE" {
		newStatus = MemberStatusActive
	} else if actionUpper == "REJECT" {
		newStatus = MemberStatusRejected
	} else {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: "action must be 'APPROVE' or 'REJECT'",
		}
	}

	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.UpdateMemberStatus(txCtx, groupID, targetUserID, string(newStatus)); err != nil {
			return err
		}

		if newStatus == MemberStatusActive {
			member, err := u.repo.GetGroupMember(txCtx, groupID, targetUserID)
			if err == nil && member != nil {
				payload := activity.MemberPayload{
					Member: *member,
				}
				_, _ = u.activity.LogEvent(
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
			_, _ = u.notification.CreateAlert(txCtx, targetUserID, &adminUserID, nil, title, content)
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
			Message: "group ID and admin user ID are required",
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

	newInviteCode := "invite-" + uuid.New().String()[:8]
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	updatedGroup, err := u.repo.ResetInviteCode(ctx, groupID, newInviteCode, expiresAt)
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
		return nil, err
	}
	if preview == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: "invalid or expired invite code",
		}
	}

	return preview, nil
}

// checkIsAdmin is a helper to verify a user's admin status in a group.
func (u *UseCase) checkIsAdmin(ctx context.Context, groupID, userID string) (bool, error) {
	member, err := u.repo.GetGroupMember(ctx, groupID, userID)
	if err != nil {
		return false, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to verify member role",
			Err:     err,
		}
	}
	if member == nil || member.Status != string(MemberStatusActive) {
		return false, nil
	}
	return member.Role == "admin", nil
}
