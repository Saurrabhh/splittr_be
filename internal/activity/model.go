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
	ActionType  string          `json:"actionType" Enums:"EXPENSE_CREATED,SETTLEMENT,MEMBER_ADDED,MEMBER_LEFT,MEMBER_KICKED,MEMBER_ROLE_UPDATED,GROUP_CREATED,GROUP_ARCHIVED,MEMBER_JOINED"`
	Description string          `json:"description"`
	EntityType  string          `json:"entityType" Enums:"EXPENSE,SETTLEMENT,MEMBER,GROUP"`
	EntityID    *string         `json:"entityId,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty" swaggertype:"object" description:"Snapshot metadata. Shape matches the entity response: EXPENSE (CreateExpenseResponse), SETTLEMENT (SettleExpenseResponse), MEMBER (Member), GROUP (GroupDetailsResponse)"`
	CreatedAt   time.Time       `json:"createdAt"`
}

type FeedItemResponse struct {
	ID          string          `json:"id"`
	EntityType  string          `json:"entityType" Enums:"EXPENSE,SETTLEMENT,MEMBER,GROUP"`
	EntityID    *string         `json:"entityId,omitempty"`
	ActionType  string          `json:"actionType" Enums:"EXPENSE_CREATED,SETTLEMENT,MEMBER_ADDED,MEMBER_LEFT,MEMBER_KICKED,MEMBER_ROLE_UPDATED,GROUP_CREATED,GROUP_ARCHIVED,MEMBER_JOINED"`
	Actor       ActorInfo       `json:"actor"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"createdAt"`
	Payload     json.RawMessage `json:"payload,omitempty" swaggertype:"object" description:"Verbatim entity snapshot. Shape matches the entity response: EXPENSE (CreateExpenseResponse), SETTLEMENT (SettleExpenseResponse), MEMBER (Member), GROUP (GroupDetailsResponse)"`
}

type ActorInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FeedResponse represents the paginated group activity feed response.
type FeedResponse struct {
	Data       []FeedItemResponse `json:"data"`
	Pagination pagination.Meta    `json:"pagination"`
}
