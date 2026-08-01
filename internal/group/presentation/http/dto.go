package http

import (
	"github.com/Saurrabhh/splittr_be/internal/group/domain"
)

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
