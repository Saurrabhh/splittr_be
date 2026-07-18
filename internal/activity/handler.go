package activity

import (
	"net/http"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
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
	r.Get("/groups/{groupId}/feed", h.GetFeed)
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
// @Security     BearerAuth
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())

	activities, err := h.uc.ListActivities(r.Context(), currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, activities)
}

// GetFeed retrieves group activities with cursor-based pagination.
// @Summary      Get group activity feed
// @Description  Get a cursor-paginated timeline of events inside a group.
// @Tags         activities
// @Produce      json
// @Param        groupId  path      string  true   "Group ID"
// @Param        limit    query     int     false  "Items per page (max 100, default 20)"
// @Param        cursor   query     string  false  "Opaque cursor token from previous response"
// @Success      200      {object}  FeedResponse
// @Failure      401      {object}  response.ErrorResponse
// @Failure      403      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Router       /groups/{groupId}/feed [get]
// @Security     BearerAuth
func (h *Handler) GetFeed(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())
	groupId := chi.URLParam(r, "groupId")

	// Shared helper: parses ?limit and ?cursor from query string
	p := pagination.ParseParams(r, 20, 100)

	feed, err := h.uc.GetGroupFeed(r.Context(), currUser.ID, groupId, p)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, feed)
}
