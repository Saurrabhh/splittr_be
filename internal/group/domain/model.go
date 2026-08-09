package domain

import "time"

type MemberStatus string // @name Group.MemberStatus

const (
	MemberStatusActive   MemberStatus = "ACTIVE"
	MemberStatusPending  MemberStatus = "PENDING"
	MemberStatusRejected MemberStatus = "REJECTED"
)

type MemberRole string // @name Group.MemberRole

const (
	MemberRoleAdmin  MemberRole = "ADMIN"
	MemberRoleMember MemberRole = "MEMBER"
)

// Group represents a bill-splitting group.
type Group struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Description          *string    `json:"description,omitempty"`
	InviteCode           *string    `json:"inviteCode,omitempty"`
	InviteCodeExpiresAt  *time.Time `json:"inviteCodeExpiresAt,omitempty"`
	RequireAdminApproval bool       `json:"requireAdminApproval"`
	CreatedBy            *string    `json:"createdBy,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	ArchivedAt           *time.Time `json:"archivedAt,omitempty"`
	Members              []Member   `json:"members,omitempty"`
} // @name Group.Group

// Member represents a user's membership details in a group, enriched with basic user details.
type Member struct {
	GroupID  string       `json:"groupId"`
	UserID   string       `json:"userId"`
	Role     MemberRole   `json:"role"`
	Status   MemberStatus `json:"status"`
	JoinedAt time.Time    `json:"joinedAt"`
	Name     string       `json:"name"`
	Email    *string      `json:"email,omitempty"`
	Phone    *string      `json:"phone,omitempty"`
} // @name Group.Member

// Preview represents a summary of a group's details before joining.
type Preview struct {
	Name                 string  `json:"name"`
	Description          *string `json:"description,omitempty"`
	MemberCount          int64   `json:"memberCount"`
	CreatorName          string  `json:"creatorName"`
	RequireAdminApproval bool    `json:"requireAdminApproval"`
} // @name Group.Preview

// JoinRequestAction is the action an admin can take on a pending join request.
type JoinRequestAction string // @name Group.JoinRequestAction

const (
	JoinRequestActionApprove JoinRequestAction = "APPROVE"
	JoinRequestActionReject  JoinRequestAction = "REJECT"
)

// DecideJoinRequestPayload is the request body for approving or rejecting a join request.
type DecideJoinRequestPayload struct {
	Action JoinRequestAction `json:"action"`
} // @name Group.DecideJoinRequestPayload
