package http

import (
	"github.com/Saurrabhh/splittr_be/internal/group/domain"
)

type CreateGroupRequest struct {
	Name                 string `json:"name" validate:"required"`
	Description          string `json:"description"`
	RequireAdminApproval bool   `json:"requireAdminApproval"`
} // @name Group.CreateRequest

type JoinGroupRequest struct {
	InviteCode string `json:"inviteCode" validate:"required"`
} // @name Group.JoinRequest

type AddMembersRequest struct {
	UserIDs []string `json:"userIds" validate:"required,min=1"`
} // @name Group.AddMembersRequest

type UpdateRoleRequest struct {
	Role domain.MemberRole `json:"role"`
} // @name Group.UpdateRoleRequest
