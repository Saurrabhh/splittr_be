-- name: CreateActivity :one
INSERT INTO activities (id, group_id, actor_id, action_type, description, entity_type, entity_id, metadata, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
RETURNING id, group_id, actor_id, action_type, description, entity_type, entity_id, metadata, created_at;

-- name: CreateActivityVisibility :exec
INSERT INTO activity_visibility (activity_id, user_id)
VALUES ($1, $2);

-- name: ListUserActivities :many
SELECT a.id, a.group_id, a.actor_id, a.action_type, a.description, a.created_at, u.name as actor_name
FROM activities a
LEFT JOIN users u ON a.actor_id = u.id
WHERE 
    a.group_id IN (
        SELECT gm.group_id FROM group_members gm WHERE gm.user_id = $1
    )
    OR
    (a.group_id IS NULL AND EXISTS (
        SELECT 1 FROM activity_visibility av WHERE av.activity_id = a.id AND av.user_id = $1
    ))
ORDER BY a.created_at DESC;

-- name: ListGroupFeedPaginated :many
SELECT 
    a.id, 
    a.group_id, 
    a.actor_id, 
    COALESCE(u.name, 'System')::varchar as actor_name, 
    a.entity_type, 
    a.entity_id, 
    a.action_type, 
    a.description, 
    a.metadata, 
    a.created_at
FROM activities a
LEFT JOIN users u ON a.actor_id = u.id
WHERE a.group_id = $1
  AND EXISTS (
      SELECT 1 FROM group_members gm WHERE gm.group_id = $1 AND gm.user_id = $5
  )
  AND (
    $3::TIMESTAMP WITH TIME ZONE IS NULL 
    OR a.created_at < $3::TIMESTAMP WITH TIME ZONE
    OR (a.created_at = $3::TIMESTAMP WITH TIME ZONE AND a.id < $4::UUID)
  )
ORDER BY a.created_at DESC, a.id DESC
LIMIT $2;

-- name: ListUserActivitiesPaginated :many
SELECT 
    a.id, 
    a.group_id, 
    a.actor_id, 
    COALESCE(u.name, 'System')::varchar as actor_name, 
    a.entity_type, 
    a.entity_id, 
    a.action_type, 
    a.description, 
    a.metadata, 
    a.created_at
FROM activities a
LEFT JOIN users u ON a.actor_id = u.id
WHERE (
    a.group_id IN (
        SELECT gm.group_id FROM group_members gm WHERE gm.user_id = $1
    )
    OR
    (a.group_id IS NULL AND EXISTS (
        SELECT 1 FROM activity_visibility av WHERE av.activity_id = a.id AND av.user_id = $1
    ))
)
AND (
  $3::TIMESTAMP WITH TIME ZONE IS NULL
  OR a.created_at < $3::TIMESTAMP WITH TIME ZONE
  OR (a.created_at = $3::TIMESTAMP WITH TIME ZONE AND a.id < $4::UUID)
)
ORDER BY a.created_at DESC, a.id DESC
LIMIT $2;

