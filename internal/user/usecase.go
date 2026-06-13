package user

import (
	"context"
	"errors"
	"fmt"
)

// UserFinder defines the consumer-side interface for retrieving users.
type UserFinder interface {
	GetByID(ctx context.Context, id string) (*User, error)
}

// UserCreator defines the consumer-side interface for creating users.
type UserCreator interface {
	Create(ctx context.Context, u *User) error
}

// Usecase handles business operations for users.
type Usecase struct {
	finder  UserFinder
	creator UserCreator
}

// NewUsecase creates a new Usecase instance.
func NewUsecase(finder UserFinder, creator UserCreator) *Usecase {
	return &Usecase{
		finder:  finder,
		creator: creator,
	}
}

// RegisterUser registers a new user in the system if they do not exist.
func (u *Usecase) RegisterUser(ctx context.Context, id string, email, phone *string, name string) (*User, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	if (email == nil || *email == "") && (phone == nil || *phone == "") {
		return nil, errors.New("either email or phone is required")
	}

	existing, err := u.finder.GetByID(ctx, id)
	if err == nil && existing != nil {
		return existing, nil
	}

	newUser := &User{
		ID:    id,
		Email: email,
		Phone: phone,
		Name:  name,
	}

	if err := u.creator.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return newUser, nil
}

// GetUserProfile retrieves the profile of a user by ID.
func (u *Usecase) GetUserProfile(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	return u.finder.GetByID(ctx, id)
}
