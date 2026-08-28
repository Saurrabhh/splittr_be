-- name: GetIdempotencyKey :one
SELECT id, key, user_id, request_path, request_method, request_hash, response_code, response_headers, response_body, created_at, locked_at
FROM idempotency_keys
WHERE user_id = $1 AND key = $2;

-- name: CreateIdempotencyKey :one
INSERT INTO idempotency_keys (key, user_id, request_path, request_method, request_hash, response_code, response_headers, response_body, locked_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
ON CONFLICT (user_id, key) DO NOTHING
RETURNING id, key, user_id, request_path, request_method, request_hash, response_code, response_headers, response_body, created_at, locked_at;

-- name: UpdateIdempotencyKeyResponse :exec
UPDATE idempotency_keys
SET response_code = $3, response_headers = $4, response_body = $5
WHERE user_id = $1 AND key = $2;

-- name: DeleteExpiredIdempotencyKeys :exec
DELETE FROM idempotency_keys
WHERE created_at < $1;
