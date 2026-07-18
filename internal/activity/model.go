package activity

import (
	"encoding/json"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
)

// Activity represents an audit log entry for actions performed in the system.
type Activity struct {
	ID          string          `json:"id"`
	GroupID     *string         `json:"groupId,omitempty"`
	ActorID     *string         `json:"actorId,omitempty"`
	ActorName   *string         `json:"actorName,omitempty"`
	ActionType  string          `json:"actionType"`
	Description string          `json:"description"`
	EntityType  string          `json:"entityType"`
	EntityID    *string         `json:"entityId,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
}

type FeedItemResponse struct {
	ID          string          `json:"id"`
	EntityType  string          `json:"entityType"`
	EntityID    *string         `json:"entityId,omitempty"`
	ActionType  string          `json:"actionType"`
	Actor       ActorInfo       `json:"actor"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"createdAt"`
	Payload     json.RawMessage `json:"payload,omitempty"`
}

type ActorInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FeedResponse re-uses the shared pagination envelope.
// No domain-specific pagination struct needed.
type FeedResponse = pagination.Response[FeedItemResponse]
