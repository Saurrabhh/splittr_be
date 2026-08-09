package domain

import (
	"context"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/google/uuid"
)

// UseCase handles business operations for users.
type UseCase struct {
	repo Repository
}

// NewUseCase creates a new UseCase instance.
func NewUseCase(repo Repository) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

// RegisterUser registers a new user in the system if they do not exist.
func (u *UseCase) RegisterUser(ctx context.Context, firebaseUID string, email, phone *string, name string) (*User, error) {
	if firebaseUID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}
	if (email == nil || *email == "") && (phone == nil || *phone == "") {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgMissingEmailOrPhone,
		}
	}

	existing, err := u.repo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogCheckUser,
			Err:     err,
		}
	}
	if existing != nil {
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
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRegisterUser,
			Err:     err,
		}
	}

	return newUser, nil
}

// GetUserProfile retrieves the profile of a user by local ID.
func (u *UseCase) GetUserProfile(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}
	usr, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveUserProfile,
			Err:     err,
		}
	}
	if usr == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgUserNotFound,
		}
	}
	return usr, nil
}

// GetUserByFirebaseUID retrieves the profile of a user by Firebase UID.
func (u *UseCase) GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*User, error) {
	if firebaseUID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}
	usr, err := u.repo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveUserProfile,
			Err:     err,
		}
	}
	if usr == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgUserNotFound,
		}
	}
	return usr, nil
}

// UpdateProfile updates the name and default currency of a user.
func (u *UseCase) UpdateProfile(ctx context.Context, userID string, name string, defaultCurrency string) (*User, error) {
	usr, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveUserProfile,
			Err:     err,
		}
	}
	if usr == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgUserNotFound,
		}
	}

	if name != "" {
		usr.Name = name
	}
	if defaultCurrency != "" {
		if len(defaultCurrency) != 3 {
			return nil, &response.AppError{
				Type:    response.TypeValidation,
				Message: response.MsgInvalidCurrency,
			}
		}
		usr.DefaultCurrency = defaultCurrency
	}

	if err := u.repo.UpdateUser(ctx, usr); err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogUpdateUserProfile,
			Err:     err,
		}
	}

	return usr, nil
}

// AddFriendByEmailOrPhone matches a user profile by email or phone and establishes a friendship relation.
func (u *UseCase) AddFriendByEmailOrPhone(ctx context.Context, userID string, email string, phone string) (*User, error) {
	if email == "" && phone == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgMissingEmailOrPhone,
		}
	}

	friend, err := u.repo.GetByEmailOrPhone(ctx, email, phone)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogLookupUser,
			Err:     err,
		}
	}
	if friend == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgUserNotFound,
		}
	}

	if friend.ID == userID {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgSelfFriendError,
		}
	}

	isFriend, err := u.repo.GetFriendship(ctx, userID, friend.ID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogVerifyFriendship,
			Err:     err,
		}
	}
	if isFriend {
		return friend, nil
	}

	if err := u.repo.CreateFriendship(ctx, userID, friend.ID); err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogAddFriend,
			Err:     err,
		}
	}

	return friend, nil
}

// RemoveFriend deletes a friendship link.
func (u *UseCase) RemoveFriend(ctx context.Context, userID string, friendID string) error {
	if friendID == "" {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	isFriend, err := u.repo.GetFriendship(ctx, userID, friendID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogVerifyFriendship,
			Err:     err,
		}
	}
	if !isFriend {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgNotFriends,
		}
	}

	if err := u.repo.DeleteFriendship(ctx, userID, friendID); err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRemoveFriend,
			Err:     err,
		}
	}
	return nil
}

// ListFriends returns a cursor-paginated list of the user's friends.
func (u *UseCase) ListFriends(ctx context.Context, userID string, p pagination.Params) (pagination.Response[User], error) {
	if userID == "" {
		return pagination.Response[User]{}, &response.AppError{Type: response.TypeValidation, Message: response.MsgInvalidParam}
	}
	cursor := pagination.ParseCursor(p.Cursor)
	friends, err := u.repo.ListFriends(ctx, userID, p.Limit+1, cursor.LastTime, cursor.LastID)
	if err != nil {
		return pagination.Response[User]{}, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveFriends,
			Err:     err,
		}
	}
	return pagination.BuildResponse(friends, p.Limit, func(usr User) string {
		return pagination.EncodeCursor(usr.CreatedAt, usr.ID)
	}), nil
}

// FriendSyncResponse contains updated friends for offline sync.
type FriendSyncResponse struct {
	NewVersion int64                  `json:"newVersion"`
	Friends    []FriendshipSyncRecord `json:"friends"`
} // @name User.FriendSyncResponse

// SyncFriends retrieves friendship changes after lastVersion for a user.
func (u *UseCase) SyncFriends(ctx context.Context, lastVersion int64, userID string, limit int32) (*FriendSyncResponse, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	records, err := u.repo.SyncFriendsBySequence(ctx, lastVersion, userID, limit)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "Failed to sync friends",
			Err:     err,
		}
	}

	var maxVersion int64 = lastVersion
	for _, r := range records {
		if r.SyncVersion > maxVersion {
			maxVersion = r.SyncVersion
		}
	}

	if records == nil {
		records = []FriendshipSyncRecord{}
	}

	return &FriendSyncResponse{
		NewVersion: maxVersion,
		Friends:    records,
	}, nil
}

