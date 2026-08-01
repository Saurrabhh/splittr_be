package http

import (
	"github.com/Saurrabhh/splittr_be/internal/expense/domain"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
)

// CreateExpenseResponse represents the response returned after creating an expense.
type CreateExpenseResponse struct {
	Expense *domain.Expense `json:"expense"`
	Splits  []domain.Split  `json:"splits"`
}

// SettleExpenseResponse represents the response returned after settling a balance.
type SettleExpenseResponse struct {
	Expense *domain.Expense `json:"expense"`
	Split   *domain.Split   `json:"split"`
}

// GetExpenseDetailsResponse represents the response containing an expense and its splits details.
type GetExpenseDetailsResponse struct {
	Expense *domain.Expense `json:"expense"`
	Splits  []domain.Split  `json:"splits"`
}

type CreateExpenseRequest struct {
	Description string              `json:"description"`
	Amount      float64             `json:"amount"`
	Currency    string              `json:"currency"`
	Category    *string             `json:"category,omitempty"`
	GroupID     *string             `json:"groupId"`
	PaidBy      string              `json:"paidBy"`
	SplitType   domain.SplitType    `json:"splitType"`
	Splits      []domain.InputSplit `json:"splits"`
}

type SettleExpenseRequest struct {
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	GroupID    *string `json:"groupId"`
	PaidBy     string  `json:"paidBy"`
	ReceivedBy string  `json:"receivedBy"`
}

// ListExpensesResponse represents the paginated expenses list response.
type ListExpensesResponse struct {
	Data       []domain.Expense `json:"data"`
	Pagination pagination.Meta  `json:"pagination"`
}
