-- +goose Up
ALTER TABLE notifications ADD COLUMN type VARCHAR(50) NOT NULL DEFAULT 'EXPENSE_ADDED';
ALTER TABLE notifications ALTER COLUMN type DROP DEFAULT;

-- +goose Down
ALTER TABLE notifications DROP COLUMN type;
