package http

import (
	"context"
	"net/http"

	_ "github.com/Saurrabhh/splittr_be/internal/activity"

	"github.com/Saurrabhh/splittr_be/internal/group/domain"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/request"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for group endpoints.
type Handler struct {
	uc *domain.UseCase
}

// NewHandler creates a new Handler instance.
func NewHandler(uc *domain.UseCase) *Handler {
	return &Handler{
		uc: uc,
	}
}

// RegisterRoutes registers the group endpoints on the router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/groups", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/preview", h.Preview)
		r.Post("/join", h.Join)
		r.Get("/", h.List)

		r.Route("/{id}", func(r chi.Router) {

			r.Get("/", h.GetDetails)
			r.Patch("/", h.Update)
			r.Delete("/", h.Archive)
			r.Get("/feed", h.GetFeed)
			r.Post("/icon", h.UploadIcon)
			r.Post("/invite-code/reset", h.ResetInviteCode)

			r.Route("/members", func(r chi.Router) {
				r.Get("/", h.ListMembers)
				r.Post("/", h.AddMembers)
				r.Route("/{userId}", func(r chi.Router) {
					r.Delete("/", h.RemoveMember)
					r.Patch("/role", h.UpdateMemberRole)
					r.Post("/decision", h.DecideJoinRequest)
				})
			})
		})
	})
}

// Create creates a new group.
// @Summary      Create group
// @Description  Create a new bill-splitting group with optional admin approval requirement.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        request body CreateGroupRequest true "Group creation data"
// @Success      201  {object}  domain.Group
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups [post]
// @Security     BearerAuth
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())

	request.Run(w, r, http.StatusCreated, func(ctx context.Context, req CreateGroupRequest) (*domain.Group, error) {
		return h.uc.CreateGroup(ctx, req.Name, req.Description, req.RequireAdminApproval, currUser.ID)
	})
}

// Join joins a group via its invite code.
// @Summary      Join group
// @Description  Join an existing group using its unique invite code.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        request body JoinGroupRequest true "Group join data"
// @Success      200  {object}  domain.JoinResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/join [post]
// @Security     BearerAuth
func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())

	request.Run(w, r, http.StatusOK, func(ctx context.Context, req JoinGroupRequest) (*domain.JoinResponse, error) {
		return h.uc.JoinGroup(ctx, req.InviteCode, currUser.ID)
	})
}

// Preview retrieves group basic info before joining.
// @Summary      Get group preview
// @Description  Get group details using invite code.
// @Tags         groups
// @Produce      json
// @Param        inviteCode   query  string  true  "Group invite code"
// @Success      200  {object}  domain.Preview
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/preview [get]
// @Security     BearerAuth
func (h *Handler) Preview(w http.ResponseWriter, r *http.Request) {
	inviteCode, ok := request.QueryParam(w, r, "inviteCode")
	if !ok {
		return
	}

	preview, err := h.uc.GetGroupPreview(r.Context(), inviteCode)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, preview)
}

// List retrieves all groups the user is a member of.
// @Summary      List groups
// @Description  Retrieve a cursor-paginated list of groups the current user belongs to.
// @Tags         groups
// @Produce      json
// @Param        limit   query  int     false  "Items per page (max 100, default 20)"
// @Param        cursor  query  string  false  "Opaque cursor token from a previous response"
// @Success      200  {object}  pagination.Response[domain.Group]
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups [get]
// @Security     BearerAuth
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())
	p := pagination.ParseParams(r, 20, 100)

	result, err := h.uc.ListUserGroups(r.Context(), currUser.ID, p)

	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

// GetDetails returns the group metadata and its members list.
// @Summary      Get group details
// @Description  Get a group's metadata and a list of all its members.
// @Tags         groups
// @Produce      json
// @Param        id path string true "Group ID"
// @Success      200  {object}  domain.Group
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/{id} [get]
// @Security     BearerAuth
func (h *Handler) GetDetails(w http.ResponseWriter, r *http.Request) {
	groupID, ok := request.URLParam(w, r, "id")
	if !ok {
		return
	}

	currUser := user.MustFrom(r.Context())

	g, err := h.uc.GetGroupDetails(r.Context(), groupID, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, g)
}

