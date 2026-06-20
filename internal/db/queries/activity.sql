-- name: CreateActivity :one
INSERT INTO activities (id, group_id, actor_id, action_type, description, created_at)
VALUES ($1, $2, $3, $4, $5, NOW())
RETURNING id, group_id, actor_id, action_type, description, created_at;

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
