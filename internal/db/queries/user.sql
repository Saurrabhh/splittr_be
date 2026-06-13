-- name: GetUserByID :one
SELECT id, email, phone, name, created_at, updated_at
FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (id, email, phone, name, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id, email, phone, name, created_at, updated_at;
