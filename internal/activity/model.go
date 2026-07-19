package activity

import (
	"encoding/json"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
)

// EntityType represents the type of entity involved in an activity.
// @enums SYSTEM EXPENSE SETTLEMENT MEMBER GROUP
type EntityType string

const (
	EntityTypeSystem     EntityType = "SYSTEM"
	EntityTypeExpense    EntityType = "EXPENSE"
	EntityTypeSettlement EntityType = "SETTLEMENT"
	EntityTypeMember     EntityType = "MEMBER"
	EntityTypeGroup      EntityType = "GROUP"
)

// ActionType represents the specific action performed.
// @enums EXPENSE_CREATED SETTLEMENT MEMBER_ADDED MEMBER_LEFT MEMBER_KICKED MEMBER_ROLE_UPDATED GROUP_CREATED GROUP_ARCHIVED MEMBER_JOINED
type ActionType string

const (
	ActionTypeExpenseCreated    ActionType = "EXPENSE_CREATED"
	ActionTypeSettlementCreated ActionType = "SETTLEMENT"
	ActionTypeMemberAdded       ActionType = "MEMBER_ADDED"
	ActionTypeMemberLeft        ActionType = "MEMBER_LEFT"
	ActionTypeMemberKicked      ActionType = "MEMBER_KICKED"
	ActionTypeMemberRoleUpdated ActionType = "MEMBER_ROLE_UPDATED"
	ActionTypeGroupCreated      ActionType = "GROUP_CREATED"
	ActionTypeGroupArchived     ActionType = "GROUP_ARCHIVED"
	ActionTypeMemberJoined      ActionType = "MEMBER_JOINED"
)

// Activity represents an audit log entry for actions performed in the system.
type Activity struct {
	ID          string     `json:"id"`
	GroupID     *string    `json:"groupId,omitempty"`
	ActorID     *string    `json:"actorId,omitempty"`
	ActorName   *string    `json:"actorName,omitempty"`
	ActionType  ActionType `json:"actionType"`
	Description string     `json:"description"`
	EntityType  EntityType `json:"entityType"`
	EntityID    *string    `json:"entityId,omitempty"`
	// Snapshot metadata. Shape matches the entity response: EXPENSE (CreateExpenseResponse), SETTLEMENT (SettleExpenseResponse), MEMBER (Member), GROUP (GroupDetailsResponse)
	Metadata    json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
	CreatedAt   time.Time       `json:"createdAt"`
}

type FeedItemResponse struct {
	ID          string     `json:"id"`
	EntityType  EntityType `json:"entityType"`
	EntityID    *string    `json:"entityId,omitempty"`
	ActionType  ActionType `json:"actionType"`
	Actor       ActorInfo  `json:"actor"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"createdAt"`
	// Verbatim entity snapshot. Shape matches the entity response: EXPENSE (CreateExpenseResponse), SETTLEMENT (SettleExpenseResponse), MEMBER (Member), GROUP (GroupDetailsResponse)
	Payload     json.RawMessage `json:"payload,omitempty" swaggertype:"object"`
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

// ListActivitiesResponse represents the paginated user activities response.
type ListActivitiesResponse struct {
	Data       []Activity      `json:"data"`
	Pagination pagination.Meta `json:"pagination"`
}

