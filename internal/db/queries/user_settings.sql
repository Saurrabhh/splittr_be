-- name: GetUserSettings :one
SELECT user_id, auto_accept_friend_requests, created_at, updated_at
FROM user_settings
WHERE user_id = $1;

-- name: CreateDefaultUserSettings :exec
INSERT INTO user_settings (user_id, auto_accept_friend_requests, created_at, updated_at)
VALUES ($1, FALSE, NOW(), NOW())
ON CONFLICT (user_id) DO NOTHING;

-- name: UpsertUserSettings :one
INSERT INTO user_settings (user_id, auto_accept_friend_requests, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
ON CONFLICT (user_id) DO UPDATE
SET auto_accept_friend_requests = EXCLUDED.auto_accept_friend_requests,
    updated_at = NOW()
RETURNING user_id, auto_accept_friend_requests, created_at, updated_at;
