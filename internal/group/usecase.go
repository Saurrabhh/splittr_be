package group

import (
	"context"
	"errors"
	"fmt"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/notification"
	"github.com/google/uuid"
)

// Repository defines the storage contract for groups and memberships.
type Repository interface {
	GetByID(ctx context.Context, id string) (*Group, error)
	GetByInviteCode(ctx context.Context, inviteCode string) (*Group, error)
	GetGroupMember(ctx context.Context, groupID, userID string) (*GroupMember, error)
	ListGroupMembers(ctx context.Context, groupID string) ([]GroupMember, error)
	ListUserGroups(ctx context.Context, userID string) ([]Group, error)
	CreateGroup(ctx context.Context, g *Group) error
	Update(ctx context.Context, g *Group) error
	Archive(ctx context.Context, id string) error
	AddGroupMember(ctx context.Context, groupID, userID, role string) error
	RemoveGroupMember(ctx context.Context, groupID, userID string) error
	UpdateGroupMemberRole(ctx context.Context, groupID, userID, role string) error
}

type ActivityLogger interface {
	LogActivity(ctx context.Context, actorID string, groupID *string, actionType string, description string, visibleToUserIDs []string) (*activity.Activity, error)
}

type NotificationSender interface {
	CreateAlert(ctx context.Context, userID string, actorID *string, activityID *string, title, content string) (*notification.Notification, error)
}

// Usecase manages business workflows for the group domain.
type Usecase struct {
	repo         Repository
	tx           db.Transactor
	activity     ActivityLogger
	notification NotificationSender
}

// NewUsecase instantiates a new Usecase.
func NewUsecase(repo Repository, tx db.Transactor, activitySvc ActivityLogger, notificationSvc NotificationSender) *Usecase {
	return &Usecase{
		repo:         repo,
		tx:           tx,
		activity:     activitySvc,
		notification: notificationSvc,
	}
}

// CreateGroup creates a new group and adds the creator as the first admin.
func (u *Usecase) CreateGroup(ctx context.Context, name, description string, creatorID string) (*Group, error) {
	if name == "" {
		return nil, errors.New("group name is required")
	}
	if creatorID == "" {
		return nil, errors.New("creator ID is required")
	}

	newGroup := &Group{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedBy: &creatorID,
	}
	if description != "" {
		newGroup.Description = &description
	}

	// Generate invite code
	inviteCode := "invite-" + uuid.New().String()[:8]
	newGroup.InviteCode = &inviteCode

	err := u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.CreateGroup(txCtx, newGroup); err != nil {
			return err
		}
		if err := u.repo.AddGroupMember(txCtx, newGroup.ID, creatorID, "admin"); err != nil {
			return err
		}
		_, err := u.activity.LogActivity(txCtx, creatorID, &newGroup.ID, "GROUP_CREATED", "created the group", nil)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("create group transaction: %w", err)
	}

	return newGroup, nil
}

// GetGroupDetails retrieves a group and its members, verifying the requester belongs to it.
func (u *Usecase) GetGroupDetails(ctx context.Context, groupID, userID string) (*Group, []GroupMember, error) {
	if groupID == "" || userID == "" {
		return nil, nil, errors.New("group ID and user ID are required")
	}

	// Access control: check if requester is a member of the group
	member, err := u.repo.GetGroupMember(ctx, groupID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("verify membership: %w", err)
	}
	if member == nil {
		return nil, nil, errors.New("access denied: not a group member")
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, nil, fmt.Errorf("get group metadata: %w", err)
	}
	if g == nil {
		return nil, nil, errors.New("group not found or archived")
	}

	members, err := u.repo.ListGroupMembers(ctx, groupID)
	if err != nil {
		return nil, nil, fmt.Errorf("list group members: %w", err)
	}

	return g, members, nil
}

// ListUserGroups returns all groups the user is a member of.
func (u *Usecase) ListUserGroups(ctx context.Context, userID string) ([]Group, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	return u.repo.ListUserGroups(ctx, userID)
}

// AddMember adds a new user to the group. Requires requester to be an admin.
func (u *Usecase) AddMember(ctx context.Context, groupID, targetUserID, actionByUserID string) error {
	if groupID == "" || targetUserID == "" || actionByUserID == "" {
		return errors.New("missing required fields")
	}

	// Verify requester is admin
	isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("only admins can add members to the group")
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if g == nil {
		return errors.New("group not found")
	}

	return u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.AddGroupMember(txCtx, groupID, targetUserID, "member"); err != nil {
			return err
		}

		desc := fmt.Sprintf("added user %s to the group", targetUserID)
		act, err := u.activity.LogActivity(txCtx, actionByUserID, &groupID, "MEMBER_ADDED", desc, nil)
		if err != nil {
			return err
		}

		_, err = u.notification.CreateAlert(
			txCtx,
			targetUserID,
			&actionByUserID,
			&act.ID,
			"Added to Group",
			fmt.Sprintf("You were added to group %s", g.Name),
		)
		return err
	})
}

