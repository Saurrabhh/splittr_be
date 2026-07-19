package notification

import (
	"net/http"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for notifications.
type Handler struct {
	uc *Usecase
}

// NewHandler creates a new Handler instance.
func NewHandler(uc *Usecase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes registers endpoints on the router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/notifications", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/{id}/read", h.MarkAsRead)
		r.Post("/read-all", h.MarkAllAsRead)
	})
}

// List lists all notifications for the current user.
// @Summary      List notifications
// @Description  Get a cursor-paginated list of notifications for the current user.
// @Tags         notifications
// @Produce      json
// @Param        limit   query  int     false  "Items per page (max 100, default 20)"
// @Param        cursor  query  string  false  "Opaque cursor token from a previous response"
// @Success      200  {object}  ListNotificationsResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /notifications [get]
// @Security     BearerAuth
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())
	p := pagination.ParseParams(r, 20, 100)
	result, err := h.uc.ListNotifications(r.Context(), currUser.ID, p)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

// MarkAsRead marks a specific notification as read.
// @Summary      Mark notification as read
// @Description  Mark a specific notification as read by ID.
// @Tags         notifications
// @Produce      json
// @Param        id path string true "Notification ID"
// @Success      200  {object}  response.MessageResponse "Success message"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /notifications/{id}/read [post]
// @Security     BearerAuth
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())

	id := chi.URLParam(r, "id")
	if id == "" {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: "notification id is required",
		})
		return
	}

	err := h.uc.MarkAsRead(r.Context(), id, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "notification marked as read"})
}

// MarkAllAsRead marks all notifications as read.
// @Summary      Mark all notifications as read
// @Description  Mark all unread notifications as read for the current user.
// @Tags         notifications
// @Produce      json
// @Success      200  {object}  response.MessageResponse "Success message"
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /notifications/read-all [post]
// @Security     BearerAuth
func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())

	err := h.uc.MarkAllAsRead(r.Context(), currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "all notifications marked as read"})
}
