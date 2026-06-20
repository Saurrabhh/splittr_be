-- name: GetUserByID :one
SELECT id, firebase_uid, email, phone, name, default_currency, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByFirebaseUID :one
SELECT id, firebase_uid, email, phone, name, default_currency, created_at, updated_at
FROM users
WHERE firebase_uid = $1;

-- name: CreateUser :one
INSERT INTO users (id, firebase_uid, email, phone, name, default_currency, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING id, firebase_uid, email, phone, name, default_currency, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
SET name = $2, default_currency = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, firebase_uid, email, phone, name, default_currency, created_at, updated_at;

-- name: GetUserByEmailOrPhone :one
SELECT id, firebase_uid, email, phone, name, default_currency, created_at, updated_at
FROM users
WHERE email = $1 OR phone = $2;

-- name: CreateFriendship :exec
INSERT INTO friendships (user_id, friend_id)
VALUES ($1, $2);

-- name: DeleteFriendship :exec
DELETE FROM friendships
WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1);

-- name: GetFriendship :one
SELECT user_id, friend_id, created_at
FROM friendships
WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1);

-- name: ListFriends :many
SELECT u.id, u.firebase_uid, u.email, u.phone, u.name, u.default_currency, u.created_at, u.updated_at
FROM users u
WHERE u.id IN (
    SELECT f.friend_id FROM friendships f WHERE f.user_id = $1
    UNION
    SELECT f.user_id FROM friendships f WHERE f.friend_id = $1
);

