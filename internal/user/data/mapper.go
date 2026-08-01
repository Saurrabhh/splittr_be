package data

import (
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/Saurrabhh/splittr_be/internal/user/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

func toDomainUser(dbUser dbgen.User) *domain.User {
	return &domain.User{
		ID:              dbUser.ID.String(),
		FirebaseUID:     dbUser.FirebaseUid,
		Email:           textToPtr(dbUser.Email),
		Phone:           textToPtr(dbUser.Phone),
		Name:            dbUser.Name,
		DefaultCurrency: dbUser.DefaultCurrency,
		CreatedAt:       dbUser.CreatedAt.Time,
		UpdatedAt:       dbUser.UpdatedAt.Time,
	}
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
