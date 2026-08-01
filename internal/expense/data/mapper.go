package data

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/Saurrabhh/splittr_be/internal/expense/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func mapExpenseRow(
	id uuid.UUID,
	description string,
	amount pgtype.Numeric,
	currency string,
	category string,
	groupID pgtype.UUID,
	paidBy uuid.UUID,
	createdBy uuid.UUID,
	isPayment bool,
	spentAt pgtype.Timestamptz,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
) domain.Expense {
	var groupIDStr *string
	if groupID.Valid {
		s := uuid.UUID(groupID.Bytes).String()
		groupIDStr = &s
	}
	return domain.Expense{
		ID:          id.String(),
		Description: description,
		Amount:      numericToFloat(amount),
		Currency:    currency,
		Category:    category,
		GroupID:     groupIDStr,
		PaidBy:      paidBy.String(),
		CreatedBy:   createdBy.String(),
		IsPayment:   isPayment,
		SpentAt:     spentAt.Time,
		CreatedAt:   createdAt.Time,
		UpdatedAt:   updatedAt.Time,
	}
}

func toExpensesFromGroupPaginated(rows []dbgen.ListExpensesByGroupPaginatedRow) []domain.Expense {
	expenses := make([]domain.Expense, 0, len(rows))
	for _, row := range rows {
		expenses = append(expenses, mapExpenseRow(
			row.ID, row.Description, row.Amount, row.Currency, row.Category,
			row.GroupID, row.PaidBy, row.CreatedBy, row.IsPayment,
			row.SpentAt, row.CreatedAt, row.UpdatedAt,
		))
	}
	return expenses
}

func toExpensesFromPersonalPaginated(rows []dbgen.ListUserPersonalExpensesPaginatedRow) []domain.Expense {
	expenses := make([]domain.Expense, 0, len(rows))
	for _, row := range rows {
		expenses = append(expenses, mapExpenseRow(
			row.ID, row.Description, row.Amount, row.Currency, row.Category,
			row.GroupID, row.PaidBy, row.CreatedBy, row.IsPayment,
			row.SpentAt, row.CreatedAt, row.UpdatedAt,
		))
	}
	return expenses
}

func toExpensesFromFriendPaginated(rows []dbgen.ListUserFriendExpensesPaginatedRow) []domain.Expense {
	expenses := make([]domain.Expense, 0, len(rows))
	for _, row := range rows {
		expenses = append(expenses, mapExpenseRow(
			row.ID, row.Description, row.Amount, row.Currency, row.Category,
			row.GroupID, row.PaidBy, row.CreatedBy, row.IsPayment,
			row.SpentAt, row.CreatedAt, row.UpdatedAt,
		))
	}
	return expenses
}

func toDomainExpense(dbg dbgen.Expense) *domain.Expense {
	var groupIDStr *string
	if dbg.GroupID.Valid {
		s := uuid.UUID(dbg.GroupID.Bytes).String()
		groupIDStr = &s
	}

	var deletedAtTime *time.Time
	if dbg.DeletedAt.Valid {
		deletedAtTime = &dbg.DeletedAt.Time
	}

	return &domain.Expense{
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
