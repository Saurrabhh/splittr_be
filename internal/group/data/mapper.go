package data

import (
	"time"

	"github.com/Saurrabhh/splittr_be/internal/group/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func mapGroupFields(id uuid.UUID, name string, description pgtype.Text, inviteCode pgtype.Text, inviteCodeExpiresAt pgtype.Timestamptz, requireAdminApproval bool, createdBy pgtype.UUID, createdAt pgtype.Timestamptz, updatedAt pgtype.Timestamptz, archivedAt pgtype.Timestamptz) *domain.Group {
	var createdByStr *string
	if createdBy.Valid {
		createdByStr = new(uuid.UUID(createdBy.Bytes).String())
	}

	var archivedAtTime *time.Time
	if archivedAt.Valid {
		archivedAtTime = &archivedAt.Time
	}

	var inviteExpiresAt *time.Time
	if inviteCodeExpiresAt.Valid {
		inviteExpiresAt = &inviteCodeExpiresAt.Time
	}

	return &domain.Group{
		ID:                   id.String(),
		Name:                 name,
		Description:          textToPtr(description),
		InviteCode:           textToPtr(inviteCode),
		InviteCodeExpiresAt:  inviteExpiresAt,
		RequireAdminApproval: requireAdminApproval,
		CreatedBy:            createdByStr,
		CreatedAt:            createdAt.Time,
		UpdatedAt:            updatedAt.Time,
		ArchivedAt:           archivedAtTime,
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

func uuidToPg(s *string, u uuid.UUID) pgtype.UUID {
	if s == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: u, Valid: true}
}
