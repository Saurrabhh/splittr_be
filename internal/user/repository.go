package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

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
func (r *DBRepository) GetByID(ctx context.Context, id string) (*User, error) {
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

	return &User{
		ID:              dbUser.ID.String(),
		FirebaseUID:     dbUser.FirebaseUid,
		Email:           textToPtr(dbUser.Email),
		Phone:           textToPtr(dbUser.Phone),
		Name:            dbUser.Name,
		DefaultCurrency: dbUser.DefaultCurrency,
		CreatedAt:       dbUser.CreatedAt.Time,
		UpdatedAt:       dbUser.UpdatedAt.Time,
	}, nil
}

// GetByFirebaseUID retrieves a user by Firebase UID.
func (r *DBRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*User, error) {
	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbUser, err := q.GetUserByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by firebase uid: %w", err)
	}

	return &User{
		ID:              dbUser.ID.String(),
		FirebaseUID:     dbUser.FirebaseUid,
		Email:           textToPtr(dbUser.Email),
		Phone:           textToPtr(dbUser.Phone),
		Name:            dbUser.Name,
		DefaultCurrency: dbUser.DefaultCurrency,
		CreatedAt:       dbUser.CreatedAt.Time,
		UpdatedAt:       dbUser.UpdatedAt.Time,
	}, nil
}

// Create inserts a new user.
func (r *DBRepository) Create(ctx context.Context, u *User) error {
	parsedID, err := uuid.Parse(u.ID)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	if u.DefaultCurrency == "" {
		u.DefaultCurrency = "INR"
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
func (r *DBRepository) UpdateUser(ctx context.Context, u *User) error {
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

// GetByEmailOrPhone retrieves a user by their email or phone number.
func (r *DBRepository) GetByEmailOrPhone(ctx context.Context, email, phone string) (*User, error) {
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

	return &User{
		ID:              dbUser.ID.String(),
		FirebaseUID:     dbUser.FirebaseUid,
		Email:           textToPtr(dbUser.Email),
		Phone:           textToPtr(dbUser.Phone),
		Name:            dbUser.Name,
		DefaultCurrency: dbUser.DefaultCurrency,
		CreatedAt:       dbUser.CreatedAt.Time,
		UpdatedAt:       dbUser.UpdatedAt.Time,
	}, nil
}

// CreateFriendship creates a friendship link.
func (r *DBRepository) CreateFriendship(ctx context.Context, userID, friendID string) error {
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

	return q.CreateFriendship(ctx, dbgen.CreateFriendshipParams{
		UserID:   parsedUser,
		FriendID: parsedFriend,
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

// GetFriendship checks if two users are friends.
func (r *DBRepository) GetFriendship(ctx context.Context, userID, friendID string) (bool, error) {
	parsedUser, err := uuid.Parse(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user uuid: %w", err)
	}
	parsedFriend, err := uuid.Parse(friendID)
	if err != nil {
		return false, fmt.Errorf("invalid friend uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	_, err = q.GetFriendship(ctx, dbgen.GetFriendshipParams{
		UserID:   parsedUser,
		FriendID: parsedFriend,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query friendship: %w", err)
	}

	return true, nil
}

// ListFriends lists all friends of a user.
func (r *DBRepository) ListFriends(ctx context.Context, userID string) ([]User, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListFriends(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("list friends: %w", err)
	}

	friends := make([]User, 0, len(rows))
	for _, row := range rows {
		friends = append(friends, User{
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

func textToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func ptrToText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}
