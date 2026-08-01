package http

import (
	"github.com/Saurrabhh/splittr_be/internal/group/domain"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
)

// ListGroupsResponse is the paginated groups list response.
type ListGroupsResponse struct {
	Data       []domain.DetailsResponse `json:"data"`
	Pagination pagination.Meta          `json:"pagination"`
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
	Role domain.MemberRole `json:"role"`
}
