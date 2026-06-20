package user

import (
	"encoding/json"
	"net/http"

	"github.com/Saurrabhh/splittr_be/internal/auth"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for user domain.
type Handler struct {
	uc *Usecase
}

// NewHandler creates a new Handler instance.
func NewHandler(uc *Usecase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes registers endpoints on the router.
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/users", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/", h.Register)
		r.Group(func(r chi.Router) {
			r.Use(h.UserContext)
			r.Get("/me", h.GetMe)
			r.Put("/me", h.UpdateMe)
		})
	})

	r.Route("/friends", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(h.UserContext)
		r.Post("/", h.AddFriend)
		r.Get("/", h.GetFriends)
		r.Delete("/{friendId}", h.RemoveFriend)
	})
}

type registerRequest struct {
	Name string `json:"name"`
}

type updateProfileRequest struct {
	Name            string `json:"name"`
	DefaultCurrency string `json:"defaultCurrency"`
}

type addFriendRequest struct {
	FriendEmail string `json:"friendEmail"`
	FriendPhone string `json:"friendPhone"`
}

// Register registers the authenticated user.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	identity := auth.IdentityFrom(r.Context())
	if identity == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized")
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrInvalidBody, "invalid request body")
		return
	}

	if req.Name == "" {
		response.BadRequest(w, response.ErrNameRequired, "name is required")
		return
	}

	var emailPtr *string
	if identity.Email != "" {
		emailPtr = &identity.Email
	}

	var phonePtr *string
	if identity.Phone != "" {
		phonePtr = &identity.Phone
	}

	u, err := h.uc.RegisterUser(r.Context(), identity.UserID, emailPtr, phonePtr, req.Name)
	if err != nil {
		response.InternalServerError(w, response.ErrInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, u)
}

// GetMe retrieves the profile of the currently authenticated user.
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized")
		return
	}
	response.JSON(w, http.StatusOK, u)
}

// UpdateMe updates the profile metadata of the current user.
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	currUser := UserFrom(r.Context())
	if currUser == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized")
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrInvalidBody, "invalid request body")
		return
	}

	u, err := h.uc.UpdateProfile(r.Context(), currUser.ID, req.Name, req.DefaultCurrency)
	if err != nil {
		response.BadRequest(w, response.ErrBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, u)
}

// AddFriend adds a user as a friend by email or phone.
func (h *Handler) AddFriend(w http.ResponseWriter, r *http.Request) {
	currUser := UserFrom(r.Context())
	if currUser == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized: missing user profile")
		return
	}

	var req addFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrInvalidBody, "invalid request body")
		return
	}

	friend, err := h.uc.AddFriendByEmailOrPhone(r.Context(), currUser.ID, req.FriendEmail, req.FriendPhone)
	if err != nil {
		response.BadRequest(w, response.ErrBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, friend)
}

// GetFriends returns all friends of the current user.
func (h *Handler) GetFriends(w http.ResponseWriter, r *http.Request) {
	currUser := UserFrom(r.Context())
	if currUser == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized: missing user profile")
		return
	}

	friends, err := h.uc.ListFriends(r.Context(), currUser.ID)
	if err != nil {
		response.InternalServerError(w, response.ErrInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, friends)
}

// RemoveFriend deletes a friendship link.
func (h *Handler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	currUser := UserFrom(r.Context())
	if currUser == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized: missing user profile")
		return
	}

	friendID := chi.URLParam(r, "friendId")
	if friendID == "" {
		response.BadRequest(w, response.ErrBadRequest, "friendId is required")
		return
	}

	err := h.uc.RemoveFriend(r.Context(), currUser.ID, friendID)
	if err != nil {
		response.BadRequest(w, response.ErrBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
