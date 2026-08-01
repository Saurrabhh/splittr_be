package http

import (
	"github.com/Saurrabhh/splittr_be/internal/group/domain"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
)

// DetailsResponse is the canonical shape for any endpoint or feed payload that returns group data.
type DetailsResponse struct {
	Group   domain.Group    `json:"group"`
	Members []domain.Member `json:"members"`
}

// ListGroupsResponse represents the paginated groups list response.
type ListGroupsResponse struct {
	Data       []DetailsResponse `json:"data"`
	Pagination pagination.Meta   `json:"pagination"`
}

type CreateGroupRequest struct {
	Name                 string `json:"name"`
	Description          string `json:"description"`
	RequireAdminApproval bool   `json:"requireAdminApproval"`
}

type JoinGroupRequest struct {
	InviteCode string `json:"inviteCode"`
}

type AddMemberRequest struct {
	UserID string `json:"userId"`
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}
