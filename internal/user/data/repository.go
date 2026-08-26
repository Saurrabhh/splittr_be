package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/Saurrabhh/splittr_be/internal/user/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const DefaultCurrency = "INR"

// DBRepository handles database operations for users.
type DBRepository struct {
	db *db.DB
	tm *db.TransactionManager
}

// NewRepository creates a new DBRepository instance.
func NewRepository(database *db.DB, tm *db.TransactionManager) *DBRepository {
	return &DBRepository{
		db: database,
		tm: tm,
	}
}

// GetByID retrieves a user by ID.
func (r *DBRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbUser, err := q.GetUserByID(ctx, parsedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	return mapUserFields(dbUser.ID, dbUser.FirebaseUid, dbUser.Email, dbUser.Phone, dbUser.Name, dbUser.DefaultCurrency, dbUser.AvatarUrl, dbUser.CreatedAt, dbUser.UpdatedAt), nil
}

// GetByFirebaseUID retrieves a user by Firebase UID.
func (r *DBRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*domain.User, error) {
	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbUser, err := q.GetUserByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by firebase uid: %w", err)
	}

	return mapUserFields(dbUser.ID, dbUser.FirebaseUid, dbUser.Email, dbUser.Phone, dbUser.Name, dbUser.DefaultCurrency, dbUser.AvatarUrl, dbUser.CreatedAt, dbUser.UpdatedAt), nil
}

