-- +goose Up
CREATE TABLE IF NOT EXISTS user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    auto_accept_friend_requests BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Backfill user_settings for existing users
INSERT INTO user_settings (user_id, auto_accept_friend_requests, created_at, updated_at)
SELECT id, FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP FROM users
ON CONFLICT (user_id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS user_settings;
