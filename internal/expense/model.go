package expense

import (
	"time"
)

// SplitType represents the method used to split an expense.
type SplitType string

const (
	SplitTypeEqual      SplitType = "EQUAL"
	SplitTypeExact      SplitType = "EXACT"
	SplitTypePercentage SplitType = "PERCENTAGE"
)

// Expense represents an expense record in the system.
type Expense struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Amount      float64    `json:"amount"`
	Currency    string     `json:"currency"`
	Category    string     `json:"category"`
	GroupID     *string    `json:"groupId,omitempty"`
	PaidBy      string     `json:"paidBy"`
	CreatedBy   string     `json:"createdBy"`
	IsPayment   bool       `json:"isPayment"`
	SpentAt     time.Time  `json:"spentAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

// ExpenseSplit represents an individual user's split share in an expense.
type ExpenseSplit struct {
	ExpenseID  string    `json:"expenseId"`
	UserID     string    `json:"userId"`
	Amount     float64   `json:"amount"`
	SplitType  SplitType `json:"splitType"`
	SplitValue *float64  `json:"splitValue,omitempty"`
	Name       string    `json:"name"`
	Email      *string   `json:"email,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
}

// InputSplit is used for parsing incoming splits in create/update requests.
type InputSplit struct {
	UserID     string   `json:"userId"`
	Amount     *float64 `json:"amount,omitempty"`     // Required if splitType is EXACT
	Percentage *float64 `json:"percentage,omitempty"` // Required if splitType is PERCENTAGE
}

// UserBalance represents the net balance of a user in a group or direct relation.
type UserBalance struct {
	UserID     string  `json:"userId"`
	UserName   string  `json:"userName"`
	NetBalance float64 `json:"netBalance"`
}

// Settlement represents a recommended transaction to resolve debts between two users.
type Settlement struct {
	FromUserID   string  `json:"fromUserId"`
	FromUserName string  `json:"fromUserName"`
	ToUserID     string  `json:"toUserId"`
	ToUserName   string  `json:"toUserName"`
	Amount       float64 `json:"amount"`
}

// BalanceResponse contains a list of member balances and a list of recommended settlement transactions.
type BalanceResponse struct {
	Balances    []UserBalance `json:"balances"`
	Settlements []Settlement  `json:"settlements"`
}

// PairwiseDebt represents a direct debt between two users inside a group before netting off.
type PairwiseDebt struct {
	CreditorID   string  `json:"creditorId"`
	CreditorName string  `json:"creditorName"`
	DebtorID     string  `json:"debtorId"`
	DebtorName   string  `json:"debtorName"`
	Amount       float64 `json:"amount"`
}

// CreateExpenseResponse represents the response returned after creating an expense.
type CreateExpenseResponse struct {
	Expense *Expense       `json:"expense"`
	Splits  []ExpenseSplit `json:"splits"`
}

// SettleExpenseResponse represents the response returned after settling a balance.
type SettleExpenseResponse struct {
	Expense *Expense      `json:"expense"`
	Split   *ExpenseSplit `json:"split"`
}

// GetExpenseDetailsResponse represents the response containing an expense and its splits details.
type GetExpenseDetailsResponse struct {
	Expense *Expense       `json:"expense"`
	Splits  []ExpenseSplit `json:"splits"`
}
