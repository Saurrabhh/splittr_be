package domain

import (
	"context"
	"time"
)

// Repository defines the storage contract for groups and memberships.
type Repository interface {
	GetByID(ctx context.Context, id string) (*Group, error)
	GetByInviteCode(ctx context.Context, inviteCode string) (*Group, error)
	GetPreviewByInviteCode(ctx context.Context, inviteCode string) (*Preview, error)
	GetGroupMember(ctx context.Context, groupID, userID string) (*Member, error)
	ListGroupMembers(ctx context.Context, groupID string, status MemberStatus) ([]Member, error)
	ListUserGroupsWithMembers(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]GroupWithMembers, error)
	CreateGroup(ctx context.Context, g *Group) error
	Update(ctx context.Context, g *Group) error
	UpdateIcon(ctx context.Context, groupID string, iconURL string) (*Group, error)
	ResetInviteCode(ctx context.Context, groupID, newInviteCode string, expiresAt time.Time) (*Group, error)
	Archive(ctx context.Context, id string) error
	AddGroupMembers(ctx context.Context, groupID string, userIDs []string, role MemberRole, status MemberStatus) ([]Member, error)
	UpdateMemberStatus(ctx context.Context, groupID, userID string, status MemberStatus) error
	RemoveGroupMember(ctx context.Context, groupID, userID string) error
	UpdateGroupMemberRole(ctx context.Context, groupID, userID string, role MemberRole) error
	SyncGroupsBySequence(ctx context.Context, lastVersion int64, userID string, limit int32) ([]GroupWithMembers, error)

	GetGroupTombstonesBySequence(ctx context.Context, lastVersion int64, userID string, limit int32) ([]Tombstone, error)
}

// GroupWithMembers wraps a Group entity with its active members.
type GroupWithMembers struct {
	Group   Group
	Members []Member
}

// Tombstone represents a deleted entity for sync.
type Tombstone struct {
	EntityID    string
	SyncVersion int64
}

