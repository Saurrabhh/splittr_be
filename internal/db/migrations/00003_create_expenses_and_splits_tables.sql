-- +goose Up

-- Expenses Table
CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    description VARCHAR(255) NOT NULL,
    amount NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    category VARCHAR(50) NOT NULL DEFAULT 'Other',
    group_id UUID REFERENCES groups(id) ON DELETE SET NULL, -- Nullable for personal/friend splits
    paid_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    is_payment BOOLEAN NOT NULL DEFAULT FALSE,
    spent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_expenses_group_id ON expenses(group_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_paid_by ON expenses(paid_by) WHERE deleted_at IS NULL;

-- Expense Splits Table
CREATE TABLE IF NOT EXISTS expense_splits (
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
    split_type VARCHAR(50) NOT NULL DEFAULT 'EQUAL', -- 'EQUAL', 'EXACT', 'PERCENTAGE'
    split_value NUMERIC(12, 2), -- Stores percentage or custom share values
    PRIMARY KEY (expense_id, user_id)
);

CREATE INDEX idx_expense_splits_user_id ON expense_splits(user_id);

-- +goose Down
DROP TABLE IF EXISTS expense_splits;
DROP TABLE IF EXISTS expenses;
