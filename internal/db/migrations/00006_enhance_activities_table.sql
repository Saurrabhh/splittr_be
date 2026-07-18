-- +goose Up
ALTER TABLE activities 
ADD COLUMN IF NOT EXISTS entity_type VARCHAR(50) NOT NULL DEFAULT 'SYSTEM',
ADD COLUMN IF NOT EXISTS entity_id UUID,
ADD COLUMN IF NOT EXISTS metadata JSONB;

CREATE INDEX IF NOT EXISTS idx_activities_feed_cursor 
ON activities (group_id, created_at DESC, id DESC) 
WHERE group_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_activities_feed_cursor;
ALTER TABLE activities DROP COLUMN IF EXISTS metadata;
ALTER TABLE activities DROP COLUMN IF EXISTS entity_id;
ALTER TABLE activities DROP COLUMN IF EXISTS entity_type;
