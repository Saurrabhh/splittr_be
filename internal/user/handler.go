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
		r.Get("/me", h.GetMe)
	})
}

type registerRequest struct {
	Name string `json:"name"`
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
	identity := auth.IdentityFrom(r.Context())
	if identity == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized")
		return
	}

	u, err := h.uc.GetUserByFirebaseUID(r.Context(), identity.UserID)
	if err != nil {
		response.InternalServerError(w, response.ErrInternalServerError, err.Error())
		return
	}
	if u == nil {
		response.NotFound(w, response.ErrUserNotFound, "user not found")
		return
	}

	response.JSON(w, http.StatusOK, u)
}
