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
// @Summary      List activity feed
// @Description  Get audit logs of all actions performed by the current user or in their groups.
// @Tags         activities
// @Produce      json
// @Success      200  {array}   Activity
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /activities [get]
// @Security     Bearer
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	activities, err := h.uc.ListActivities(r.Context(), currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, activities)
}
