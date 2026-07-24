package group

import (
	"time"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
)

type MemberStatus string

const (
	MemberStatusActive   MemberStatus = "ACTIVE"
	MemberStatusPending  MemberStatus = "PENDING"
	MemberStatusRejected MemberStatus = "REJECTED"
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
}

// Member represents a user's membership details in a group, enriched with basic user details.
type Member struct {
	GroupID  string    `json:"groupId"`
	UserID   string    `json:"userId"`
	Role     string    `json:"role"`
	Status   string    `json:"status"`
	JoinedAt time.Time `json:"joinedAt"`
	Name     string    `json:"name"`
	Email    *string   `json:"email,omitempty"`
	Phone    *string   `json:"phone,omitempty"`
}

// DetailsResponse is the canonical shape for any endpoint or feed payload that
// returns group data. It embeds Group (all fields inline) and adds the Members array.
type DetailsResponse struct {
	Group
	Members []Member `json:"members"`
}

// ListGroupsResponse represents the paginated groups list response.
type ListGroupsResponse struct {
	Data       []DetailsResponse `json:"data"`
	Pagination pagination.Meta   `json:"pagination"`
}

// Preview represents a summary of a group's details before joining.
type Preview struct {
	Name                 string  `json:"name"`
	Description          *string `json:"description,omitempty"`
	MemberCount          int64   `json:"memberCount"`
	CreatorName          string  `json:"creatorName"`
	RequireAdminApproval bool    `json:"requireAdminApproval"`
}

// JoinResponse returned when joining a group (either active or pending).
type JoinResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Group   *Group `json:"group,omitempty"`
}

// DecideJoinRequestPayload body for approving or rejecting a join request.
type DecideJoinRequestPayload struct {
	Action string `json:"action"` // "APPROVE" or "REJECT"
}
