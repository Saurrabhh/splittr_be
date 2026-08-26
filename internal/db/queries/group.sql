-- name: CreateGroup :one
INSERT INTO groups (id, name, description, invite_code, invite_code_expires_at, require_admin_approval, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
RETURNING id, name, description, invite_code, invite_code_expires_at, require_admin_approval, created_by, created_at, updated_at, archived_at, icon_url;

-- name: GetGroupByID :one
SELECT id, name, description, invite_code, invite_code_expires_at, require_admin_approval, created_by, created_at, updated_at, archived_at, icon_url
FROM groups
WHERE id = $1 AND archived_at IS NULL;

-- name: UpdateGroup :one
UPDATE groups
SET name = $2, description = $3, require_admin_approval = $4, updated_at = NOW()
WHERE id = $1 AND archived_at IS NULL
RETURNING id, name, description, invite_code, invite_code_expires_at, require_admin_approval, created_by, created_at, updated_at, archived_at, icon_url;

-- name: ArchiveGroup :exec
UPDATE groups
SET archived_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: AddGroupMembers :exec
INSERT INTO group_members (group_id, user_id, role, status, joined_at)
SELECT sqlc.arg(group_id)::uuid, unnest(sqlc.arg(user_ids)::uuid[]), sqlc.arg(role)::text, sqlc.arg(status)::text, NOW()
ON CONFLICT (group_id, user_id) DO NOTHING;

-- name: RemoveGroupMember :exec
DELETE FROM group_members
WHERE group_id = $1 AND user_id = $2;

-- name: UpdateGroupMemberRole :exec
UPDATE group_members
SET role = $3
WHERE group_id = $1 AND user_id = $2;

-- name: UpdateMemberStatus :exec
UPDATE group_members
SET status = $3
WHERE group_id = $1 AND user_id = $2;

-- name: ResetGroupInviteCode :one
UPDATE groups
SET invite_code = $2, invite_code_expires_at = $3, updated_at = NOW()
WHERE id = $1 AND archived_at IS NULL
RETURNING id, name, description, invite_code, invite_code_expires_at, require_admin_approval, created_by, created_at, updated_at, archived_at, icon_url;

-- name: GetGroupMember :one
SELECT group_id, user_id, role, status, joined_at
FROM group_members
WHERE group_id = $1 AND user_id = $2;

-- name: ListGroupMembers :many
SELECT gm.group_id, gm.user_id, gm.role, gm.status, gm.joined_at, u.name, u.email, u.phone
FROM group_members gm
JOIN users u ON gm.user_id = u.id
WHERE gm.group_id = $1 AND ($2::text = '' OR gm.status::text = $2::text);

-- name: ListUserGroups :many
SELECT g.id, g.name, g.description, g.invite_code, g.invite_code_expires_at, g.require_admin_approval, g.created_by, g.created_at, g.updated_at, g.archived_at, g.icon_url
FROM groups g
JOIN group_members gm ON g.id = gm.group_id
WHERE gm.user_id = $1 AND gm.status = 'ACTIVE' AND g.archived_at IS NULL;

-- name: GetGroupByInviteCode :one
SELECT id, name, description, invite_code, invite_code_expires_at, require_admin_approval, created_by, created_at, updated_at, archived_at, icon_url
FROM groups
WHERE invite_code = $1 AND archived_at IS NULL;

-- name: ListUserGroupsWithMembers :many
SELECT
    g.id, g.name, g.description, g.invite_code, g.invite_code_expires_at, g.require_admin_approval, g.created_by,
    g.created_at, g.updated_at, g.archived_at, g.icon_url,
    COALESCE(
        json_agg(
            json_build_object(
                'groupId',  gm2.group_id,
                'userId',   gm2.user_id,
                'role',     gm2.role,
                'status',   gm2.status,
                'joinedAt', gm2.joined_at,
                'name',     u.name,
                'email',    u.email,
                'phone',    u.phone
            ) ORDER BY gm2.joined_at
        ) FILTER (WHERE gm2.user_id IS NOT NULL),
        '[]'
    )::jsonb AS members
FROM groups g
JOIN group_members gm ON gm.group_id = g.id AND gm.user_id = $1 AND gm.status = 'ACTIVE'
LEFT JOIN group_members gm2 ON gm2.group_id = g.id
LEFT JOIN users u ON u.id = gm2.user_id
WHERE g.archived_at IS NULL
GROUP BY g.id
ORDER BY g.created_at DESC;

-- name: ListUserGroupsWithMembersPaginated :many
SELECT
    g.id, g.name, g.description, g.invite_code, g.invite_code_expires_at, g.require_admin_approval, g.created_by,
    g.created_at, g.updated_at, g.archived_at, g.icon_url,
    COALESCE(
        json_agg(
            json_build_object(
                'groupId',  gm2.group_id,
                'userId',   gm2.user_id,
                'role',     gm2.role,
                'status',   gm2.status,
                'joinedAt', gm2.joined_at,
                'name',     u.name,
                'email',    u.email,
                'phone',    u.phone
            ) ORDER BY gm2.joined_at
        ) FILTER (WHERE gm2.user_id IS NOT NULL),
        '[]'
    )::jsonb AS members
FROM groups g
JOIN group_members gm ON gm.group_id = g.id AND gm.user_id = $1 AND gm.status = 'ACTIVE'
LEFT JOIN group_members gm2 ON gm2.group_id = g.id
LEFT JOIN users u ON u.id = gm2.user_id
WHERE g.archived_at IS NULL
  AND (
    $3::TIMESTAMP WITH TIME ZONE IS NULL
    OR g.created_at < $3::TIMESTAMP WITH TIME ZONE
    OR (g.created_at = $3::TIMESTAMP WITH TIME ZONE AND g.id < $4::UUID)
  )
GROUP BY g.id
ORDER BY g.created_at DESC, g.id DESC
LIMIT $2;

-- name: GetGroupPreviewByInviteCode :one
SELECT 
    g.name AS group_name,
    g.description AS group_description,
    COUNT(gm.user_id)::BIGINT AS member_count,
    u.name AS creator_name
FROM groups g
LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.status = 'ACTIVE'
LEFT JOIN users u ON g.created_by = u.id
WHERE g.invite_code = $1 AND g.archived_at IS NULL
GROUP BY g.id, g.name, g.description, u.name;

-- name: SyncGroupsBySequence :many
SELECT 
    g.id, g.name, g.description, g.invite_code, g.invite_code_expires_at, g.require_admin_approval, g.created_by, 
    g.created_at, g.updated_at, g.archived_at, g.icon_url, g.sync_version,
    COALESCE(
        json_agg(
            json_build_object(
                'groupId',  gm2.group_id,
                'userId',   gm2.user_id,
                'role',     gm2.role,
                'status',   gm2.status,
                'joinedAt', gm2.joined_at,
                'name',     u.name,
                'email',    u.email,
                'phone',    u.phone
            ) ORDER BY gm2.joined_at
        ) FILTER (WHERE gm2.user_id IS NOT NULL),
        '[]'
    )::jsonb AS members
FROM groups g
JOIN group_members gm ON g.id = gm.group_id AND gm.user_id = $2
LEFT JOIN group_members gm2 ON gm2.group_id = g.id
LEFT JOIN users u ON u.id = gm2.user_id
WHERE g.sync_version > $1
GROUP BY g.id
ORDER BY g.sync_version ASC
LIMIT $3;

-- name: GetGroupTombstonesBySequence :many
SELECT entity_id, sync_version
FROM entity_tombstones
WHERE entity_type = 'GROUP'
  AND user_id = $1
  AND sync_version > $2
ORDER BY sync_version ASC
LIMIT $3;

-- name: UpdateGroupIcon :one
UPDATE groups
SET icon_url = $2, updated_at = NOW()
WHERE id = $1 AND archived_at IS NULL
RETURNING id, name, description, invite_code, invite_code_expires_at, require_admin_approval, created_by, created_at, updated_at, archived_at, icon_url;

