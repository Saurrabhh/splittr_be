package domain

import "time"

// SplitType represents the method used to split an expense.
// @enums EQUAL EXACT PERCENTAGE
type SplitType string // @name Expense.SplitType

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
	SyncVersion int64      `json:"syncVersion"`
} // @name Expense.Expense


// Split represents an individual user's split share in an expense.
type Split struct {
	ExpenseID  string    `json:"expenseId"`
	UserID     string    `json:"userId"`
	Amount     float64   `json:"amount"`
	SplitType  SplitType `json:"splitType"`
	SplitValue *float64  `json:"splitValue,omitempty"`
	Name       string    `json:"name"`
	Email      *string   `json:"email,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
} // @name Expense.Split

// InputSplit is used for parsing incoming splits in create/update requests.
type InputSplit struct {
	UserID     string   `json:"userId"`
	Amount     *float64 `json:"amount,omitempty"`     // Required if splitType is EXACT
	Percentage *float64 `json:"percentage,omitempty"` // Required if splitType is PERCENTAGE
} // @name Expense.InputSplit

// UserBalance represents the net balance of a user in a group or direct relation.
type UserBalance struct {
	UserID     string  `json:"userId"`
	UserName   string  `json:"userName"`
	NetBalance float64 `json:"netBalance"`
} // @name Expense.UserBalance

// Settlement represents a recommended transaction to resolve debts between two users.
type Settlement struct {
	FromUserID   string  `json:"fromUserId"`
	FromUserName string  `json:"fromUserName"`
	ToUserID     string  `json:"toUserId"`
	ToUserName   string  `json:"toUserName"`
	Amount       float64 `json:"amount"`
} // @name Expense.Settlement

// PairwiseDebt represents a direct debt between two users inside a group before netting off.
type PairwiseDebt struct {
	CreditorID   string  `json:"creditorId"`
	CreditorName string  `json:"creditorName"`
	DebtorID     string  `json:"debtorId"`
	DebtorName   string  `json:"debtorName"`
	Amount       float64 `json:"amount"`
} // @name Expense.PairwiseDebt

// ExpenseWithSplits is the unified response shape for create, get-by-ID, and list operations.
type ExpenseWithSplits struct {
	Expense
	Splits []Split `json:"splits"`
} // @name Expense.ExpenseWithSplits
