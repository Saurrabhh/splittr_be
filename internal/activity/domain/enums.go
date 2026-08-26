package domain

// EntityType represents the domain category of the payload discriminator.
// @enums EXPENSE SETTLEMENT MEMBER GROUP
type EntityType string // @name Activity.EntityType

const (
	EntityTypeExpense    EntityType = "EXPENSE"
	EntityTypeSettlement EntityType = "SETTLEMENT"
	EntityTypeMember     EntityType = "MEMBER"
	EntityTypeGroup      EntityType = "GROUP"
)

// ActionType represents the specific action performed.
// @enums EXPENSE_CREATED SETTLEMENT MEMBER_ADDED MEMBER_LEFT MEMBER_KICKED MEMBER_ROLE_UPDATED GROUP_CREATED GROUP_UPDATED GROUP_ARCHIVED MEMBER_JOINED
type ActionType string // @name Activity.ActionType

const (
	ActionTypeExpenseCreated    ActionType = "EXPENSE_CREATED"
	ActionTypeSettlementCreated ActionType = "SETTLEMENT"
	ActionTypeMemberAdded       ActionType = "MEMBER_ADDED"
	ActionTypeMemberLeft        ActionType = "MEMBER_LEFT"
	ActionTypeMemberKicked      ActionType = "MEMBER_KICKED"
	ActionTypeMemberRoleUpdated ActionType = "MEMBER_ROLE_UPDATED"
	ActionTypeGroupCreated      ActionType = "GROUP_CREATED"
	ActionTypeGroupUpdated      ActionType = "GROUP_UPDATED"
	ActionTypeGroupArchived     ActionType = "GROUP_ARCHIVED"
	ActionTypeMemberJoined      ActionType = "MEMBER_JOINED"
)
