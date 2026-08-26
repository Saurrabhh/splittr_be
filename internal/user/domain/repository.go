package domain

import (
	"context"
	"time"
)

// Repository defines the storage contract for users, settings, and friendships.
type Repository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (*User, error)
	Create(ctx context.Context, u *User) error
	UpdateUser(ctx context.Context, u *User) error
	UpdateAvatar(ctx context.Context, userID string, avatarURL string) (*User, error)
	GetByEmailOrPhone(ctx context.Context, email, phone string) (*User, error)
	GetByEmailOrPhoneWithSettings(ctx context.Context, email, phone string) (*UserWithSettings, error)

	CreateDefaultSettings(ctx context.Context, userID string) error
	GetSettings(ctx context.Context, userID string) (*UserSettings, error)
	UpsertSettings(ctx context.Context, settings *UserSettings) error

	CreateFriendship(ctx context.Context, userID, friendID string, status FriendshipStatus, actionUserID string) error
	UpdateFriendshipStatus(ctx context.Context, userID, friendID string, status FriendshipStatus, actionUserID string) error
	DeleteFriendship(ctx context.Context, userID, friendID string) error
	GetFriendship(ctx context.Context, userID, friendID string) (*Friendship, error)
	ListFriends(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]User, error)
	ListFriendsByStatus(ctx context.Context, userID string, status FriendshipStatus) ([]FriendWithStatus, error)
	SyncFriendsBySequence(ctx context.Context, lastVersion int64, userID string, limit int32) ([]FriendshipSyncRecord, error)
	GetFriendTombstonesBySequence(ctx context.Context, lastVersion int64, userID string, limit int32) ([]Tombstone, error)
}

type FriendshipSyncRecord struct {
	UserID       string           `json:"userId"`
	FriendID     string           `json:"friendId"`
	Status       FriendshipStatus `json:"status"`
	ActionUserID string           `json:"actionUserId"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
	SyncVersion  int64            `json:"syncVersion"`
} // @name User.FriendshipSyncRecord

// Tombstone represents a deleted entity for sync.
type Tombstone struct {
	EntityID    string
	SyncVersion int64
}

