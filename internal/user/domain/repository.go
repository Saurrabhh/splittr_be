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
	SyncFriendsBySequence(ctx context.Context, lastVersion int64, userID string, limit int32) ([]FriendshipSyncRecord, error)
}

type FriendshipSyncRecord struct {
	UserID      string    `json:"userId"`
	FriendID    string    `json:"friendId"`
	CreatedAt   time.Time `json:"createdAt"`
	SyncVersion int64     `json:"syncVersion"`
}

