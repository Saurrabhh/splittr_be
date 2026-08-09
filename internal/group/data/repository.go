package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/Saurrabhh/splittr_be/internal/group/domain"
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
func (r *DBRepository) CreateGroup(ctx context.Context, g *domain.Group) error {
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

	var expiresAt pgtype.Timestamptz
	if g.InviteCodeExpiresAt != nil {
		expiresAt = pgtype.Timestamptz{Time: *g.InviteCodeExpiresAt, Valid: true}
	}

	dbGroup, err := q.CreateGroup(ctx, dbgen.CreateGroupParams{
		ID:                   parsedID,
		Name:                 g.Name,
		Description:          ptrToText(g.Description),
		InviteCode:           ptrToText(g.InviteCode),
		InviteCodeExpiresAt:  expiresAt,
		RequireAdminApproval: g.RequireAdminApproval,
		CreatedBy:            uuidToPg(g.CreatedBy, parsedCreator),
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
func (r *DBRepository) GetByID(ctx context.Context, id string) (*domain.Group, error) {
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

	return mapGroupFields(dbGroup.ID, dbGroup.Name, dbGroup.Description, dbGroup.InviteCode, dbGroup.InviteCodeExpiresAt, dbGroup.RequireAdminApproval, dbGroup.CreatedBy, dbGroup.CreatedAt, dbGroup.UpdatedAt, dbGroup.ArchivedAt), nil
}

// GetByInviteCode retrieves a group by its invite code.
func (r *DBRepository) GetByInviteCode(ctx context.Context, inviteCode string) (*domain.Group, error) {
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

	return mapGroupFields(dbGroup.ID, dbGroup.Name, dbGroup.Description, dbGroup.InviteCode, dbGroup.InviteCodeExpiresAt, dbGroup.RequireAdminApproval, dbGroup.CreatedBy, dbGroup.CreatedAt, dbGroup.UpdatedAt, dbGroup.ArchivedAt), nil
}

// GetPreviewByInviteCode retrieves preview details of a group using its invite code.
func (r *DBRepository) GetPreviewByInviteCode(ctx context.Context, inviteCode string) (*domain.Preview, error) {
	if inviteCode == "" {
		return nil, errors.New("invite code is required")
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	row, err := q.GetGroupPreviewByInviteCode(ctx, ptrToText(&inviteCode))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query group preview by invite code: %w", err)
	}

	var desc *string
	if row.GroupDescription.Valid {
		desc = &row.GroupDescription.String
	}

	var creatorName string
	if row.CreatorName.Valid {
		creatorName = row.CreatorName.String
	}

	return &domain.Preview{
		Name:        row.GroupName,
		Description: desc,
		MemberCount: row.MemberCount,
		CreatorName: creatorName,
	}, nil
}

// Update updates group name and description.
func (r *DBRepository) Update(ctx context.Context, g *domain.Group) error {
	parsedID, err := uuid.Parse(g.ID)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbGroup, err := q.UpdateGroup(ctx, dbgen.UpdateGroupParams{
		ID:                   parsedID,
		Name:                 g.Name,
		Description:          ptrToText(g.Description),
		RequireAdminApproval: g.RequireAdminApproval,
	})
	if err != nil {
		return fmt.Errorf("update group: %w", err)
	}

	g.UpdatedAt = dbGroup.UpdatedAt.Time
	return nil
}

// ResetInviteCode resets group invite code and expiration timestamp.
func (r *DBRepository) ResetInviteCode(ctx context.Context, groupID, newInviteCode string, expiresAt time.Time) (*domain.Group, error) {
	parsedID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbGroup, err := q.ResetGroupInviteCode(ctx, dbgen.ResetGroupInviteCodeParams{
		ID:                  parsedID,
		InviteCode:          ptrToText(&newInviteCode),
		InviteCodeExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("reset group invite code: %w", err)
	}

	return mapGroupFields(dbGroup.ID, dbGroup.Name, dbGroup.Description, dbGroup.InviteCode, dbGroup.InviteCodeExpiresAt, dbGroup.RequireAdminApproval, dbGroup.CreatedBy, dbGroup.CreatedAt, dbGroup.UpdatedAt, dbGroup.ArchivedAt), nil
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

func parseGroupIDAndUserID(groupID, userID string) (uuid.UUID, uuid.UUID, error) {
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("invalid group uuid: %w", err)
	}
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("invalid user uuid: %w", err)
	}
	return parsedGroupID, parsedUserID, nil
}

// AddGroupMember adds a member to the group with status.
func (r *DBRepository) AddGroupMember(ctx context.Context, groupID, userID string, role domain.MemberRole, status domain.MemberStatus) error {
	parsedGroupID, parsedUserID, err := parseGroupIDAndUserID(groupID, userID)
	if err != nil {
		return err
	}
	if status == "" {
		status = domain.MemberStatusActive
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	err = q.AddGroupMember(ctx, dbgen.AddGroupMemberParams{
		GroupID: parsedGroupID,
		UserID:  parsedUserID,
		Role:    dbgen.MemberRole(role),
		Status:  dbgen.MemberStatus(status),
	})
	if err != nil {
		return fmt.Errorf("add group member: %w", err)
	}
	return nil
}

// AddGroupMembers adds multiple members to the group with the given role and status, returning the added members.
func (r *DBRepository) AddGroupMembers(ctx context.Context, groupID string, userIDs []string, role domain.MemberRole, status domain.MemberStatus) ([]domain.Member, error) {
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group uuid: %w", err)
	}

	if len(userIDs) == 0 {
		return []domain.Member{}, nil
	}

	parsedUserIDs := make([]uuid.UUID, 0, len(userIDs))
	userMap := make(map[string]bool, len(userIDs))
	for _, uID := range userIDs {
		parsed, err := uuid.Parse(uID)
		if err != nil {
			return nil, fmt.Errorf("invalid user uuid: %w", err)
		}
		parsedUserIDs = append(parsedUserIDs, parsed)
		userMap[uID] = true
	}

	if status == "" {
		status = domain.MemberStatusActive
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	err = q.AddGroupMembers(ctx, dbgen.AddGroupMembersParams{
		GroupID: parsedGroupID,
		UserIds: parsedUserIDs,
		Role:    dbgen.MemberRole(role),
		Status:  dbgen.MemberStatus(status),
	})
	if err != nil {
		return nil, fmt.Errorf("add group members: %w", err)
	}

	allMembers, err := r.ListGroupMembers(ctx, groupID, "")
	if err != nil {
		return nil, fmt.Errorf("fetch group members after batch add: %w", err)
	}

	addedMembers := make([]domain.Member, 0, len(userIDs))
	for _, m := range allMembers {
		if userMap[m.UserID] {
			addedMembers = append(addedMembers, m)
		}
	}

	return addedMembers, nil
}

// UpdateMemberStatus updates a member's status in a group.
func (r *DBRepository) UpdateMemberStatus(ctx context.Context, groupID, userID string, status domain.MemberStatus) error {
	parsedGroupID, parsedUserID, err := parseGroupIDAndUserID(groupID, userID)
	if err != nil {
		return err
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	err = q.UpdateMemberStatus(ctx, dbgen.UpdateMemberStatusParams{
		GroupID: parsedGroupID,
		UserID:  parsedUserID,
		Status:  dbgen.MemberStatus(status),
	})
	if err != nil {
		return fmt.Errorf("update member status: %w", err)
	}
	return nil
}

// RemoveGroupMember removes a member from the group.
func (r *DBRepository) RemoveGroupMember(ctx context.Context, groupID, userID string) error {
	parsedGroupID, parsedUserID, err := parseGroupIDAndUserID(groupID, userID)
	if err != nil {
		return err
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
func (r *DBRepository) UpdateGroupMemberRole(ctx context.Context, groupID, userID string, role domain.MemberRole) error {
	parsedGroupID, parsedUserID, err := parseGroupIDAndUserID(groupID, userID)
	if err != nil {
		return err
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	err = q.UpdateGroupMemberRole(ctx, dbgen.UpdateGroupMemberRoleParams{
		GroupID: parsedGroupID,
		UserID:  parsedUserID,
		Role:    dbgen.MemberRole(role),
	})
	if err != nil {
		return fmt.Errorf("update group member role: %w", err)
	}
	return nil
}

// GetGroupMember retrieves a single member details (e.g. for membership validation).
func (r *DBRepository) GetGroupMember(ctx context.Context, groupID, userID string) (*domain.Member, error) {
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

	return &domain.Member{
		GroupID:  gm.GroupID.String(),
		UserID:   gm.UserID.String(),
		Role:     domain.MemberRole(gm.Role),
		Status:   domain.MemberStatus(gm.Status),
		JoinedAt: gm.JoinedAt.Time,
	}, nil
}

// ListGroupMembers lists members of a group with user details, optionally filtered by status.
func (r *DBRepository) ListGroupMembers(ctx context.Context, groupID string, status domain.MemberStatus) ([]domain.Member, error) {
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListGroupMembers(ctx, dbgen.ListGroupMembersParams{
		GroupID: parsedGroupID,
		Column2: string(status),
	})
	if err != nil {
		return nil, fmt.Errorf("list group members: %w", err)
	}

	members := make([]domain.Member, 0, len(rows))
	for _, row := range rows {
		members = append(members, domain.Member{
			GroupID:  row.GroupID.String(),
			UserID:   row.UserID.String(),
			Role:     domain.MemberRole(row.Role),
			Status:   domain.MemberStatus(row.Status),
			JoinedAt: row.JoinedAt.Time,
			Name:     row.Name,
			Email:    textToPtr(row.Email),
			Phone:    textToPtr(row.Phone),
		})
	}
	return members, nil
}

// ListUserGroupsWithMembers lists all groups a user is an ACTIVE member of with cursor-based pagination.
func (r *DBRepository) ListUserGroupsWithMembers(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.GroupWithMembers, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user uuid: %w", err)
	}

	var pgLastTime pgtype.Timestamptz
	if lastTime != nil {
		pgLastTime = pgtype.Timestamptz{Time: *lastTime, Valid: true}
	}
	var lastIDUUID uuid.UUID
	if lastID != nil {
		if parsed, err := uuid.Parse(*lastID); err == nil {
			lastIDUUID = parsed
		}
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListUserGroupsWithMembersPaginated(ctx, dbgen.ListUserGroupsWithMembersPaginatedParams{
		UserID:  parsedUserID,
		Limit:   limit,
		Column3: pgLastTime,
		Column4: lastIDUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list user groups with members paginated: %w", err)
	}

	result := make([]domain.GroupWithMembers, 0, len(rows))
	for _, row := range rows {
		g := mapGroupFields(row.ID, row.Name, row.Description, row.InviteCode, row.InviteCodeExpiresAt, row.RequireAdminApproval, row.CreatedBy, row.CreatedAt, row.UpdatedAt, row.ArchivedAt)

		var members []domain.Member
		if err := json.Unmarshal(row.Members, &members); err != nil {
			return nil, fmt.Errorf("decode members json for group %s: %w", row.ID, err)
		}

		result = append(result, domain.GroupWithMembers{Group: *g, Members: members})
	}
	return result, nil
}
