package domain

import (
	"context"
	"time"
)

// Repository defines the storage contract for users.
type Repository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (*User, error)
	Create(ctx context.Context, u *User) error
	UpdateUser(ctx context.Context, u *User) error
	GetByEmailOrPhone(ctx context.Context, email, phone string) (*User, error)
	CreateFriendship(ctx context.Context, userID, friendID string) error
	DeleteFriendship(ctx context.Context, userID, friendID string) error
	GetFriendship(ctx context.Context, userID, friendID string) (bool, error)
	ListFriends(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]User, error)
}
