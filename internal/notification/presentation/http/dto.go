package http

import (
	"github.com/Saurrabhh/splittr_be/internal/notification/domain"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
)

// ListNotificationsResponse represents the paginated user notifications response.
type ListNotificationsResponse struct {
	Data       []domain.Notification `json:"data"`
	Pagination pagination.Meta       `json:"pagination"`
}
