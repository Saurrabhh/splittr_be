package http

import (
	"context"
	"net/http"

	"github.com/Saurrabhh/splittr_be/internal/notification/domain"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/request"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for notifications.
type Handler struct {
	uc *domain.UseCase
}

// NewHandler creates a new Handler instance.
func NewHandler(uc *domain.UseCase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes registers endpoints on the router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/notifications", func(r chi.Router) {
		r.Get("/", h.List)
		r.Patch("/{id}", h.MarkAsRead)
		r.Patch("/", h.MarkAllAsRead)
	})
}

// List lists all notifications for the current user.
// @Summary      List notifications
// @Description  Get a cursor-paginated list of notifications for the current user.
// @Tags         notifications
// @Produce      json
// @Param        limit   query  int     false  "Items per page (max 100, default 20)"
// @Param        cursor  query  string  false  "Opaque cursor token from a previous response"
// @Success      200  {object}  pagination.Response[domain.Notification]
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
// @Accept       json
// @Produce      json
// @Param        id path string true "Notification ID"
// @Param        request body UpdateNotificationRequest true "Update notification payload"
// @Success      200  {object}  response.MessageResponse "Success message"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /notifications/{id} [patch]
// @Security     BearerAuth
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())

	id, ok := request.URLParam(w, r, "id")
	if !ok {
		return
	}

	request.Run(w, r, http.StatusOK, func(ctx context.Context, req UpdateNotificationRequest) (*response.MessageResponse, error) {
		err := h.uc.MarkAsRead(ctx, id, currUser.ID)
		if err != nil {
			return nil, err
		}
		return &response.MessageResponse{Message: "notification marked as read"}, nil
	})
}

// MarkAllAsRead marks all notifications as read.
// @Summary      Mark all notifications as read
// @Description  Mark all unread notifications as read for the current user.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        request body BulkUpdateNotificationRequest true "Bulk update payload"
// @Success      200  {object}  response.MessageResponse "Success message"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /notifications [patch]
// @Security     BearerAuth
func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())

	request.Run(w, r, http.StatusOK, func(ctx context.Context, req BulkUpdateNotificationRequest) (*response.MessageResponse, error) {
		err := h.uc.MarkAllAsRead(ctx, currUser.ID)
		if err != nil {
			return nil, err
		}
		return &response.MessageResponse{Message: "all notifications marked as read"}, nil
	})
}
