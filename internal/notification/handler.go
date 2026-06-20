package notification

import (
	"net/http"

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
// @Description  Get all notifications in the tray for the current user.
// @Tags         notifications
// @Produce      json
// @Success      200  {array}   Notification
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /notifications [get]
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

	notifs, err := h.uc.ListNotifications(r.Context(), currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, notifs)
}

// MarkAsRead marks a specific notification as read.
// @Summary      Mark notification as read
// @Description  Mark a specific notification as read by ID.
// @Tags         notifications
// @Produce      json
// @Param        id path string true "Notification ID"
// @Success      200  {object}  map[string]string "Success message"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /notifications/{id}/read [post]
// @Security     Bearer
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

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

	response.JSON(w, http.StatusOK, map[string]string{"message": "notification marked as read"})
}

// MarkAllAsRead marks all notifications as read.
// @Summary      Mark all notifications as read
// @Description  Mark all unread notifications as read for the current user.
// @Tags         notifications
// @Produce      json
// @Success      200  {object}  map[string]string "Success message"
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /notifications/read-all [post]
// @Security     Bearer
func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	err := h.uc.MarkAllAsRead(r.Context(), currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "all notifications marked as read"})
}
