package domain

// AlertType represents the machine-readable category of a notification alert.
type AlertType string // @name AlertType

const (
	AlertTypeExpenseAdded        AlertType = "EXPENSE_ADDED"
	AlertTypePaymentReceived     AlertType = "PAYMENT_RECEIVED"
	AlertTypeJoinRequestPending  AlertType = "JOIN_REQUEST_PENDING"
	AlertTypeJoinRequestApproved AlertType = "JOIN_REQUEST_APPROVED"
	AlertTypeJoinRequestRejected AlertType = "JOIN_REQUEST_REJECTED"
)
