package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository handles database operations for users.
type Repository struct {
	db *db.DB
	tm *db.TransactionManager
}

// NewRepository creates a new Repository instance.
func NewRepository(database *db.DB, tm *db.TransactionManager) *Repository {
	return &Repository{
		db: database,
		tm: tm,
	}
}

// GetByID retrieves a user by ID.
func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbUser, err := q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	return &User{
		ID:        dbUser.ID,
		Email:     textToPtr(dbUser.Email),
		Phone:     textToPtr(dbUser.Phone),
		Name:      dbUser.Name,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
	}, nil
}

// Create inserts a new user.
func (r *Repository) Create(ctx context.Context, u *User) error {
	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbUser, err := q.CreateUser(ctx, dbgen.CreateUserParams{
		ID:    u.ID,
		Email: ptrToText(u.Email),
		Phone: ptrToText(u.Phone),
		Name:  u.Name,
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
