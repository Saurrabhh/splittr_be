-- name: GetUserByID :one
SELECT id, firebase_uid, email, phone, name, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByFirebaseUID :one
SELECT id, firebase_uid, email, phone, name, created_at, updated_at
FROM users
WHERE firebase_uid = $1;

-- name: CreateUser :one
INSERT INTO users (id, firebase_uid, email, phone, name, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id, firebase_uid, email, phone, name, created_at, updated_at;