// Update updates group name, description, and admin approval requirement.
// @Summary      Update group
// @Description  Update group details. Admin privileges required.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id path string true "Group ID"
// @Param        request body UpdateGroupRequest true "Group update payload"
// @Success      200  {object}  domain.Group
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/{id} [patch]
// @Security     BearerAuth
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	groupID, ok := request.URLParam(w, r, "id")
	if !ok {
		return
	}
	currUser := user.MustFrom(r.Context())

	request.Run(w, r, http.StatusOK, func(ctx context.Context, req UpdateGroupRequest) (*domain.Group, error) {
		var desc string
		if req.Description != nil {
			desc = *req.Description
		}
		return h.uc.UpdateGroup(ctx, groupID, req.Name, desc, req.RequireAdminApproval, currUser.ID)
	})
}


// ListMembers lists group members with optional status filter (PENDING, REJECTED restricted to admins).
// @Summary      List group members
// @Description  Retrieve group members. Query status=PENDING or REJECTED requires admin privileges.
// @Tags         groups
// @Produce      json
// @Param        id path string true "Group ID"
// @Param        status query string false "Filter status: ACTIVE, PENDING, REJECTED, ALL"
// @Success      200  {array}   domain.Member
// @Failure      401  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/{id}/members [get]
// @Security     BearerAuth
func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	groupID, ok := request.URLParam(w, r, "id")
	if !ok {
		return
	}
	statusFilter := r.URL.Query().Get("status")
	currUser := user.MustFrom(r.Context())

	members, err := h.uc.ListMembers(r.Context(), groupID, statusFilter, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, members)
}

// AddMembers adds multiple users to the group.
// @Summary      Add group members
// @Description  Add multiple users to a group by their User IDs. Only admins can add members.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id path string true "Group ID"
// @Param        request body AddMembersRequest true "User IDs of members to add"
// @Success      201  {array}   domain.Member
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/{id}/members [post]
// @Security     BearerAuth
func (h *Handler) AddMembers(w http.ResponseWriter, r *http.Request) {
	groupID, ok := request.URLParam(w, r, "id")
	if !ok {
		return
	}

	currUser := user.MustFrom(r.Context())

	request.Run(w, r, http.StatusCreated, func(ctx context.Context, req AddMembersRequest) ([]domain.Member, error) {
		return h.uc.AddMembers(ctx, groupID, req.UserIDs, currUser.ID)
	})
}

// RemoveMember removes a user from the group (or leaves the group).
// @Summary      Remove group member / Leave group
// @Description  Remove a user from a group, or leave the group if removing yourself.
// @Tags         groups
// @Param        id path string true "Group ID"
// @Param        userId path string true "User ID of the member to remove"
// @Success      204  "No Content"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/{id}/members/{userId} [delete]
// @Security     BearerAuth
func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	groupID, ok := request.URLParam(w, r, "id")
	if !ok {
		return
	}
	targetUserID, ok := request.URLParam(w, r, "userId")
	if !ok {
		return
	}

	currUser := user.MustFrom(r.Context())

	err := h.uc.RemoveMember(r.Context(), groupID, targetUserID, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateMemberRole updates a member's role (admin vs member).
// @Summary      Update member role
// @Description  Update the role (e.g. ADMIN, MEMBER) of a user inside the group.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id path string true "Group ID"
// @Param        userId path string true "User ID of the member to update"
// @Param        request body UpdateRoleRequest true "New role value"
// @Success      200  {object}  domain.Member
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/{id}/members/{userId}/role [patch]
// @Security     BearerAuth
func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	groupID, ok := request.URLParam(w, r, "id")
	if !ok {
		return
	}
	targetUserID, ok := request.URLParam(w, r, "userId")
	if !ok {
		return
	}

	currUser := user.MustFrom(r.Context())

	request.Run(w, r, http.StatusOK, func(ctx context.Context, req UpdateRoleRequest) (*domain.Member, error) {
		return h.uc.UpdateMemberRole(ctx, groupID, targetUserID, req.Role, currUser.ID)
	})
}

