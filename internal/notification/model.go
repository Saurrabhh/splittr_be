package notification

import (
	"time"
)

// Notification represents a user-specific alert message in their notification tray.
type Notification struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	ActorID    *string   `json:"actorId,omitempty"`
	ActorName  *string   `json:"actorName,omitempty"`
	ActivityID *string   `json:"activityId,omitempty"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	IsRead     bool      `json:"isRead"`
	CreatedAt  time.Time `json:"createdAt"`
}
