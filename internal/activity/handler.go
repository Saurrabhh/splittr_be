package activity

import (
	"net/http"

	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for user activities.
type Handler struct {
	uc *Usecase
}

// NewHandler creates a new Handler instance.
func NewHandler(uc *Usecase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes registers endpoints on the router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/activities", h.List)
}

// List retrieves the activity feed for the current user.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized: missing user profile")
		return
	}

	activities, err := h.uc.ListActivities(r.Context(), currUser.ID)
	if err != nil {
		response.InternalServerError(w, response.ErrInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, activities)
}
