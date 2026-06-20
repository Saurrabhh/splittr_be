package group

import (
	"time"
)

// Group represents a bill-splitting group.
type Group struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	InviteCode  *string    `json:"inviteCode,omitempty"`
	CreatedBy   *string    `json:"createdBy,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	ArchivedAt  *time.Time `json:"archivedAt,omitempty"`
}

// GroupMember represents a user's membership details in a group, enriched with basic user details.
type GroupMember struct {
	GroupID  string    `json:"groupId"`
	UserID   string    `json:"userId"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joinedAt"`
	Name     string    `json:"name"`
	Email    *string   `json:"email,omitempty"`
	Phone    *string   `json:"phone,omitempty"`
}
