-- +goose Up

ALTER TABLE activities ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE activities ADD COLUMN IF NOT EXISTS metadata BYTEA;
ALTER TABLE activities ALTER COLUMN entity_id DROP NOT NULL;
ALTER TABLE activities ALTER COLUMN actor_id DROP NOT NULL;

CREATE TABLE IF NOT EXISTS activity_visibility (
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (activity_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_activity_visibility_user_id ON activity_visibility(user_id);

ALTER TABLE notifications ADD COLUMN IF NOT EXISTS activity_id UUID REFERENCES activities(id) ON DELETE CASCADE;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS title VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS content TEXT NOT NULL DEFAULT '';
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS type VARCHAR(50) NOT NULL DEFAULT '';
ALTER TABLE notifications ALTER COLUMN actor_id DROP NOT NULL;

-- +goose Down
ALTER TABLE notifications DROP COLUMN IF EXISTS type;
ALTER TABLE notifications DROP COLUMN IF EXISTS content;
ALTER TABLE notifications DROP COLUMN IF EXISTS title;
ALTER TABLE notifications DROP COLUMN IF EXISTS activity_id;
DROP TABLE IF EXISTS activity_visibility;
ALTER TABLE activities DROP COLUMN IF EXISTS metadata;
ALTER TABLE activities DROP COLUMN IF EXISTS description;
