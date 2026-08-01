package http

import (
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/user/domain"
)

type RegisterRequest struct {
	Name string `json:"name"`
}

type UpdateProfileRequest struct {
	Name            string `json:"name"`
	DefaultCurrency string `json:"defaultCurrency"`
}

type AddFriendRequest struct {
	FriendEmail string `json:"friendEmail"`
	FriendPhone string `json:"friendPhone"`
}

// ListFriendsResponse represents the paginated friends list response.
type ListFriendsResponse struct {
	Data       []domain.User   `json:"data"`
	Pagination pagination.Meta `json:"pagination"`
}
