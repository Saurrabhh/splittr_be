package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Repository defines the storage contract for users.
type Repository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (*User, error)
	Create(ctx context.Context, u *User) error
}

// Usecase handles business operations for users.
type Usecase struct {
	repo Repository
}

// NewUsecase creates a new Usecase instance.
func NewUsecase(repo Repository) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

// RegisterUser registers a new user in the system if they do not exist.
func (u *Usecase) RegisterUser(ctx context.Context, firebaseUID string, email, phone *string, name string) (*User, error) {
	if firebaseUID == "" {
		return nil, errors.New("firebaseUID is required")
	}
	if (email == nil || *email == "") && (phone == nil || *phone == "") {
		return nil, errors.New("either email or phone is required")
	}

	existing, err := u.repo.GetByFirebaseUID(ctx, firebaseUID)
	if err == nil && existing != nil {
		return existing, nil
	}

	newUser := &User{
		ID:          uuid.New().String(),
		FirebaseUID: firebaseUID,
		Email:       email,
		Phone:       phone,
		Name:        name,
	}

	if err := u.repo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return newUser, nil
}

// GetUserProfile retrieves the profile of a user by local ID.
func (u *Usecase) GetUserProfile(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	return u.repo.GetByID(ctx, id)
}

// GetUserByFirebaseUID retrieves the profile of a user by Firebase UID.
func (u *Usecase) GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*User, error) {
	if firebaseUID == "" {
		return nil, errors.New("firebaseUID is required")
	}
	return u.repo.GetByFirebaseUID(ctx, firebaseUID)
}


