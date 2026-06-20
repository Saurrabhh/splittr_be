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