// Create inserts a new user.
func (r *DBRepository) Create(ctx context.Context, u *domain.User) error {
	parsedID, err := uuid.Parse(u.ID)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	if u.DefaultCurrency == "" {
		u.DefaultCurrency = DefaultCurrency
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbUser, err := q.CreateUser(ctx, dbgen.CreateUserParams{
		ID:              parsedID,
		FirebaseUid:     u.FirebaseUID,
		Email:           ptrToText(u.Email),
		Phone:           ptrToText(u.Phone),
		Name:            u.Name,
		DefaultCurrency: u.DefaultCurrency,
	})
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	u.CreatedAt = dbUser.CreatedAt.Time
	u.UpdatedAt = dbUser.UpdatedAt.Time

	return nil
}

// UpdateUser updates an existing user profile (name and default currency).
func (r *DBRepository) UpdateUser(ctx context.Context, u *domain.User) error {
	parsedID, err := uuid.Parse(u.ID)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbUser, err := q.UpdateUser(ctx, dbgen.UpdateUserParams{
		ID:              parsedID,
		Name:            u.Name,
		DefaultCurrency: u.DefaultCurrency,
	})
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	u.UpdatedAt = dbUser.UpdatedAt.Time
	return nil
}

// UpdateAvatar updates a user's avatar URL.
func (r *DBRepository) UpdateAvatar(ctx context.Context, userID string, avatarURL string) (*domain.User, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbUser, err := q.UpdateUserAvatar(ctx, dbgen.UpdateUserAvatarParams{
		ID:        parsedID,
		AvatarUrl: pgtype.Text{String: avatarURL, Valid: avatarURL != ""},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("update user avatar: %w", err)
	}

	return mapUserFields(dbUser.ID, dbUser.FirebaseUid, dbUser.Email, dbUser.Phone, dbUser.Name, dbUser.DefaultCurrency, dbUser.AvatarUrl, dbUser.CreatedAt, dbUser.UpdatedAt), nil
}

// GetByEmailOrPhone retrieves a user by their email or phone number.
func (r *DBRepository) GetByEmailOrPhone(ctx context.Context, email, phone string) (*domain.User, error) {
	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbUser, err := q.GetUserByEmailOrPhone(ctx, dbgen.GetUserByEmailOrPhoneParams{
		Email: ptrToText(&email),
		Phone: ptrToText(&phone),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by email or phone: %w", err)
	}

	return mapUserFields(dbUser.ID, dbUser.FirebaseUid, dbUser.Email, dbUser.Phone, dbUser.Name, dbUser.DefaultCurrency, dbUser.AvatarUrl, dbUser.CreatedAt, dbUser.UpdatedAt), nil
}

// GetByEmailOrPhoneWithSettings retrieves a user along with their settings.
func (r *DBRepository) GetByEmailOrPhoneWithSettings(ctx context.Context, email, phone string) (*domain.UserWithSettings, error) {
	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	row, err := q.GetUserWithSettingsByEmailOrPhone(ctx, dbgen.GetUserWithSettingsByEmailOrPhoneParams{
		Email: ptrToText(&email),
		Phone: ptrToText(&phone),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query user with settings: %w", err)
	}

	return &domain.UserWithSettings{
		User: domain.User{
			ID:              row.ID.String(),
			FirebaseUID:     row.FirebaseUid,
			Email:           textToPtr(row.Email),
			Phone:           textToPtr(row.Phone),
			Name:            row.Name,
			DefaultCurrency: row.DefaultCurrency,
			CreatedAt:       row.CreatedAt.Time,
			UpdatedAt:       row.UpdatedAt.Time,
		},
		AutoAcceptFriendRequests: row.AutoAcceptFriendRequests,
	}, nil
}

// CreateDefaultSettings creates a default user_settings row for a newly registered user.
func (r *DBRepository) CreateDefaultSettings(ctx context.Context, userID string) error {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	return q.CreateDefaultUserSettings(ctx, parsedID)
}

// GetSettings retrieves settings for a user.
func (r *DBRepository) GetSettings(ctx context.Context, userID string) (*domain.UserSettings, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	s, err := q.GetUserSettings(ctx, parsedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.UserSettings{
				UserID:                   userID,
				AutoAcceptFriendRequests: false,
			}, nil
		}
		return nil, fmt.Errorf("query user settings: %w", err)
	}

	return &domain.UserSettings{
		UserID:                   s.UserID.String(),
		AutoAcceptFriendRequests: s.AutoAcceptFriendRequests,
		CreatedAt:                s.CreatedAt.Time,
		UpdatedAt:                s.UpdatedAt.Time,
	}, nil
}

// UpsertSettings creates or updates user settings.
func (r *DBRepository) UpsertSettings(ctx context.Context, settings *domain.UserSettings) error {
	parsedID, err := uuid.Parse(settings.UserID)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	row, err := q.UpsertUserSettings(ctx, dbgen.UpsertUserSettingsParams{
		UserID:                   parsedID,
		AutoAcceptFriendRequests: settings.AutoAcceptFriendRequests,
	})
	if err != nil {
		return fmt.Errorf("upsert user settings: %w", err)
	}

	settings.CreatedAt = row.CreatedAt.Time
	settings.UpdatedAt = row.UpdatedAt.Time
	return nil
}

// CreateFriendship creates or updates a friendship link with status.
func (r *DBRepository) CreateFriendship(ctx context.Context, userID, friendID string, status domain.FriendshipStatus, actionUserID string) error {
	parsedUser, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user uuid: %w", err)
	}
	parsedFriend, err := uuid.Parse(friendID)
	if err != nil {
		return fmt.Errorf("invalid friend uuid: %w", err)
	}
	parsedActionUser, err := uuid.Parse(actionUserID)
	if err != nil {
		return fmt.Errorf("invalid action user uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	return q.CreateFriendship(ctx, dbgen.CreateFriendshipParams{
		UserID:       parsedUser,
		FriendID:     parsedFriend,
		Status:       string(status),
		ActionUserID: pgtype.UUID{Bytes: parsedActionUser, Valid: true},
	})
}

// UpdateFriendshipStatus updates the status of an existing friendship.
func (r *DBRepository) UpdateFriendshipStatus(ctx context.Context, userID, friendID string, status domain.FriendshipStatus, actionUserID string) error {
	parsedUser, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user uuid: %w", err)
	}
	parsedFriend, err := uuid.Parse(friendID)
	if err != nil {
		return fmt.Errorf("invalid friend uuid: %w", err)
	}
	parsedActionUser, err := uuid.Parse(actionUserID)
	if err != nil {
		return fmt.Errorf("invalid action user uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	return q.UpdateFriendshipStatus(ctx, dbgen.UpdateFriendshipStatusParams{
		UserID:       parsedUser,
		FriendID:     parsedFriend,
		Status:       string(status),
		ActionUserID: pgtype.UUID{Bytes: parsedActionUser, Valid: true},
	})
}

// DeleteFriendship deletes a friendship link.
func (r *DBRepository) DeleteFriendship(ctx context.Context, userID, friendID string) error {
	parsedUser, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user uuid: %w", err)
	}
	parsedFriend, err := uuid.Parse(friendID)
	if err != nil {
		return fmt.Errorf("invalid friend uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	return q.DeleteFriendship(ctx, dbgen.DeleteFriendshipParams{
		UserID:   parsedUser,
		FriendID: parsedFriend,
	})
}

