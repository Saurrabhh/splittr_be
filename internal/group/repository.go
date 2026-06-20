package group

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// DBRepository handles database operations for groups.
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

// CreateGroup inserts a new group record.
func (r *DBRepository) CreateGroup(ctx context.Context, g *Group) error {
	parsedID, err := uuid.Parse(g.ID)
	if err != nil {
		return fmt.Errorf("invalid group uuid: %w", err)
	}

	var parsedCreator uuid.UUID
	if g.CreatedBy != nil {
		parsedCreator, err = uuid.Parse(*g.CreatedBy)
		if err != nil {
			return fmt.Errorf("invalid creator uuid: %w", err)
		}
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbGroup, err := q.CreateGroup(ctx, dbgen.CreateGroupParams{
		ID:          parsedID,
		Name:        g.Name,
		Description: ptrToText(g.Description),
		InviteCode:  ptrToText(g.InviteCode),
		CreatedBy:   uuidToPg(g.CreatedBy, parsedCreator),
	})
	if err != nil {
		return fmt.Errorf("insert group: %w", err)
	}

	g.CreatedAt = dbGroup.CreatedAt.Time
	g.UpdatedAt = dbGroup.UpdatedAt.Time
	if dbGroup.ArchivedAt.Valid {
		g.ArchivedAt = &dbGroup.ArchivedAt.Time
	}

	return nil
}


// GetByID retrieves a group by its ID.
func (r *DBRepository) GetByID(ctx context.Context, id string) (*Group, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbGroup, err := q.GetGroupByID(ctx, parsedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query group: %w", err)
	}

	return toDomainGroup(dbGroup), nil
}

// GetByInviteCode retrieves a group by its invite code.
func (r *DBRepository) GetByInviteCode(ctx context.Context, inviteCode string) (*Group, error) {
	if inviteCode == "" {
		return nil, errors.New("invite code is required")
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbGroup, err := q.GetGroupByInviteCode(ctx, ptrToText(&inviteCode))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query group by invite code: %w", err)
	}

	return toDomainGroup(dbGroup), nil
}

// Update updates group name and description.
func (r *DBRepository) Update(ctx context.Context, g *Group) error {
	parsedID, err := uuid.Parse(g.ID)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbGroup, err := q.UpdateGroup(ctx, dbgen.UpdateGroupParams{
		ID:          parsedID,
		Name:        g.Name,
		Description: ptrToText(g.Description),
	})
	if err != nil {
		return fmt.Errorf("update group: %w", err)
	}

	g.UpdatedAt = dbGroup.UpdatedAt.Time
	return nil
}

// Archive soft-deletes a group by setting its archived_at timestamp.
func (r *DBRepository) Archive(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	if err := q.ArchiveGroup(ctx, parsedID); err != nil {
		return fmt.Errorf("archive group: %w", err)
	}
	return nil
}

// AddGroupMember adds a member to the group.
func (r *DBRepository) AddGroupMember(ctx context.Context, groupID, userID, role string) error {
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		return fmt.Errorf("invalid group uuid: %w", err)
	}
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	err = q.AddGroupMember(ctx, dbgen.AddGroupMemberParams{
		GroupID: parsedGroupID,
		UserID:  parsedUserID,
		Role:    role,
	})
	if err != nil {
		return fmt.Errorf("add group member: %w", err)
	}
	return nil
}

// RemoveGroupMember removes a member from the group.
func (r *DBRepository) RemoveGroupMember(ctx context.Context, groupID, userID string) error {
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		return fmt.Errorf("invalid group uuid: %w", err)
	}
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	err = q.RemoveGroupMember(ctx, dbgen.RemoveGroupMemberParams{
		GroupID: parsedGroupID,
		UserID:  parsedUserID,
	})
	if err != nil {
		return fmt.Errorf("remove group member: %w", err)
	}
	return nil
}


// UpdateGroupMemberRole updates a member's role.
func (r *DBRepository) UpdateGroupMemberRole(ctx context.Context, groupID, userID, role string) error {
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		return fmt.Errorf("invalid group uuid: %w", err)
	}
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	err = q.UpdateGroupMemberRole(ctx, dbgen.UpdateGroupMemberRoleParams{
		GroupID: parsedGroupID,
		UserID:  parsedUserID,
		Role:    role,
	})
	if err != nil {
		return fmt.Errorf("update group member role: %w", err)
	}
	return nil
}

// GetGroupMember retrieves a single member details (e.g. for membership validation).
func (r *DBRepository) GetGroupMember(ctx context.Context, groupID, userID string) (*GroupMember, error) {
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group uuid: %w", err)
	}
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	gm, err := q.GetGroupMember(ctx, dbgen.GetGroupMemberParams{
		GroupID: parsedGroupID,
		UserID:  parsedUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query group member: %w", err)
	}

	return &GroupMember{
		GroupID:  gm.GroupID.String(),
		UserID:   gm.UserID.String(),
		Role:     gm.Role,
		JoinedAt: gm.JoinedAt.Time,
	}, nil
}

// ListGroupMembers lists all members of a group with user details.
func (r *DBRepository) ListGroupMembers(ctx context.Context, groupID string) ([]GroupMember, error) {
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListGroupMembers(ctx, parsedGroupID)
	if err != nil {
		return nil, fmt.Errorf("list group members: %w", err)
	}

	members := make([]GroupMember, 0, len(rows))
	for _, row := range rows {
		members = append(members, GroupMember{
			GroupID:  row.GroupID.String(),
			UserID:   row.UserID.String(),
			Role:     row.Role,
			JoinedAt: row.JoinedAt.Time,
			Name:     row.Name,
			Email:    textToPtr(row.Email),
			Phone:    textToPtr(row.Phone),
		})
	}
	return members, nil
}

// ListUserGroups lists all groups a user is member of.
func (r *DBRepository) ListUserGroups(ctx context.Context, userID string) ([]Group, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListUserGroups(ctx, parsedUserID)
	if err != nil {
		return nil, fmt.Errorf("list user groups: %w", err)
	}

	groups := make([]Group, 0, len(rows))
	for _, row := range rows {
		groups = append(groups, *toDomainGroup(row))
	}
	return groups, nil
}

// Helper converters
func toDomainGroup(dbg dbgen.Group) *Group {
	var createdByStr *string
	if dbg.CreatedBy.Valid {
		s := uuid.UUID(dbg.CreatedBy.Bytes).String()
		createdByStr = &s
	}

	var archivedAtTime *time.Time
	if dbg.ArchivedAt.Valid {
		archivedAtTime = &dbg.ArchivedAt.Time
	}

	return &Group{
		ID:          dbg.ID.String(),
		Name:        dbg.Name,
		Description: textToPtr(dbg.Description),
		InviteCode:  textToPtr(dbg.InviteCode),
		CreatedBy:   createdByStr,
		CreatedAt:   dbg.CreatedAt.Time,
		UpdatedAt:   dbg.UpdatedAt.Time,
		ArchivedAt:  archivedAtTime,
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
