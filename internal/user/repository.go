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
		ID:          dbUser.ID.String(),
		FirebaseUID: dbUser.FirebaseUid,
		Email:       textToPtr(dbUser.Email),
		Phone:       textToPtr(dbUser.Phone),
		Name:        dbUser.Name,
		CreatedAt:   dbUser.CreatedAt.Time,
		UpdatedAt:   dbUser.UpdatedAt.Time,
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
		ID:          dbUser.ID.String(),
		FirebaseUID: dbUser.FirebaseUid,
		Email:       textToPtr(dbUser.Email),
		Phone:       textToPtr(dbUser.Phone),
		Name:        dbUser.Name,
		CreatedAt:   dbUser.CreatedAt.Time,
		UpdatedAt:   dbUser.UpdatedAt.Time,
	}, nil
}

// Create inserts a new user.
func (r *DBRepository) Create(ctx context.Context, u *User) error {
	parsedID, err := uuid.Parse(u.ID)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbUser, err := q.CreateUser(ctx, dbgen.CreateUserParams{
		ID:          parsedID,
		FirebaseUid: u.FirebaseUID,
		Email:       ptrToText(u.Email),
		Phone:       ptrToText(u.Phone),
		Name:        u.Name,
	})
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	u.CreatedAt = dbUser.CreatedAt.Time
	u.UpdatedAt = dbUser.UpdatedAt.Time

	return nil
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
