-- +goose Up
CREATE TABLE IF NOT EXISTS idempotency_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    request_path VARCHAR(255) NOT NULL,
    request_method VARCHAR(16) NOT NULL,
    request_hash VARCHAR(64) NOT NULL,
    response_code INT,
    response_headers JSONB,
    response_body BYTEA,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    locked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT unq_idempotency_user_key UNIQUE (user_id, key)
);

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_user_key ON idempotency_keys(user_id, key);
CREATE INDEX IF NOT EXISTS idx_idempotency_created_at ON idempotency_keys(created_at);

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys;
