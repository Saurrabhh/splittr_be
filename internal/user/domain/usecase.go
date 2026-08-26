package domain

import (
	"context"
	"io"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/storage"
	"github.com/google/uuid"
)

// UseCase handles business operations for users, settings, and friendships.
type UseCase struct {
	repo    Repository
	storage storage.Service
}

// NewUseCase creates a new UseCase instance.
func NewUseCase(repo Repository, storageSvc storage.Service) *UseCase {
	return &UseCase{
		repo:    repo,
		storage: storageSvc,
	}
}

// RegisterUser registers a new user in the system if they do not exist, and creates default user settings.
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

	// Create default user_settings row for newly registered user
	if err := u.repo.CreateDefaultSettings(ctx, newUser.ID); err != nil {
		// Non-fatal fallback
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

// AddFriendByEmailOrPhone matches a user profile by email or phone and establishes a friendship or pending request.
func (u *UseCase) AddFriendByEmailOrPhone(ctx context.Context, userID string, email string, phone string) (*FriendWithStatus, error) {
	if email == "" && phone == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgMissingEmailOrPhone,
		}
	}

	friendWithSettings, err := u.repo.GetByEmailOrPhoneWithSettings(ctx, email, phone)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogLookupUser,
			Err:     err,
		}
	}
	if friendWithSettings == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgUserNotFound,
		}
	}

	friend := &friendWithSettings.User

	if friend.ID == userID {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgSelfFriendError,
		}
	}

	existingFriendship, err := u.repo.GetFriendship(ctx, userID, friend.ID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogVerifyFriendship,
			Err:     err,
		}
	}
	if existingFriendship != nil {
		if existingFriendship.Status == Blocked {
			return nil, &response.AppError{
				Type:    response.TypeValidation,
				Message: "Cannot add friend: user is blocked",
			}
		}
		return &FriendWithStatus{
			User:         *friend,
			Status:       existingFriendship.Status,
			ActionUserID: existingFriendship.ActionUserID,
		}, nil
	}

	status := Pending
	if friendWithSettings.AutoAcceptFriendRequests {
		status = Accepted
	}

	if err := u.repo.CreateFriendship(ctx, userID, friend.ID, status, userID); err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogAddFriend,
			Err:     err,
		}
	}

	return &FriendWithStatus{
		User:         *friend,
		Status:       status,
		ActionUserID: userID,
	}, nil
}

// UpdateFriendshipStatus updates the status of an existing friendship (ACCEPTED, DECLINED, BLOCKED) and returns the updated friend details.
func (u *UseCase) UpdateFriendshipStatus(ctx context.Context, userID string, friendID string, status FriendshipStatus) (*FriendWithStatus, error) {
	if friendID == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	if status != Accepted && status != Declined && status != Blocked {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "Invalid friendship status",
		}
	}

	existing, err := u.repo.GetFriendship(ctx, userID, friendID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogVerifyFriendship,
			Err:     err,
		}
	}
	if existing == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgNotFriends,
		}
	}

	if err := u.repo.UpdateFriendshipStatus(ctx, userID, friendID, status, userID); err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "Failed to update friendship status",
			Err:     err,
		}
	}

	friendUser, err := u.repo.GetByID(ctx, friendID)
	if err != nil || friendUser == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgUserNotFound,
		}
	}

	return &FriendWithStatus{
		User:         *friendUser,
		Status:       status,
		ActionUserID: userID,
	}, nil
}

// RemoveFriend deletes a friendship link.
func (u *UseCase) RemoveFriend(ctx context.Context, userID string, friendID string) error {
	if friendID == "" {
		return &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidParam,
		}
	}

	existing, err := u.repo.GetFriendship(ctx, userID, friendID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogVerifyFriendship,
			Err:     err,
		}
	}
	if existing == nil {
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

// ListFriends returns a cursor-paginated list of the user's accepted friends.
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

// ListFriendsByStatus returns a list of friends filtered by status (ACCEPTED, PENDING, BLOCKED).
func (u *UseCase) ListFriendsByStatus(ctx context.Context, userID string, status FriendshipStatus) ([]FriendWithStatus, error) {
	if userID == "" {
		return nil, &response.AppError{Type: response.TypeValidation, Message: response.MsgInvalidParam}
	}
	if status == "" {
		status = Accepted
	}

	friends, err := u.repo.ListFriendsByStatus(ctx, userID, status)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveFriends,
			Err:     err,
		}
	}
	return friends, nil
}

// GetUserSettings retrieves the settings for a user.
func (u *UseCase) GetUserSettings(ctx context.Context, userID string) (*UserSettings, error) {
	if userID == "" {
		return nil, &response.AppError{Type: response.TypeValidation, Message: response.MsgInvalidParam}
	}
	settings, err := u.repo.GetSettings(ctx, userID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "Failed to retrieve user settings",
			Err:     err,
		}
	}
	return settings, nil
}

// UpdateUserSettings creates or updates the user settings.
func (u *UseCase) UpdateUserSettings(ctx context.Context, userID string, autoAcceptFriendRequests bool) (*UserSettings, error) {
	if userID == "" {
		return nil, &response.AppError{Type: response.TypeValidation, Message: response.MsgInvalidParam}
	}

	settings := &UserSettings{
		UserID:                   userID,
		AutoAcceptFriendRequests: autoAcceptFriendRequests,
	}

	if err := u.repo.UpsertSettings(ctx, settings); err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "Failed to update user settings",
			Err:     err,
		}
	}

	return settings, nil
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

// UploadAvatar streams the user avatar image to storage provider and updates avatar URL in DB.
func (u *UseCase) UploadAvatar(ctx context.Context, userID string, file io.Reader, fileName, contentType string) (*User, error) {
	if u.storage == nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "Storage service not initialized",
		}
	}

	res, err := u.storage.Upload(ctx, storage.UploadParams{
		File:        file,
		FileName:    fileName,
		ContentType: contentType,
		Folder:      "splittr/users/" + userID,
		Width:       500,
		Height:      500,
	})
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "Failed to upload avatar image to storage provider",
			Err:     err,
		}
	}

	updatedUser, err := u.repo.UpdateAvatar(ctx, userID, res.URL)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "Failed to update user avatar URL in database",
			Err:     err,
		}
	}

	return updatedUser, nil
}
