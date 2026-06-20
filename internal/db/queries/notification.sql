-- name: CreateNotification :one
INSERT INTO notifications (id, user_id, actor_id, activity_id, title, content, is_read, created_at)
VALUES ($1, $2, $3, $4, $5, $6, FALSE, NOW())
RETURNING id, user_id, actor_id, activity_id, title, content, is_read, created_at;

-- name: ListUserNotifications :many
SELECT n.id, n.user_id, n.actor_id, n.activity_id, n.title, n.content, n.is_read, n.created_at, u.name as actor_name
FROM notifications n
LEFT JOIN users u ON n.actor_id = u.id
WHERE n.user_id = $1
ORDER BY n.created_at DESC;

-- name: MarkNotificationAsRead :exec
UPDATE notifications
SET is_read = TRUE
WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsAsRead :exec
UPDATE notifications
SET is_read = TRUE
WHERE user_id = $1;
