package http

import (
	"github.com/Saurrabhh/splittr_be/internal/group/domain"
)

type CreateGroupRequest struct {
	Name                 string `json:"name" validate:"required"`
	Description          string `json:"description"`
	RequireAdminApproval bool   `json:"requireAdminApproval"`
} // @name CreateGroupRequest

type JoinGroupRequest struct {
	InviteCode string `json:"inviteCode" validate:"required"`
} // @name JoinGroupRequest

type AddMemberRequest struct {
	UserID string `json:"userId" validate:"required"`
} // @name AddMemberRequest

type UpdateRoleRequest struct {
	Role domain.MemberRole `json:"role"`
} // @name UpdateRoleRequest
