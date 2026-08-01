package http

import (
	"github.com/Saurrabhh/splittr_be/internal/expense/domain"
)

// ExpenseResponse is the unified shape for create, get-by-ID, and list items.
// All three endpoints return { expense: {...}, splits: [...] }.
type ExpenseResponse = domain.ExpenseWithSplits

// CreateExpenseRequest is the payload for POST /expenses.
type CreateExpenseRequest struct {
	Description string              `json:"description"`
	Amount      float64             `json:"amount"`
	Currency    string              `json:"currency"`
	Category    *string             `json:"category,omitempty"`
	GroupID     *string             `json:"groupId"`
	PaidBy      string              `json:"paidBy"`
	SplitType   domain.SplitType    `json:"splitType"`
	Splits      []domain.InputSplit `json:"splits"`
} // @name CreateExpenseRequest

// SettleExpenseRequest is the payload for POST /expenses/settle.
type SettleExpenseRequest struct {
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	GroupID    *string `json:"groupId"`
	PaidBy     string  `json:"paidBy"`
	ReceivedBy string  `json:"receivedBy"`
} // @name SettleExpenseRequest
