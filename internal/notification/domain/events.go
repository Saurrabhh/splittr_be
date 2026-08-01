package domain

import "fmt"

// Alert defines a typed, machine-readable notification delivered to a user.
// Implementations are unexported to enforce factory constructors.
type Alert interface {
	AlertType() AlertType
	Title() string
	Content() string
}

type alert struct {
	alertType AlertType
	title     string
	content   string
}

func (a alert) AlertType() AlertType { return a.alertType }
func (a alert) Title() string        { return a.title }
func (a alert) Content() string      { return a.content }

// Type-safe Factory Constructors

func NewExpenseAddedAlert(description string, amount float64, currency string) Alert {
	return alert{
		alertType: AlertTypeExpenseAdded,
		title:     "New Expense",
		content:   fmt.Sprintf("New expense '%s' of %.2f %s added", description, amount, currency),
	}
}

func NewPaymentReceivedAlert(amount float64, currency string) Alert {
	return alert{
		alertType: AlertTypePaymentReceived,
		title:     "Payment Received",
		content:   fmt.Sprintf("Payment of %.2f %s received", amount, currency),
	}
}

func NewJoinRequestPendingAlert(groupName string) Alert {
	return alert{
		alertType: AlertTypeJoinRequestPending,
		title:     "Join Request Pending",
		content:   fmt.Sprintf("A user requested to join group %s", groupName),
	}
}

func NewJoinRequestApprovedAlert() Alert {
	return alert{
		alertType: AlertTypeJoinRequestApproved,
		title:     "Join Request Approved",
		content:   "Your request to join the group was approved.",
	}
}

func NewJoinRequestRejectedAlert() Alert {
	return alert{
		alertType: AlertTypeJoinRequestRejected,
		title:     "Join Request Rejected",
		content:   "Your request to join the group was declined.",
	}
}
