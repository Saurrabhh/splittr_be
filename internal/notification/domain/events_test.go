package domain_test

import (
	"testing"

	"github.com/Saurrabhh/splittr_be/internal/notification/domain"
	"github.com/stretchr/testify/assert"
)

func TestAlertFactories(t *testing.T) {
	tests := []struct {
		name        string
		alert       domain.Alert
		wantType    domain.AlertType
		wantTitle   string
		wantContent string
	}{
		{
			name:        "expense added",
			alert:       domain.NewExpenseAddedAlert("Dinner", 100, "INR"),
			wantType:    domain.AlertTypeExpenseAdded,
			wantTitle:   "New Expense",
			wantContent: "New expense 'Dinner' of 100.00 INR added",
		},
		{
			name:        "payment received",
			alert:       domain.NewPaymentReceivedAlert(50, "INR"),
			wantType:    domain.AlertTypePaymentReceived,
			wantTitle:   "Payment Received",
			wantContent: "Payment of 50.00 INR received",
		},
		{
			name:        "join request pending",
			alert:       domain.NewJoinRequestPendingAlert("Road Trip"),
			wantType:    domain.AlertTypeJoinRequestPending,
			wantTitle:   "Join Request Pending",
			wantContent: "A user requested to join group Road Trip",
		},
		{
			name:        "join request approved",
			alert:       domain.NewJoinRequestApprovedAlert(),
			wantType:    domain.AlertTypeJoinRequestApproved,
			wantTitle:   "Join Request Approved",
			wantContent: "Your request to join the group was approved.",
		},
		{
			name:        "join request rejected",
			alert:       domain.NewJoinRequestRejectedAlert(),
			wantType:    domain.AlertTypeJoinRequestRejected,
			wantTitle:   "Join Request Rejected",
			wantContent: "Your request to join the group was declined.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.alert.AlertType())
			assert.Equal(t, tt.wantTitle, tt.alert.Title())
			assert.Equal(t, tt.wantContent, tt.alert.Content())
		})
	}
}