// DecideJoinRequest approves or rejects a pending join request. Admin only.
// @Summary      Decide join request
// @Description  Approve or reject a pending member join request. Admin privileges required.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id path string true "Group ID"
// @Param        userId path string true "Target User ID"
// @Param        request body domain.DecideJoinRequestPayload true "Action (APPROVE or REJECT)"
// @Success      200  {object}  domain.Member
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/{id}/members/{userId}/decision [post]
// @Security     BearerAuth
func (h *Handler) DecideJoinRequest(w http.ResponseWriter, r *http.Request) {
	groupID, ok := request.URLParam(w, r, "id")
	if !ok {
		return
	}
	targetUserID, ok := request.URLParam(w, r, "userId")
	if !ok {
		return
	}
	currUser := user.MustFrom(r.Context())

	request.Run(w, r, http.StatusOK, func(ctx context.Context, req domain.DecideJoinRequestPayload) (*domain.Member, error) {
		return h.uc.DecideJoinRequest(ctx, groupID, targetUserID, req.Action, currUser.ID)
	})
}

// ResetInviteCode generates a new 7-day invite code for the group. Admin only.
// @Summary      Reset group invite code
// @Description  Generate a new invite code with a 7-day expiration timestamp. Admin privileges required.
// @Tags         groups
// @Produce      json
// @Param        id path string true "Group ID"
// @Success      200  {object}  domain.Group
// @Failure      401  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/{id}/invite-code/reset [post]
// @Security     BearerAuth
func (h *Handler) ResetInviteCode(w http.ResponseWriter, r *http.Request) {
	groupID, ok := request.URLParam(w, r, "id")
	if !ok {
		return
	}
	currUser := user.MustFrom(r.Context())

	g, err := h.uc.ResetInviteCode(r.Context(), groupID, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, g)
}

// Archive archives (soft-deletes) the group.
// @Summary      Archive group
// @Description  Soft-delete a bill splitting group. Only group creators can archive.
// @Tags         groups
// @Produce      json
// @Param        id path string true "Group ID"
// @Success      204  "No Content"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/{id} [delete]
// @Security     BearerAuth
func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	groupID, ok := request.URLParam(w, r, "id")
	if !ok {
		return
	}

	currUser := user.MustFrom(r.Context())

	err := h.uc.ArchiveGroup(r.Context(), groupID, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetFeed retrieves group activities with cursor-based pagination.
// @Summary      Get group activity feed
// @Description  Get a cursor-paginated timeline of events inside a group.
// @Tags         groups
// @Produce      json
// @Param        id       path      string  true   "Group ID"
// @Param        limit    query     int     false  "Items per page (max 100, default 20)"
// @Param        cursor   query     string  false  "Opaque cursor token from previous response"
// @Success      200      {object}  pagination.Response[activity.Activity]
// @Failure      401      {object}  response.ErrorResponse
// @Failure      403      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Router       /groups/{id}/feed [get]
// @Security     BearerAuth
func (h *Handler) GetFeed(w http.ResponseWriter, r *http.Request) {
	groupID, ok := request.URLParam(w, r, "id")
	if !ok {
		return
	}

	currUser := user.MustFrom(r.Context())

	p := pagination.ParseParams(r, 20, 100)

	feed, err := h.uc.GetGroupFeed(r.Context(), groupID, currUser.ID, p)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, feed)
}

// UploadIcon handles multipart image upload for a group icon.
// @Summary      Upload group icon
// @Description  Upload an image file (JPEG, PNG, WEBP, max 2MB) as group icon.
// @Tags         groups
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path string true "Group ID"
// @Param        file formData file true "Image File"
// @Success      200 {object} domain.Group
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /groups/{id}/icon [post]
// @Security     BearerAuth
func (h *Handler) UploadIcon(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())
	groupID := chi.URLParam(r, "id")

	r.Body = http.MaxBytesReader(w, r.Body, 2*1024*1024)
	if err := r.ParseMultipartForm(2 * 1024 * 1024); err != nil {
		response.BadRequest(w, "File size exceeds 2MB limit")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "Missing file in request")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		response.BadRequest(w, "Invalid file type. Only JPEG, PNG, WEBP allowed")
		return
	}

	updatedGroup, err := h.uc.UploadGroupIcon(r.Context(), groupID, currUser.ID, file, header.Filename, contentType)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, updatedGroup)
}

