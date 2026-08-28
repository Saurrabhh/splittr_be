package http

import (
	"net/http"
	"strconv"

	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/sync/domain"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for unified sync.
type Handler struct {
	uc *domain.UseCase
}

// NewHandler creates a new Handler instance.
func NewHandler(uc *domain.UseCase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes mounts sync routes on a Chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/sync", h.Sync)
}

// Sync retrieves delta sync for friends, groups, and expenses.
// @Summary      Unified batch sync
// @Description  Retrieve delta updates and tombstones for friends, groups, and expenses after given sequence versions in a single round-trip.
// @Tags         sync
// @Produce      json
// @Param        friendsVersion  query  int64  false  "Last received friends sequence version"
// @Param        groupsVersion   query  int64  false  "Last received groups sequence version"
// @Param        expensesVersion query  int64  false  "Last received expenses sequence version"
// @Param        limit           query  int    false  "Maximum items per entity category (default 100)"
// @Success      200  {object}  domain.SyncResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /sync [get]
// @Security     BearerAuth
func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())

	parseVersion := func(param string) int64 {
		val := r.URL.Query().Get(param)
		if val == "" {
			return 0
		}
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0
		}
		return v
	}

	limit := int32(100)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	params := domain.SyncParams{
		FriendsVersion:  parseVersion("friendsVersion"),
		GroupsVersion:   parseVersion("groupsVersion"),
		ExpensesVersion: parseVersion("expensesVersion"),
		Limit:           limit,
	}

	res, err := h.uc.Sync(r.Context(), currUser.ID, params)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, res)
}
