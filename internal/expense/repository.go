package expense

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
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
func (r *DBRepository) CreateExpense(ctx context.Context, e *Expense) error {
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
func (r *DBRepository) CreateExpenseSplit(ctx context.Context, s *ExpenseSplit) error {
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
func (r *DBRepository) GetExpenseByID(ctx context.Context, id string) (*Expense, error) {
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
func (r *DBRepository) ListExpenseSplits(ctx context.Context, expenseID string) ([]ExpenseSplit, error) {
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

	splits := make([]ExpenseSplit, 0, len(rows))
	for _, row := range rows {
		var splitVal *float64
		if row.SplitValue.Valid {
			v := numericToFloat(row.SplitValue)
			splitVal = &v
		}

		splits = append(splits, ExpenseSplit{
			ExpenseID:  row.ExpenseID.String(),
			UserID:     row.UserID.String(),
			Amount:     numericToFloat(row.Amount),
			SplitType:  SplitType(row.SplitType),
			SplitValue: splitVal,
			Name:       row.Name,
			Email:      textToPtr(row.Email),
			Phone:      textToPtr(row.Phone),
		})
	}

	return splits, nil
}

// ListExpensesByGroup lists expenses for a group.
func (r *DBRepository) ListExpensesByGroup(ctx context.Context, groupID string) ([]Expense, error) {
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListExpensesByGroup(ctx, pgtype.UUID{Bytes: parsedGroupID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list group expenses: %w", err)
	}

	expenses := make([]Expense, 0, len(rows))
	for _, row := range rows {
		var groupIDStr *string
		if row.GroupID.Valid {
			s := uuid.UUID(row.GroupID.Bytes).String()
			groupIDStr = &s
		}

		expenses = append(expenses, Expense{
			ID:          row.ID.String(),
			Description: row.Description,
			Amount:      numericToFloat(row.Amount),
			Currency:    row.Currency,
			GroupID:     groupIDStr,
			PaidBy:      row.PaidBy.String(),
			CreatedBy:   row.CreatedBy.String(),
			IsPayment:   row.IsPayment,
			SpentAt:     row.SpentAt.Time,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}
	return expenses, nil
}

// ListUserPersonalExpenses lists a user's private budgeting expenses.
func (r *DBRepository) ListUserPersonalExpenses(ctx context.Context, userID string) ([]Expense, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListUserPersonalExpenses(ctx, parsedUserID)
	if err != nil {
		return nil, fmt.Errorf("list personal expenses: %w", err)
	}

	expenses := make([]Expense, 0, len(rows))
	for _, row := range rows {
		var groupIDStr *string
		if row.GroupID.Valid {
			s := uuid.UUID(row.GroupID.Bytes).String()
			groupIDStr = &s
		}

		expenses = append(expenses, Expense{
			ID:          row.ID.String(),
			Description: row.Description,
			Amount:      numericToFloat(row.Amount),
			Currency:    row.Currency,
			GroupID:     groupIDStr,
			PaidBy:      row.PaidBy.String(),
			CreatedBy:   row.CreatedBy.String(),
			IsPayment:   row.IsPayment,
			SpentAt:     row.SpentAt.Time,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}
	return expenses, nil
}

// ListUserFriendExpenses lists direct non-group friend splits.
func (r *DBRepository) ListUserFriendExpenses(ctx context.Context, userID string) ([]Expense, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	rows, err := q.ListUserFriendExpenses(ctx, parsedUserID)
	if err != nil {
		return nil, fmt.Errorf("list friend expenses: %w", err)
	}

	expenses := make([]Expense, 0, len(rows))
	for _, row := range rows {
		var groupIDStr *string
		if row.GroupID.Valid {
			s := uuid.UUID(row.GroupID.Bytes).String()
			groupIDStr = &s
		}

		expenses = append(expenses, Expense{
			ID:          row.ID.String(),
			Description: row.Description,
			Amount:      numericToFloat(row.Amount),
			Currency:    row.Currency,
			GroupID:     groupIDStr,
			PaidBy:      row.PaidBy.String(),
			CreatedBy:   row.CreatedBy.String(),
			IsPayment:   row.IsPayment,
			SpentAt:     row.SpentAt.Time,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}
	return expenses, nil
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
func (r *DBRepository) GetGroupBalances(ctx context.Context, groupID string) ([]UserBalance, error) {
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

	balances := make([]UserBalance, 0, len(rows))
	for _, row := range rows {
		balances = append(balances, UserBalance{
			UserID:     row.UserID.String(),
			UserName:   row.UserName,
			NetBalance: numericToFloat(row.NetBalance),
		})
	}
	return balances, nil
}

// GetFriendBalances returns direct friend-to-friend balances.
func (r *DBRepository) GetFriendBalances(ctx context.Context, userID string) ([]UserBalance, error) {
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

	balances := make([]UserBalance, 0, len(rows))
	for _, row := range rows {
		balances = append(balances, UserBalance{
			UserID:     row.FriendID.String(),
			UserName:   row.FriendName,
			NetBalance: numericToFloat(row.NetBalance),
		})
	}
	return balances, nil
}

// GetGroupPairwiseDebts returns direct pairwise splits inside a group.
func (r *DBRepository) GetGroupPairwiseDebts(ctx context.Context, groupID string) ([]PairwiseDebt, error) {
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

	debts := make([]PairwiseDebt, 0, len(rows))
	for _, row := range rows {
		debts = append(debts, PairwiseDebt{
			CreditorID:   row.CreditorID.String(),
			CreditorName: row.CreditorName,
			DebtorID:     row.DebtorID.String(),
			DebtorName:   row.DebtorName,
			Amount:       numericToFloat(row.TotalAmount),
		})
	}
	return debts, nil
}

// Helpers
func toDomainExpense(dbg dbgen.Expense) *Expense {
	var groupIDStr *string
	if dbg.GroupID.Valid {
		s := uuid.UUID(dbg.GroupID.Bytes).String()
		groupIDStr = &s
	}

	var deletedAtTime *time.Time
	if dbg.DeletedAt.Valid {
		deletedAtTime = &dbg.DeletedAt.Time
	}

	return &Expense{
		ID:          dbg.ID.String(),
		Description: dbg.Description,
		Amount:      numericToFloat(dbg.Amount),
		Currency:    dbg.Currency,
		GroupID:     groupIDStr,
		PaidBy:      dbg.PaidBy.String(),
		CreatedBy:   dbg.CreatedBy.String(),
		IsPayment:   dbg.IsPayment,
		SpentAt:     dbg.SpentAt.Time,
		CreatedAt:   dbg.CreatedAt.Time,
		UpdatedAt:   dbg.UpdatedAt.Time,
		DeletedAt:   deletedAtTime,
	}
}

func floatToNumeric(f float64) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(fmt.Sprintf("%.2f", f))
	return num
}

func numericToFloat(num pgtype.Numeric) float64 {
	if !num.Valid || num.Int == nil {
		return 0.0
	}
	fVal, _ := new(big.Float).SetInt(num.Int).Float64()
	return fVal * math.Pow10(int(num.Exp))
}

func textToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
