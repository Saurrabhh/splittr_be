package domain

import "context"

// Repository manages persistence of idempotency keys and cached HTTP response payloads.
type Repository interface {
	Get(ctx context.Context, userID, key string) (*IdempotencyRecord, error)
	Create(ctx context.Context, rec *IdempotencyRecord) (*IdempotencyRecord, error)
	UpdateResponse(ctx context.Context, userID, key string, code int, headers map[string]string, body []byte) error
}
