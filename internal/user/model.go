package user

import (
	"time"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
)

// User represents a user in the system.
type User struct {
	ID              string    `json:"id"`
	FirebaseUID     string    `json:"-"`
	Email           *string   `json:"email,omitempty"`
	Phone           *string   `json:"phone,omitempty"`
	Name            string    `json:"name"`
	DefaultCurrency string    `json:"defaultCurrency"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type registerRequest struct {
	Name string `json:"name"`
}

type updateProfileRequest struct {
	Name            string `json:"name"`
	DefaultCurrency string `json:"defaultCurrency"`
}

type addFriendRequest struct {
	FriendEmail string `json:"friendEmail"`
	FriendPhone string `json:"friendPhone"`
}

// ListFriendsResponse represents the paginated friends list response.
type ListFriendsResponse struct {
	Data       []User          `json:"data"`
	Pagination pagination.Meta `json:"pagination"`
}
