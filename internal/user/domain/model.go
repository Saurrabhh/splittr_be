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
} // @name User
