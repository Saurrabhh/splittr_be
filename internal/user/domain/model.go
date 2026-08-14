package domain

import "time"

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
} // @name User.User

// UserWithSettings represents a user along with their preferences.
type UserWithSettings struct {
	User
	AutoAcceptFriendRequests bool `json:"autoAcceptFriendRequests"`
} // @name User.UserWithSettings

// UserSettings represents individual user preferences.
type UserSettings struct {
	UserID                   string    `json:"userId"`
	AutoAcceptFriendRequests bool      `json:"autoAcceptFriendRequests"`
	CreatedAt                time.Time `json:"createdAt"`
	UpdatedAt                time.Time `json:"updatedAt"`
} // @name User.UserSettings

// Friendship represents a relation between two users with status.
type Friendship struct {
	UserID       string    `json:"userId"`
	FriendID     string    `json:"friendId"`
	Status       string    `json:"status"` // PENDING, ACCEPTED, DECLINED, BLOCKED
	ActionUserID string    `json:"actionUserId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
} // @name User.Friendship

// FriendWithStatus represents a friend along with current friendship status.
type FriendWithStatus struct {
	User
	Status       string `json:"status"`
	ActionUserID string `json:"actionUserId"`
} // @name User.FriendWithStatus
