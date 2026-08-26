-- +goose Up

-- 1. Composite B-Tree Cursor Indexes
CREATE INDEX IF NOT EXISTS idx_groups_cursor 
ON groups (created_at DESC, id DESC) 
WHERE archived_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_expenses_group_cursor 
ON expenses (group_id, created_at DESC, id DESC) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_expenses_personal_cursor 
ON expenses (paid_by, created_at DESC, id DESC) 
WHERE group_id IS NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_expenses_cursor 
ON expenses (created_at DESC, id DESC) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_activities_group_cursor 
ON activities (group_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_activities_cursor 
ON activities (created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_activity_visibility_user_activity 
ON activity_visibility (user_id, activity_id);

CREATE INDEX IF NOT EXISTS idx_notifications_user_cursor 
ON notifications (user_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_users_cursor 
ON users (created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_friendships_user_status 
ON friendships (user_id, status);

CREATE INDEX IF NOT EXISTS idx_friendships_friend_status 
ON friendships (friend_id, status);

-- 2. Entity Tombstones Table for Sync Deletions
CREATE TABLE IF NOT EXISTS entity_tombstones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(32) NOT NULL, -- 'GROUP', 'FRIENDSHIP', 'EXPENSE'
    entity_id UUID NOT NULL,
    user_id UUID NOT NULL,
    sync_version BIGINT NOT NULL DEFAULT nextval('global_sync_seq'),
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_entity_tombstones_user_sync 
ON entity_tombstones (user_id, sync_version);

CREATE INDEX IF NOT EXISTS idx_entity_tombstones_lookup 
ON entity_tombstones (entity_type, entity_id);

-- 3. Trigger: Touch Group Sync Version on Member Changes
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION touch_group_sync_version()
RETURNS TRIGGER AS $$
DECLARE
    v_group_id UUID;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        v_group_id := OLD.group_id;
    ELSE
        v_group_id := NEW.group_id;
    END IF;

    UPDATE groups
    SET sync_version = nextval('global_sync_seq'), updated_at = NOW()
    WHERE id = v_group_id;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_group_members_touch_group ON group_members;
CREATE TRIGGER trg_group_members_touch_group
AFTER INSERT OR UPDATE OR DELETE ON group_members
FOR EACH ROW EXECUTE FUNCTION touch_group_sync_version();

-- 4. Trigger: Record Tombstone on Group Member Removal/Leave
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION record_group_membership_tombstone()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO entity_tombstones (entity_type, entity_id, user_id, sync_version)
        VALUES ('GROUP', OLD.group_id, OLD.user_id, nextval('global_sync_seq'));
    ELSIF (TG_OP = 'UPDATE' AND NEW.status = 'REJECTED' AND OLD.status != 'REJECTED') THEN
        INSERT INTO entity_tombstones (entity_type, entity_id, user_id, sync_version)
        VALUES ('GROUP', NEW.group_id, NEW.user_id, nextval('global_sync_seq'));
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_group_members_tombstone ON group_members;
CREATE TRIGGER trg_group_members_tombstone
AFTER DELETE OR UPDATE OF status ON group_members
FOR EACH ROW EXECUTE FUNCTION record_group_membership_tombstone();

-- 5. Trigger: Record Tombstone on Group Archive
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION record_group_archive_tombstones()
RETURNS TRIGGER AS $$
BEGIN
    IF (NEW.archived_at IS NOT NULL AND OLD.archived_at IS NULL) THEN
        INSERT INTO entity_tombstones (entity_type, entity_id, user_id, sync_version)
        SELECT 'GROUP', NEW.id, gm.user_id, nextval('global_sync_seq')
        FROM group_members gm
        WHERE gm.group_id = NEW.id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_groups_archive_tombstone ON groups;
CREATE TRIGGER trg_groups_archive_tombstone
AFTER UPDATE OF archived_at ON groups
FOR EACH ROW EXECUTE FUNCTION record_group_archive_tombstones();

-- 6. Trigger: Record Tombstones on Friendship Deletion/Rejection
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION record_friendship_tombstones()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO entity_tombstones (entity_type, entity_id, user_id, sync_version)
        VALUES ('FRIENDSHIP', OLD.friend_id, OLD.user_id, nextval('global_sync_seq')),
               ('FRIENDSHIP', OLD.user_id, OLD.friend_id, nextval('global_sync_seq'));
    ELSIF (TG_OP = 'UPDATE' AND NEW.status IN ('REJECTED', 'BLOCKED') AND OLD.status = 'ACCEPTED') THEN
        INSERT INTO entity_tombstones (entity_type, entity_id, user_id, sync_version)
        VALUES ('FRIENDSHIP', NEW.friend_id, NEW.user_id, nextval('global_sync_seq')),
               ('FRIENDSHIP', NEW.user_id, NEW.friend_id, nextval('global_sync_seq'));
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_friendships_tombstone ON friendships;
CREATE TRIGGER trg_friendships_tombstone
AFTER DELETE OR UPDATE OF status ON friendships
FOR EACH ROW EXECUTE FUNCTION record_friendship_tombstones();

-- +goose Down
DROP TRIGGER IF EXISTS trg_friendships_tombstone ON friendships;
DROP FUNCTION IF EXISTS record_friendship_tombstones();

DROP TRIGGER IF EXISTS trg_groups_archive_tombstone ON groups;
DROP FUNCTION IF EXISTS record_group_archive_tombstones();

DROP TRIGGER IF EXISTS trg_group_members_tombstone ON group_members;
DROP FUNCTION IF EXISTS record_group_membership_tombstone();

DROP TRIGGER IF EXISTS trg_group_members_touch_group ON group_members;
DROP FUNCTION IF EXISTS touch_group_sync_version();

DROP TABLE IF EXISTS entity_tombstones;

DROP INDEX IF EXISTS idx_friendships_friend_status;
DROP INDEX IF EXISTS idx_friendships_user_status;
DROP INDEX IF EXISTS idx_users_cursor;
DROP INDEX IF EXISTS idx_notifications_user_cursor;
DROP INDEX IF EXISTS idx_activity_visibility_user_activity;
DROP INDEX IF EXISTS idx_activities_cursor;
DROP INDEX IF EXISTS idx_activities_group_cursor;
DROP INDEX IF EXISTS idx_expenses_cursor;
DROP INDEX IF EXISTS idx_expenses_personal_cursor;
DROP INDEX IF EXISTS idx_expenses_group_cursor;
DROP INDEX IF EXISTS idx_groups_cursor;
