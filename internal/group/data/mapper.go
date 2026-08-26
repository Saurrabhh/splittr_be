package data

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/Saurrabhh/splittr_be/internal/group/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func mapGroupFields(id uuid.UUID, name string, description pgtype.Text, inviteCode pgtype.Text, inviteCodeExpiresAt pgtype.Timestamptz, requireAdminApproval bool, createdBy pgtype.UUID, createdAt pgtype.Timestamptz, updatedAt pgtype.Timestamptz, archivedAt pgtype.Timestamptz, iconUrl pgtype.Text) *domain.Group {
	var createdByStr *string
	if createdBy.Valid {
		s := uuid.UUID(createdBy.Bytes).String()
		createdByStr = &s
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
		IconURL:              textToPtr(iconUrl),
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

func toGroupsWithMembersFromSyncBySequence(rows []dbgen.SyncGroupsBySequenceRow) ([]domain.GroupWithMembers, error) {
	result := make([]domain.GroupWithMembers, 0, len(rows))
	for _, r := range rows {
		g := mapGroupFields(r.ID, r.Name, r.Description, r.InviteCode, r.InviteCodeExpiresAt, r.RequireAdminApproval, r.CreatedBy, r.CreatedAt, r.UpdatedAt, r.ArchivedAt, r.IconUrl)
		g.SyncVersion = r.SyncVersion

		var members []domain.Member
		if len(r.Members) > 0 {
			if err := json.Unmarshal(r.Members, &members); err != nil {
				return nil, fmt.Errorf("decode members json for group %s: %w", r.ID, err)
			}
		}
		if members == nil {
			members = []domain.Member{}
		}

		result = append(result, domain.GroupWithMembers{
			Group:   *g,
			Members: members,
		})
	}
	return result, nil
}



