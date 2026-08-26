package data

import (
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/Saurrabhh/splittr_be/internal/user/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func mapUserFields(id uuid.UUID, firebaseUID string, email pgtype.Text, phone pgtype.Text, name string, defaultCurrency string, avatarUrl pgtype.Text, createdAt pgtype.Timestamptz, updatedAt pgtype.Timestamptz) *domain.User {
	return &domain.User{
		ID:              id.String(),
		FirebaseUID:     firebaseUID,
		Email:           textToPtr(email),
		Phone:           textToPtr(phone),
		Name:            name,
		DefaultCurrency: defaultCurrency,
		AvatarURL:       textToPtr(avatarUrl),
		CreatedAt:       createdAt.Time,
		UpdatedAt:       updatedAt.Time,
	}
}

func toDomainUser(dbUser dbgen.User) *domain.User {
	return mapUserFields(dbUser.ID, dbUser.FirebaseUid, dbUser.Email, dbUser.Phone, dbUser.Name, dbUser.DefaultCurrency, dbUser.AvatarUrl, dbUser.CreatedAt, dbUser.UpdatedAt)
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
