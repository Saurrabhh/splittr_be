-- name: CreateGroup :one
INSERT INTO groups (id, name, description, invite_code, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id, name, description, invite_code, created_by, created_at, updated_at, archived_at;

-- name: GetGroupByID :one
SELECT id, name, description, invite_code, created_by, created_at, updated_at, archived_at
FROM groups
WHERE id = $1 AND archived_at IS NULL;

-- name: UpdateGroup :one
UPDATE groups
SET name = $2, description = $3, updated_at = NOW()
WHERE id = $1 AND archived_at IS NULL
RETURNING id, name, description, invite_code, created_by, created_at, updated_at, archived_at;

-- name: ArchiveGroup :exec
UPDATE groups
SET archived_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: AddGroupMember :exec
INSERT INTO group_members (group_id, user_id, role, joined_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (group_id, user_id) DO NOTHING;

-- name: RemoveGroupMember :exec
DELETE FROM group_members
WHERE group_id = $1 AND user_id = $2;

-- name: UpdateGroupMemberRole :exec
UPDATE group_members
SET role = $3
WHERE group_id = $1 AND user_id = $2;

-- name: GetGroupMember :one
SELECT group_id, user_id, role, joined_at
FROM group_members
WHERE group_id = $1 AND user_id = $2;

-- name: ListGroupMembers :many
SELECT gm.group_id, gm.user_id, gm.role, gm.joined_at, u.name, u.email, u.phone
FROM group_members gm
JOIN users u ON gm.user_id = u.id
WHERE gm.group_id = $1;

-- name: ListUserGroups :many
SELECT g.id, g.name, g.description, g.invite_code, g.created_by, g.created_at, g.updated_at, g.archived_at
FROM groups g
JOIN group_members gm ON g.id = gm.group_id
WHERE gm.user_id = $1 AND g.archived_at IS NULL;

-- name: GetGroupByInviteCode :one
SELECT id, name, description, invite_code, created_by, created_at, updated_at, archived_at
FROM groups
WHERE invite_code = $1 AND archived_at IS NULL;
