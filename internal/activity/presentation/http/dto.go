package http

import (
	"github.com/Saurrabhh/splittr_be/internal/activity/domain"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
)

// ListActivitiesResponse represents the paginated user activities response.
type ListActivitiesResponse struct {
	Data       []domain.Activity `json:"data"`
	Pagination pagination.Meta   `json:"pagination"`
}

// FeedResponse represents the paginated group activity feed response.
type FeedResponse struct {
	Data       []domain.Activity `json:"data"`
	Pagination pagination.Meta   `json:"pagination"`
}
