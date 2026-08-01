-- +goose Up
CREATE TYPE member_role AS ENUM ('ADMIN', 'MEMBER');
CREATE TYPE member_status AS ENUM ('ACTIVE', 'PENDING', 'REJECTED');

ALTER TABLE group_members ADD COLUMN role_new member_role;
ALTER TABLE group_members ALTER COLUMN role DROP DEFAULT;

UPDATE group_members SET role_new = UPPER(role)::member_role;

ALTER TABLE group_members ALTER COLUMN role_new SET NOT NULL;

ALTER TABLE group_members DROP COLUMN role;
ALTER TABLE group_members RENAME COLUMN role_new TO role;

ALTER TABLE group_members ALTER COLUMN role SET DEFAULT 'MEMBER';

ALTER TABLE group_members ADD COLUMN status_new member_status;
ALTER TABLE group_members ALTER COLUMN status DROP DEFAULT;

UPDATE group_members SET status_new = status::member_status;

ALTER TABLE group_members ALTER COLUMN status_new SET NOT NULL;

ALTER TABLE group_members DROP COLUMN status;
ALTER TABLE group_members RENAME COLUMN status_new TO status;

ALTER TABLE group_members ALTER COLUMN status SET DEFAULT 'ACTIVE';

-- +goose Down
ALTER TABLE group_members ALTER COLUMN role DROP DEFAULT;
ALTER TABLE group_members ADD COLUMN role_old VARCHAR(50);

UPDATE group_members SET role_old = LOWER(role::text);

ALTER TABLE group_members ALTER COLUMN role_old SET NOT NULL;

ALTER TABLE group_members DROP COLUMN role;
ALTER TABLE group_members RENAME COLUMN role_old TO role;

ALTER TABLE group_members ALTER COLUMN role SET DEFAULT 'member';

ALTER TABLE group_members ALTER COLUMN status DROP DEFAULT;
ALTER TABLE group_members ADD COLUMN status_old VARCHAR(20);

UPDATE group_members SET status_old = status::text;

ALTER TABLE group_members ALTER COLUMN status_old SET NOT NULL;

ALTER TABLE group_members DROP COLUMN status;
ALTER TABLE group_members RENAME COLUMN status_old TO status;

ALTER TABLE group_members ALTER COLUMN status SET DEFAULT 'ACTIVE';

DROP TYPE IF EXISTS member_status;
DROP TYPE IF EXISTS member_role;