// RemoveMember removes a user from the group.
// A user can remove themselves (leave). Admins can remove anyone.
func (u *Usecase) RemoveMember(ctx context.Context, groupID, targetUserID, actionByUserID string) error {
	if groupID == "" || targetUserID == "" || actionByUserID == "" {
		return errors.New("missing required fields")
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if g == nil {
		return errors.New("group not found")
	}

	// Check permissions
	isSelf := targetUserID == actionByUserID
	isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
	if err != nil {
		return err
	}

	if !isSelf && !isAdmin {
		return errors.New("unauthorized: only admins can remove other members")
	}

	// Safe-guard: Check if we are removing an admin, and ensure there's at least one admin left.
	targetMember, err := u.repo.GetGroupMember(ctx, groupID, targetUserID)
	if err != nil {
		return err
	}
	if targetMember == nil {
		return errors.New("user is not a member of the group")
	}

	if targetMember.Role == "admin" {
		members, err := u.repo.ListGroupMembers(ctx, groupID)
		if err != nil {
			return err
		}

		adminCount := 0
		for _, m := range members {
			if m.Role == "admin" {
				adminCount++
			}
		}

		if adminCount == 1 {
			if len(members) > 1 {
				return errors.New("cannot remove the sole admin of a group containing other members. Promote another user to admin first")
			}
			return u.tx.RunInTx(ctx, func(txCtx context.Context) error {
				if err := u.repo.RemoveGroupMember(txCtx, groupID, targetUserID); err != nil {
					return err
				}
				_, err := u.activity.LogActivity(txCtx, actionByUserID, &groupID, "MEMBER_LEFT", "left the group", nil)
				if err != nil {
					return err
				}
				return u.repo.Archive(txCtx, groupID)
			})
		}
	}

	return u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.RemoveGroupMember(txCtx, groupID, targetUserID); err != nil {
			return err
		}

		actionType := "MEMBER_LEFT"
		desc := "left the group"
		if !isSelf {
			actionType = "MEMBER_KICKED"
			desc = fmt.Sprintf("removed user %s from the group", targetUserID)
		}

		act, err := u.activity.LogActivity(txCtx, actionByUserID, &groupID, actionType, desc, nil)
		if err != nil {
			return err
		}

		if !isSelf {
			_, _ = u.notification.CreateAlert(
				txCtx,
				targetUserID,
				&actionByUserID,
				&act.ID,
				"Removed from Group",
				fmt.Sprintf("You were removed from group %s by an admin", g.Name),
			)
		}
		return nil
	})
}

// UpdateMemberRole changes a member's role (admin <-> member).
func (u *Usecase) UpdateMemberRole(ctx context.Context, groupID, targetUserID, role, actionByUserID string) error {
	if role != "admin" && role != "member" {
		return errors.New("invalid role: must be admin or member")
	}

	g, err := u.repo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if g == nil {
		return errors.New("group not found")
	}

	// Verify requester is admin
	isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("unauthorized: only admins can manage member roles")
	}

	// Check if demoting the last admin to member
	targetMember, err := u.repo.GetGroupMember(ctx, groupID, targetUserID)
	if err != nil {
		return err
	}
	if targetMember == nil {
		return errors.New("user is not a member of the group")
	}

	if targetMember.Role == "admin" && role == "member" {
		members, err := u.repo.ListGroupMembers(ctx, groupID)
		if err != nil {
			return err
		}

		adminCount := 0
		for _, m := range members {
			if m.Role == "admin" {
				adminCount++
			}
		}

		if adminCount == 1 {
			return errors.New("cannot demote the sole admin of the group. Promote another user to admin first")
		}
	}

	return u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.UpdateGroupMemberRole(txCtx, groupID, targetUserID, role); err != nil {
			return err
		}

		desc := fmt.Sprintf("updated user %s's role to %s", targetUserID, role)
		act, err := u.activity.LogActivity(txCtx, actionByUserID, &groupID, "MEMBER_ROLE_UPDATED", desc, nil)
		if err != nil {
			return err
		}

		_, err = u.notification.CreateAlert(
			txCtx,
			targetUserID,
			&actionByUserID,
			&act.ID,
			"Role Updated",
			fmt.Sprintf("Your role in group %s was updated to %s", g.Name, role),
		)
		return err
	})
}

// ArchiveGroup soft-deletes the group. Requires requester to be an admin.
func (u *Usecase) ArchiveGroup(ctx context.Context, groupID, actionByUserID string) error {
	isAdmin, err := u.checkIsAdmin(ctx, groupID, actionByUserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("unauthorized: only admins can archive the group")
	}

	return u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.Archive(txCtx, groupID); err != nil {
			return err
		}
		_, err := u.activity.LogActivity(txCtx, actionByUserID, &groupID, "GROUP_ARCHIVED", "archived the group", nil)
		return err
	})
}

// JoinGroup matches a group by invite code and adds the user as a member.
func (u *Usecase) JoinGroup(ctx context.Context, inviteCode, userID string) (*Group, error) {
	if inviteCode == "" || userID == "" {
		return nil, errors.New("invite code and user ID are required")
	}

	g, err := u.repo.GetByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.New("invalid or expired invite code")
	}

	// Check if already a member
	existing, err := u.repo.GetGroupMember(ctx, g.ID, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return g, nil
	}

	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.AddGroupMember(txCtx, g.ID, userID, "member"); err != nil {
			return err
		}
		_, err := u.activity.LogActivity(txCtx, userID, &g.ID, "MEMBER_JOINED", "joined the group via invite code", nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	return g, nil
}

// checkIsAdmin is a helper to verify a user's admin status in a group.
func (u *Usecase) checkIsAdmin(ctx context.Context, groupID, userID string) (bool, error) {
	member, err := u.repo.GetGroupMember(ctx, groupID, userID)
	if err != nil {
		return false, fmt.Errorf("check membership: %w", err)
	}
	if member == nil {
		return false, nil
	}
	return member.Role == "admin", nil
}
