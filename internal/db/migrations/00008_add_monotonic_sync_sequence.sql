-- +goose Up
CREATE SEQUENCE IF NOT EXISTS global_sync_seq START 1;

ALTER TABLE expenses ADD COLUMN IF NOT EXISTS sync_version BIGINT NOT NULL DEFAULT nextval('global_sync_seq');
ALTER TABLE groups ADD COLUMN IF NOT EXISTS sync_version BIGINT NOT NULL DEFAULT nextval('global_sync_seq');
ALTER TABLE friendships ADD COLUMN IF NOT EXISTS sync_version BIGINT NOT NULL DEFAULT nextval('global_sync_seq');

CREATE INDEX IF NOT EXISTS idx_expenses_sync_version ON expenses(sync_version);
CREATE INDEX IF NOT EXISTS idx_groups_sync_version ON groups(sync_version);
CREATE INDEX IF NOT EXISTS idx_friendships_sync_version ON friendships(sync_version);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_sync_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.sync_version = nextval('global_sync_seq');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


DROP TRIGGER IF EXISTS trg_expenses_sync_version ON expenses;
CREATE TRIGGER trg_expenses_sync_version BEFORE UPDATE ON expenses
FOR EACH ROW EXECUTE FUNCTION update_sync_version();

DROP TRIGGER IF EXISTS trg_groups_sync_version ON groups;
CREATE TRIGGER trg_groups_sync_version BEFORE UPDATE ON groups
FOR EACH ROW EXECUTE FUNCTION update_sync_version();

DROP TRIGGER IF EXISTS trg_friendships_sync_version ON friendships;
CREATE TRIGGER trg_friendships_sync_version BEFORE UPDATE ON friendships
FOR EACH ROW EXECUTE FUNCTION update_sync_version();

-- +goose Down
DROP TRIGGER IF EXISTS trg_friendships_sync_version ON friendships;
DROP TRIGGER IF EXISTS trg_groups_sync_version ON groups;
DROP TRIGGER IF EXISTS trg_expenses_sync_version ON expenses;
DROP FUNCTION IF EXISTS update_sync_version();

DROP INDEX IF EXISTS idx_friendships_sync_version;
DROP INDEX IF EXISTS idx_groups_sync_version;
DROP INDEX IF EXISTS idx_expenses_sync_version;

ALTER TABLE friendships DROP COLUMN IF EXISTS sync_version;
ALTER TABLE groups DROP COLUMN IF EXISTS sync_version;
ALTER TABLE expenses DROP COLUMN IF EXISTS sync_version;

DROP SEQUENCE IF EXISTS global_sync_seq;
