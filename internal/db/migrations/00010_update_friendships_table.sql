-- +goose Up
ALTER TABLE friendships
ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ACCEPTED', 'DECLINED', 'BLOCKED')),
ADD COLUMN action_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Backfill existing friendships to ACCEPTED status
UPDATE friendships
SET status = 'ACCEPTED', updated_at = CURRENT_TIMESTAMP
WHERE status = 'PENDING';

CREATE INDEX idx_friendships_user_status ON friendships(user_id, status);
CREATE INDEX idx_friendships_friend_status ON friendships(friend_id, status);

-- +goose Down
DROP INDEX IF EXISTS idx_friendships_friend_status;
DROP INDEX IF EXISTS idx_friendships_user_status;
ALTER TABLE friendships
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS action_user_id,
DROP COLUMN IF EXISTS status;
