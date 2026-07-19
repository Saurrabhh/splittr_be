package group

import (
	"time"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
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

// Member represents a user's membership details in a group, enriched with basic user details.
type Member struct {
	GroupID  string    `json:"groupId"`
	UserID   string    `json:"userId"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joinedAt"`
	Name     string    `json:"name"`
	Email    *string   `json:"email,omitempty"`
	Phone    *string   `json:"phone,omitempty"`
}

// GroupDetailsResponse is the canonical shape for any endpoint or feed payload that
// returns group data. It embeds Group (all fields inline) and adds the Members array.
// Both the list and detail endpoints return this type so Flutter only needs one class.
type GroupDetailsResponse struct {
	Group
	Members []Member `json:"members"`
}

// ListGroupsResponse represents the paginated groups list response.
type ListGroupsResponse struct {
	Data       []GroupDetailsResponse `json:"data"`
	Pagination pagination.Meta        `json:"pagination"`
}
