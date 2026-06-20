package group

import (
	"encoding/json"
	"net/http"

	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for group endpoints.
type Handler struct {
	uc *Usecase
}

// NewHandler creates a new Handler instance.
func NewHandler(uc *Usecase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes registers the group endpoints on the router.
// Note: It assumes that the router already has authentication and user context middlewares applied.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/groups", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Post("/join", h.Join)
		r.Get("/", h.List)
		
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetDetails)
			r.Delete("/", h.Archive)
			
			r.Route("/members", func(r chi.Router) {
				r.Post("/", h.AddMember)
				r.Route("/{userId}", func(r chi.Router) {
					r.Delete("/", h.RemoveMember)
					r.Put("/role", h.UpdateMemberRole)
				})
			})
		})
	})
}

type createGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type joinGroupRequest struct {
	InviteCode string `json:"inviteCode"`
}

// Create creates a new group.
// @Summary      Create group
// @Description  Create a new bill-splitting group.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        request body createGroupRequest true "Group creation data"
// @Success      201  {object}  Group
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups [post]
// @Security     BearerAuth
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.ErrInvalidBody, "invalid request body")
		return
	}

	g, err := h.uc.CreateGroup(r.Context(), req.Name, req.Description, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, g)
}

// Join joins a group via its invite code.
// @Summary      Join group
// @Description  Join an existing group using its unique invite code.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        request body joinGroupRequest true "Group join data"
// @Success      200  {object}  Group
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/join [post]
// @Security     BearerAuth
func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	var req joinGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.ErrInvalidBody, "invalid request body")
		return
	}

	if req.InviteCode == "" {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: "inviteCode is required",
		})
		return
	}

	g, err := h.uc.JoinGroup(r.Context(), req.InviteCode, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, g)
}

// List retrieves all groups the user is a member of.
// @Summary      List groups
// @Description  Retrieve all bill-splitting groups the current user belongs to.
// @Tags         groups
// @Produce      json
// @Success      200  {array}   Group
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups [get]
// @Security     BearerAuth
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	groups, err := h.uc.ListUserGroups(r.Context(), currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, groups)
}

type groupDetailsResponse struct {
	Group
	Members []GroupMember `json:"members"`
}

// GetDetails returns the group metadata and its members list.
// @Summary      Get group details
// @Description  Get a group's metadata and a list of all its members.
// @Tags         groups
// @Produce      json
// @Param        id path string true "Group ID"
// @Success      200  {object}  groupDetailsResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/{id} [get]
// @Security     BearerAuth
func (h *Handler) GetDetails(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")
	if groupID == "" {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: "group id is required",
		})
		return
	}

	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	g, members, err := h.uc.GetGroupDetails(r.Context(), groupID, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, groupDetailsResponse{
		Group:   *g,
		Members: members,
	})
}

type addMemberRequest struct {
	UserID string `json:"userId"`
}

// AddMember adds a user to the group.
// @Summary      Add group member
// @Description  Add a user to a group by their User ID. Only admins can add members.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id path string true "Group ID"
// @Param        request body addMemberRequest true "User ID of the member to add"
// @Success      200  {object}  map[string]string "Success message"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/{id}/members [post]
// @Security     BearerAuth
func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")
	if groupID == "" {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: "group id is required",
		})
		return
	}

	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.ErrInvalidBody, "invalid request body")
		return
	}

	if req.UserID == "" {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: "userId is required",
		})
		return
	}

	err := h.uc.AddMember(r.Context(), groupID, req.UserID, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "member added successfully"})
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
	groupID := chi.URLParam(r, "id")
	targetUserID := chi.URLParam(r, "userId")
	if groupID == "" || targetUserID == "" {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: "group id and user id are required",
		})
		return
	}

	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	err := h.uc.RemoveMember(r.Context(), groupID, targetUserID, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type updateRoleRequest struct {
	Role string `json:"role"`
}

// UpdateMemberRole updates a member's role (admin vs member).
// @Summary      Update member role
// @Description  Update the role (e.g. ADMIN, MEMBER) of a user inside the group.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id path string true "Group ID"
// @Param        userId path string true "User ID of the member to update"
// @Param        request body updateRoleRequest true "New role value"
// @Success      200  {object}  map[string]string "Success message"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/{id}/members/{userId}/role [put]
// @Security     BearerAuth
func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")
	targetUserID := chi.URLParam(r, "userId")
	if groupID == "" || targetUserID == "" {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: "group id and user id are required",
		})
		return
	}

	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.ErrInvalidBody, "invalid request body")
		return
	}

	err := h.uc.UpdateMemberRole(r.Context(), groupID, targetUserID, req.Role, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "role updated successfully"})
}

// Archive archives (soft-deletes) the group.
// @Summary      Archive group
// @Description  Soft-delete a bill splitting group. Only group creators can archive.
// @Tags         groups
// @Produce      json
// @Param        id path string true "Group ID"
// @Success      200  {object}  map[string]string "Success message"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /groups/{id} [delete]
// @Security     BearerAuth
func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")
	if groupID == "" {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: "group id is required",
		})
		return
	}

	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	err := h.uc.ArchiveGroup(r.Context(), groupID, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "group archived successfully"})
}
