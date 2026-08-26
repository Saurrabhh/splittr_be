package domain

import "fmt"

// Event defines the interface for type-safe activity creation.
// Implementations are unexported to enforce factory constructors.
type Event interface {
	ActionType() ActionType
	EntityType() EntityType
	EntityID() string
	Description() string
	Payload() ActivityPayload
}

type event struct {
	actionType  ActionType
	entityType  EntityType
	entityID    string
	description string
	payload     ActivityPayload
}

func (e event) ActionType() ActionType   { return e.actionType }
func (e event) EntityType() EntityType   { return e.entityType }
func (e event) EntityID() string         { return e.entityID }
func (e event) Description() string      { return e.description }
func (e event) Payload() ActivityPayload { return e.payload }

// Type-safe Factory Constructors

func NewExpenseCreatedEvent(expenseID string, payload ExpensePayload, description string) Event {
	payload.Type = EntityTypeExpense
	return event{
		actionType:  ActionTypeExpenseCreated,
		entityType:  EntityTypeExpense,
		entityID:    expenseID,
		description: description,
		payload:     payload,
	}
}

func NewSettlementCreatedEvent(expenseID string, payload SettlementPayload, description string) Event {
	payload.Type = EntityTypeSettlement
	return event{
		actionType:  ActionTypeSettlementCreated,
		entityType:  EntityTypeSettlement,
		entityID:    expenseID,
		description: description,
		payload:     payload,
	}
}

func NewMemberAddedEvent(memberID string, payload MemberPayload) Event {
	payload.Type = EntityTypeMember
	return event{
		actionType:  ActionTypeMemberAdded,
		entityType:  EntityTypeMember,
		entityID:    memberID,
		description: "added a member to the group",
		payload:     payload,
	}
}

func NewMemberRoleUpdatedEvent(memberID string, newRole string, payload MemberPayload) Event {
	payload.Type = EntityTypeMember
	return event{
		actionType:  ActionTypeMemberRoleUpdated,
		entityType:  EntityTypeMember,
		entityID:    memberID,
		description: fmt.Sprintf("updated member role to %s", newRole),
		payload:     payload,
	}
}

func NewGroupCreatedEvent(groupID string, payload GroupPayload) Event {
	payload.Type = EntityTypeGroup
	return event{
		actionType:  ActionTypeGroupCreated,
		entityType:  EntityTypeGroup,
		entityID:    groupID,
		description: "created the group",
		payload:     payload,
	}
}

func NewGroupUpdatedEvent(groupID string, payload GroupPayload) Event {
	payload.Type = EntityTypeGroup
	return event{
		actionType:  ActionTypeGroupUpdated,
		entityType:  EntityTypeGroup,
		entityID:    groupID,
		description: "updated group details",
		payload:     payload,
	}
}

func NewGroupArchivedEvent(groupID string, payload GroupPayload) Event {
	payload.Type = EntityTypeGroup
	return event{
		actionType:  ActionTypeGroupArchived,
		entityType:  EntityTypeGroup,
		entityID:    groupID,
		description: "archived the group",
		payload:     payload,
	}
}

func NewMemberJoinedEvent(memberID string, payload MemberPayload) Event {
	payload.Type = EntityTypeMember
	return event{
		actionType:  ActionTypeMemberJoined,
		entityType:  EntityTypeMember,
		entityID:    memberID,
		description: "joined the group",
		payload:     payload,
	}
}

func NewMemberLeftEvent(memberID string, payload MemberPayload) Event {
	payload.Type = EntityTypeMember
	return event{
		actionType:  ActionTypeMemberLeft,
		entityType:  EntityTypeMember,
		entityID:    memberID,
		description: "left the group",
		payload:     payload,
	}
}

func NewMemberKickedEvent(memberID string, payload MemberPayload) Event {
	payload.Type = EntityTypeMember
	return event{
		actionType:  ActionTypeMemberKicked,
		entityType:  EntityTypeMember,
		entityID:    memberID,
		description: "removed a member from the group",
		payload:     payload,
	}
}
