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
	UpdateUser(ctx context.Context, u *User) error
	GetByEmailOrPhone(ctx context.Context, email, phone string) (*User, error)
	CreateFriendship(ctx context.Context, userID, friendID string) error
	DeleteFriendship(ctx context.Context, userID, friendID string) error
	GetFriendship(ctx context.Context, userID, friendID string) (bool, error)
	ListFriends(ctx context.Context, userID string) ([]User, error)
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

// UpdateProfile updates the name and default currency of a user.
func (u *Usecase) UpdateProfile(ctx context.Context, userID string, name string, defaultCurrency string) (*User, error) {
	usr, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if usr == nil {
		return nil, errors.New("user not found")
	}

	if name != "" {
		usr.Name = name
	}
	if defaultCurrency != "" {
		if len(defaultCurrency) != 3 {
			return nil, errors.New("invalid currency code: must be 3 characters")
		}
		usr.DefaultCurrency = defaultCurrency
	}

	if err := u.repo.UpdateUser(ctx, usr); err != nil {
		return nil, err
	}

	return usr, nil
}

// AddFriendByEmailOrPhone matches a user profile by email or phone and establishes a friendship relation.
func (u *Usecase) AddFriendByEmailOrPhone(ctx context.Context, userID string, email string, phone string) (*User, error) {
	if email == "" && phone == "" {
		return nil, errors.New("email or phone must be provided")
	}

	friend, err := u.repo.GetByEmailOrPhone(ctx, email, phone)
	if err != nil {
		return nil, err
	}
	if friend == nil {
		return nil, errors.New("no user found matching the provided email or phone")
	}

	if friend.ID == userID {
		return nil, errors.New("you cannot add yourself as a friend")
	}

	// Check if already friends
	isFriend, err := u.repo.GetFriendship(ctx, userID, friend.ID)
	if err != nil {
		return nil, err
	}
	if isFriend {
		return friend, nil // already friends, return success
	}

	// Add friendship relation
	if err := u.repo.CreateFriendship(ctx, userID, friend.ID); err != nil {
		return nil, err
	}

	return friend, nil
}

// RemoveFriend deletes a friendship link.
func (u *Usecase) RemoveFriend(ctx context.Context, userID string, friendID string) error {
	if friendID == "" {
		return errors.New("friendID is required")
	}

	isFriend, err := u.repo.GetFriendship(ctx, userID, friendID)
	if err != nil {
		return err
	}
	if !isFriend {
		return errors.New("users are not friends")
	}

	return u.repo.DeleteFriendship(ctx, userID, friendID)
}

// ListFriends retrieves the list of user profiles representing friends.
func (u *Usecase) ListFriends(ctx context.Context, userID string) ([]User, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListFriends(ctx, userID)
}


