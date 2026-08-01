package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/Saurrabhh/splittr_be/internal/expense/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// DBRepository implements database operations for expenses.
type DBRepository struct {
	db *db.DB
	tm *db.TransactionManager
}

// NewRepository creates a new DBRepository.
func NewRepository(database *db.DB, tm *db.TransactionManager) *DBRepository {
	return &DBRepository{
		db: database,
		tm: tm,
	}
}

// CreateExpense inserts an expense record.
func (r *DBRepository) CreateExpense(ctx context.Context, e *domain.Expense) error {
	parsedID, err := uuid.Parse(e.ID)
	if err != nil {
		return fmt.Errorf("invalid expense uuid: %w", err)
	}

	parsedPaidBy, err := uuid.Parse(e.PaidBy)
	if err != nil {
		return fmt.Errorf("invalid paidBy uuid: %w", err)
	}

	parsedCreatedBy, err := uuid.Parse(e.CreatedBy)
	if err != nil {
		return fmt.Errorf("invalid createdBy uuid: %w", err)
	}

	var pgGroupID pgtype.UUID
	if e.GroupID != nil && *e.GroupID != "" {
		gUUID, err := uuid.Parse(*e.GroupID)
		if err != nil {
			return fmt.Errorf("invalid group uuid: %w", err)
		}
		pgGroupID = pgtype.UUID{Bytes: gUUID, Valid: true}
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbExpense, err := q.CreateExpense(ctx, dbgen.CreateExpenseParams{
		ID:          parsedID,
		Description: e.Description,
		Amount:      floatToNumeric(e.Amount),
		Currency:    e.Currency,
		GroupID:     pgGroupID,
		PaidBy:      parsedPaidBy,
		CreatedBy:   parsedCreatedBy,
		IsPayment:   e.IsPayment,
		SpentAt:     pgtype.Timestamptz{Time: e.SpentAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("insert expense: %w", err)
	}

	e.CreatedAt = dbExpense.CreatedAt.Time
	e.UpdatedAt = dbExpense.UpdatedAt.Time
	return nil
}

// CreateExpenseSplit inserts a split share.
func (r *DBRepository) CreateExpenseSplit(ctx context.Context, s *domain.Split) error {
	parsedExpenseID, err := uuid.Parse(s.ExpenseID)
	if err != nil {
		return fmt.Errorf("invalid expense uuid: %w", err)
	}

	parsedUserID, err := uuid.Parse(s.UserID)
	if err != nil {
		return fmt.Errorf("invalid user uuid: %w", err)
	}

	var splitVal pgtype.Numeric
	if s.SplitValue != nil {
		splitVal = floatToNumeric(*s.SplitValue)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	err = q.CreateExpenseSplit(ctx, dbgen.CreateExpenseSplitParams{
		ExpenseID:  parsedExpenseID,
		UserID:     parsedUserID,
		Amount:     floatToNumeric(s.Amount),
		SplitType:  string(s.SplitType),
		SplitValue: splitVal,
	})
	if err != nil {
		return fmt.Errorf("insert expense split: %w", err)
	}

	return nil
}

// GetExpenseByID retrieves an expense by its ID.
func (r *DBRepository) GetExpenseByID(ctx context.Context, id string) (*domain.Expense, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	dbExpense, err := q.GetExpenseByID(ctx, parsedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query expense: %w", err)
	}

	return toDomainExpense(dbExpense), nil
}

// ListExpenseSplits lists all splits of a specific expense.
func (r *DBRepository) ListExpenseSplits(ctx context.Context, expenseID string) ([]domain.Split, error) {
	parsedExpenseID, err := uuid.Parse(expenseID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListExpenseSplits(ctx, parsedExpenseID)
	if err != nil {
		return nil, fmt.Errorf("query splits: %w", err)
	}

	splits := make([]domain.Split, 0, len(rows))
	for _, row := range rows {
		var splitVal *float64
		if row.SplitValue.Valid {
			splitVal = new(numericToFloat(row.SplitValue))
		}

		splits = append(splits, domain.Split{
			ExpenseID:  row.ExpenseID.String(),
			UserID:     row.UserID.String(),
			Amount:     numericToFloat(row.Amount),
			SplitType:  domain.SplitType(row.SplitType),
			SplitValue: splitVal,
			Name:       row.Name,
			Email:      textToPtr(row.Email),
			Phone:      textToPtr(row.Phone),
		})
	}

	return splits, nil
}

// ListExpenseSplitsByIDs fetches all splits for a batch of expense IDs in a single query.
// This is the optimised alternative to calling ListExpenseSplits N times.
func (r *DBRepository) ListExpenseSplitsByIDs(ctx context.Context, expenseIDs []string) ([]domain.Split, error) {
	if len(expenseIDs) == 0 {
		return nil, nil
	}

	uuids := make([]uuid.UUID, 0, len(expenseIDs))
	for _, id := range expenseIDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid expense uuid %q: %w", id, err)
		}
		uuids = append(uuids, parsed)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListExpenseSplitsByIDs(ctx, uuids)
	if err != nil {
		return nil, fmt.Errorf("query splits by ids: %w", err)
	}

	splits := make([]domain.Split, 0, len(rows))
	for _, row := range rows {
		var splitVal *float64
		if row.SplitValue.Valid {
			splitVal = new(numericToFloat(row.SplitValue))
		}
		splits = append(splits, domain.Split{
			ExpenseID:  row.ExpenseID.String(),
			UserID:     row.UserID.String(),
			Amount:     numericToFloat(row.Amount),
			SplitType:  domain.SplitType(row.SplitType),
			SplitValue: splitVal,
			Name:       row.Name,
			Email:      textToPtr(row.Email),
			Phone:      textToPtr(row.Phone),
		})
	}
	return splits, nil
}

// parseCursorArgs converts optional cursor fields to pg types for paginated queries.
func parseCursorArgs(lastTime *time.Time, lastID *string) (pgtype.Timestamptz, uuid.UUID) {
	var pgLastTime pgtype.Timestamptz
	if lastTime != nil {
		pgLastTime = pgtype.Timestamptz{Time: *lastTime, Valid: true}
	}
	var lastIDUUID uuid.UUID
	if lastID != nil {
		if parsed, err := uuid.Parse(*lastID); err == nil {
			lastIDUUID = parsed
		}
	}
	return pgLastTime, lastIDUUID
}

// ListExpensesByGroup lists expenses for a group with cursor-based pagination.
func (r *DBRepository) ListExpensesByGroup(ctx context.Context, groupID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.Expense, error) {
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	pgLastTime, lastIDUUID := parseCursorArgs(lastTime, lastID)
	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListExpensesByGroupPaginated(ctx, dbgen.ListExpensesByGroupPaginatedParams{
		GroupID: pgtype.UUID{Bytes: parsedGroupID, Valid: true},
		Limit:   limit,
		Column3: pgLastTime,
		Column4: lastIDUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list group expenses: %w", err)
	}

	return toExpensesFromGroupPaginated(rows), nil
}

// ListUserPersonalExpenses lists a user's private budgeting expenses with cursor-based pagination.
func (r *DBRepository) ListUserPersonalExpenses(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.Expense, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	pgLastTime, lastIDUUID := parseCursorArgs(lastTime, lastID)
	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListUserPersonalExpensesPaginated(ctx, dbgen.ListUserPersonalExpensesPaginatedParams{
		PaidBy:  parsedUserID,
		Limit:   limit,
		Column3: pgLastTime,
		Column4: lastIDUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list personal expenses: %w", err)
	}

	return toExpensesFromPersonalPaginated(rows), nil
}

// ListUserFriendExpenses lists direct non-group friend splits with cursor-based pagination.
func (r *DBRepository) ListUserFriendExpenses(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.Expense, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	pgLastTime, lastIDUUID := parseCursorArgs(lastTime, lastID)
	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListUserFriendExpensesPaginated(ctx, dbgen.ListUserFriendExpensesPaginatedParams{
		PaidBy:  parsedUserID,
		Limit:   limit,
		Column3: pgLastTime,
		Column4: lastIDUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list friend expenses: %w", err)
	}

	return toExpensesFromFriendPaginated(rows), nil
}

// DeleteExpense soft deletes an expense.
func (r *DBRepository) DeleteExpense(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	if err := q.DeleteExpense(ctx, parsedID); err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}
	return nil
}

// GetGroupBalances returns aggregated balances inside a group.
func (r *DBRepository) GetGroupBalances(ctx context.Context, groupID string) ([]domain.UserBalance, error) {
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.GetGroupBalances(ctx, parsedGroupID)
	if err != nil {
		return nil, fmt.Errorf("query group balances: %w", err)
	}

	balances := make([]domain.UserBalance, 0, len(rows))
	for _, row := range rows {
		balances = append(balances, domain.UserBalance{
			UserID:     row.UserID.String(),
			UserName:   row.UserName,
			NetBalance: numericToFloat(row.NetBalance),
		})
	}
	return balances, nil
}

// GetFriendBalances returns direct friend-to-friend balances.
func (r *DBRepository) GetFriendBalances(ctx context.Context, userID string) ([]domain.UserBalance, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.GetFriendBalances(ctx, parsedUserID)
	if err != nil {
		return nil, fmt.Errorf("query friend balances: %w", err)
	}

	balances := make([]domain.UserBalance, 0, len(rows))
	for _, row := range rows {
		balances = append(balances, domain.UserBalance{
			UserID:     row.FriendID.String(),
			UserName:   row.FriendName,
			NetBalance: numericToFloat(row.NetBalance),
		})
	}
	return balances, nil
}

// GetGroupPairwiseDebts returns direct pairwise splits inside a group.
func (r *DBRepository) GetGroupPairwiseDebts(ctx context.Context, groupID string) ([]domain.PairwiseDebt, error) {
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.GetGroupPairwiseDebts(ctx, pgtype.UUID{Bytes: parsedGroupID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("query pairwise debts: %w", err)
	}

	debts := make([]domain.PairwiseDebt, 0, len(rows))
	for _, row := range rows {
		debts = append(debts, domain.PairwiseDebt{
			CreditorID:   row.CreditorID.String(),
			CreditorName: row.CreditorName,
			DebtorID:     row.DebtorID.String(),
			DebtorName:   row.DebtorName,
			Amount:       numericToFloat(row.TotalAmount),
		})
	}
	return debts, nil
}
