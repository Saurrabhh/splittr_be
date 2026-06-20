-- name: CreateExpense :one
INSERT INTO expenses (id, description, amount, currency, category, group_id, paid_by, created_by, is_payment, spent_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
RETURNING id, description, amount, currency, category, group_id, paid_by, created_by, is_payment, spent_at, created_at, updated_at, deleted_at;

-- name: CreateExpenseSplit :exec
INSERT INTO expense_splits (expense_id, user_id, amount, split_type, split_value)
VALUES ($1, $2, $3, $4, $5);

-- name: GetExpenseByID :one
SELECT id, description, amount, currency, category, group_id, paid_by, created_by, is_payment, spent_at, created_at, updated_at, deleted_at
FROM expenses
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListExpenseSplits :many
SELECT expense_id, user_id, amount, split_type, split_value, u.name, u.email, u.phone
FROM expense_splits es
JOIN users u ON es.user_id = u.id
WHERE es.expense_id = $1;

-- name: ListExpensesByGroup :many
SELECT id, description, amount, currency, category, group_id, paid_by, created_by, is_payment, spent_at, created_at, updated_at
FROM expenses
WHERE group_id = $1 AND deleted_at IS NULL
ORDER BY spent_at DESC;

-- name: ListUserPersonalExpenses :many
-- Retrieve expenses logged purely for personal budgeting (paid by user, and user is the ONLY split member)
SELECT e.id, e.description, e.amount, e.currency, e.category, e.group_id, e.paid_by, e.created_by, e.is_payment, e.spent_at, e.created_at, e.updated_at
FROM expenses e
WHERE e.paid_by = $1 
  AND e.group_id IS NULL 
  AND e.deleted_at IS NULL
  AND (
      SELECT COUNT(*) 
      FROM expense_splits es 
      WHERE es.expense_id = e.id
  ) = 1
ORDER BY e.spent_at DESC;

-- name: ListUserFriendExpenses :many
-- Retrieve direct (non-group) splits between current user and any other user
SELECT DISTINCT e.id, e.description, e.amount, e.currency, e.category, e.group_id, e.paid_by, e.created_by, e.is_payment, e.spent_at, e.created_at, e.updated_at
FROM expenses e
JOIN expense_splits es ON e.id = es.expense_id
WHERE e.group_id IS NULL 
  AND e.deleted_at IS NULL
  AND (e.paid_by = $1 OR es.user_id = $1)
  AND (
      SELECT COUNT(*) 
      FROM expense_splits es2 
      WHERE es2.expense_id = e.id
  ) > 1
ORDER BY e.spent_at DESC;

-- name: DeleteExpense :exec
UPDATE expenses
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: GetGroupBalances :many
SELECT 
    u.id AS user_id,
    u.name AS user_name,
    CAST(COALESCE(paid.total_paid, 0.00) - COALESCE(owed.total_owed, 0.00) AS NUMERIC) AS net_balance
FROM users u
JOIN group_members gm ON u.id = gm.user_id AND gm.group_id = $1
LEFT JOIN (
    SELECT paid_by, SUM(amount) AS total_paid
    FROM expenses
    WHERE group_id = $1 AND deleted_at IS NULL
    GROUP BY paid_by
) paid ON u.id = paid.paid_by
LEFT JOIN (
    SELECT es.user_id, SUM(es.amount) AS total_owed
    FROM expense_splits es
    JOIN expenses e ON es.expense_id = e.id
    WHERE e.group_id = $1 AND e.deleted_at IS NULL
    GROUP BY es.user_id
) owed ON u.id = owed.user_id;

-- name: GetFriendBalances :many
SELECT 
    friend.id AS friend_id,
    friend.name AS friend_name,
    CAST(COALESCE(owed_to_me.amount, 0.00) - COALESCE(owed_by_me.amount, 0.00) AS NUMERIC) AS net_balance
FROM users friend
LEFT JOIN (
    SELECT es.user_id AS friend_id, SUM(es.amount) AS amount
    FROM expense_splits es
    JOIN expenses e ON es.expense_id = e.id
    WHERE e.paid_by = $1 AND e.group_id IS NULL AND e.deleted_at IS NULL AND es.user_id != $1
    GROUP BY es.user_id
) owed_to_me ON friend.id = owed_to_me.friend_id
LEFT JOIN (
    SELECT e.paid_by AS friend_id, SUM(es.amount) AS amount
    FROM expense_splits es
    JOIN expenses e ON es.expense_id = e.id
    WHERE e.paid_by != $1 AND e.group_id IS NULL AND e.deleted_at IS NULL AND es.user_id = $1
    GROUP BY e.paid_by
) owed_by_me ON friend.id = owed_by_me.friend_id
WHERE friend.id != $1 AND (owed_to_me.amount IS NOT NULL OR owed_by_me.amount IS NOT NULL);

-- name: GetGroupPairwiseDebts :many
SELECT 
    e.paid_by AS creditor_id,
    c.name AS creditor_name,
    es.user_id AS debtor_id,
    d.name AS debtor_name,
    CAST(SUM(es.amount) AS NUMERIC) AS total_amount
FROM expense_splits es
JOIN expenses e ON es.expense_id = e.id
JOIN users c ON e.paid_by = c.id
JOIN users d ON es.user_id = d.id
WHERE e.group_id = $1 AND e.deleted_at IS NULL AND es.user_id != e.paid_by
GROUP BY e.paid_by, c.name, es.user_id, d.name;
