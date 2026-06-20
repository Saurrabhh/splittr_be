package activity

import (
	"time"
)

// Activity represents an audit log entry for actions performed in the system.
type Activity struct {
	ID          string    `json:"id"`
	GroupID     *string   `json:"groupId,omitempty"`
	ActorID     *string   `json:"actorId,omitempty"`
	ActorName   *string   `json:"actorName,omitempty"`
	ActionType  string    `json:"actionType"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}