// GetFriendship retrieves a friendship relation between two users.
func (r *DBRepository) GetFriendship(ctx context.Context, userID, friendID string) (*domain.Friendship, error) {
	parsedUser, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user uuid: %w", err)
	}
	parsedFriend, err := uuid.Parse(friendID)
	if err != nil {
		return nil, fmt.Errorf("invalid friend uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	row, err := q.GetFriendship(ctx, dbgen.GetFriendshipParams{
		UserID:   parsedUser,
		FriendID: parsedFriend,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query friendship: %w", err)
	}

	actionUserID := ""
	if row.ActionUserID.Valid {
		actionUserID = uuid.UUID(row.ActionUserID.Bytes).String()
	}

	return &domain.Friendship{
		UserID:       row.UserID.String(),
		FriendID:     row.FriendID.String(),
		Status:       domain.FriendshipStatus(row.Status),
		ActionUserID: actionUserID,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

// ListFriends retrieves a user's friends with status = ACCEPTED.
func (r *DBRepository) ListFriends(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.User, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	pgLastTime, lastIDUUID, err := db.ParseCursor(lastTime, lastID)
	if err != nil {
		return nil, err
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListFriendsPaginated(ctx, dbgen.ListFriendsPaginatedParams{
		UserID:  parsedID,
		Limit:   limit,
		Column3: pgLastTime,
		Column4: lastIDUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list friends: %w", err)
	}

	friends := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		friends = append(friends, domain.User{
			ID:              row.ID.String(),
			FirebaseUID:     row.FirebaseUid,
			Email:           textToPtr(row.Email),
			Phone:           textToPtr(row.Phone),
			Name:            row.Name,
			DefaultCurrency: row.DefaultCurrency,
			CreatedAt:       row.CreatedAt.Time,
			UpdatedAt:       row.UpdatedAt.Time,
		})
	}
	return friends, nil
}

// ListFriendsByStatus retrieves friends filtered by specific friendship status (ACCEPTED, PENDING, BLOCKED).
func (r *DBRepository) ListFriendsByStatus(ctx context.Context, userID string, status domain.FriendshipStatus) ([]domain.FriendWithStatus, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListFriendsByStatus(ctx, dbgen.ListFriendsByStatusParams{
		UserID: parsedID,
		Status: string(status),
	})
	if err != nil {
		return nil, fmt.Errorf("list friends by status: %w", err)
	}

	friends := make([]domain.FriendWithStatus, 0, len(rows))
	for _, row := range rows {
		actionUserID := ""
		if row.ActionUserID.Valid {
			actionUserID = uuid.UUID(row.ActionUserID.Bytes).String()
		}

		friends = append(friends, domain.FriendWithStatus{
			User: domain.User{
				ID:              row.ID.String(),
				FirebaseUID:     row.FirebaseUid,
				Email:           textToPtr(row.Email),
				Phone:           textToPtr(row.Phone),
				Name:            row.Name,
				DefaultCurrency: row.DefaultCurrency,
				CreatedAt:       row.CreatedAt.Time,
				UpdatedAt:       row.UpdatedAt.Time,
			},
			Status:       domain.FriendshipStatus(row.Status),
			ActionUserID: actionUserID,
		})
	}
	return friends, nil
}

// SyncFriendsBySequence retrieves friendship changes updated or created after lastVersion for a user.
func (r *DBRepository) SyncFriendsBySequence(ctx context.Context, lastVersion int64, userID string, limit int32) ([]domain.FriendshipSyncRecord, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.SyncFriendsBySequence(ctx, dbgen.SyncFriendsBySequenceParams{
		SyncVersion: lastVersion,
		UserID:      parsedUserID,
		Limit:       limit,
	})
	if err != nil {
		return nil, fmt.Errorf("sync friends by sequence: %w", err)
	}

	records := make([]domain.FriendshipSyncRecord, 0, len(rows))
	for _, row := range rows {
		actionUserID := ""
		if row.ActionUserID.Valid {
			actionUserID = uuid.UUID(row.ActionUserID.Bytes).String()
		}

		records = append(records, domain.FriendshipSyncRecord{
			UserID:       row.UserID.String(),
			FriendID:     row.FriendID.String(),
			Status:       domain.FriendshipStatus(row.Status),
			ActionUserID: actionUserID,
			CreatedAt:    row.CreatedAt.Time,
			UpdatedAt:    row.UpdatedAt.Time,
			SyncVersion:  row.SyncVersion,
		})
	}
	return records, nil
}

// GetFriendTombstonesBySequence retrieves deleted friendship tombstones after lastVersion for a user.
func (r *DBRepository) GetFriendTombstonesBySequence(ctx context.Context, lastVersion int64, userID string, limit int32) ([]domain.Tombstone, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.GetFriendTombstonesBySequence(ctx, dbgen.GetFriendTombstonesBySequenceParams{
		UserID:      parsedUserID,
		SyncVersion: lastVersion,
		Limit:       limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get friend tombstones: %w", err)
	}

	tombstones := make([]domain.Tombstone, 0, len(rows))
	for _, row := range rows {
		tombstones = append(tombstones, domain.Tombstone{
			EntityID:    row.EntityID.String(),
			SyncVersion: row.SyncVersion,
		})
	}
	return tombstones, nil
}

