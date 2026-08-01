package activity

import (
	"github.com/Saurrabhh/splittr_be/internal/activity/data"
	"github.com/Saurrabhh/splittr_be/internal/activity/domain"
	activityhttp "github.com/Saurrabhh/splittr_be/internal/activity/presentation/http"
	"github.com/Saurrabhh/splittr_be/internal/db"
)

// Type aliases for domain entities and value objects
type Activity = domain.Activity
type ActorInfo = domain.ActorInfo
type EntityType = domain.EntityType
type ActionType = domain.ActionType
type ActivityPayload = domain.ActivityPayload
type ExpensePayload = domain.ExpensePayload
type SettlementPayload = domain.SettlementPayload
type MemberPayload = domain.MemberPayload
type GroupPayload = domain.GroupPayload
type Event = domain.Event
type Repository = domain.Repository
type UseCase = domain.UseCase
type Handler = activityhttp.Handler

// Data Type Aliases
type DBRepository = data.DBRepository

// EntityType constants
const (
	EntityTypeExpense    = domain.EntityTypeExpense
	EntityTypeSettlement = domain.EntityTypeSettlement
	EntityTypeMember     = domain.EntityTypeMember
	EntityTypeGroup      = domain.EntityTypeGroup
)

// ActionType constants
const (
	ActionTypeExpenseCreated    = domain.ActionTypeExpenseCreated
	ActionTypeSettlementCreated = domain.ActionTypeSettlementCreated
	ActionTypeMemberAdded       = domain.ActionTypeMemberAdded
	ActionTypeMemberLeft        = domain.ActionTypeMemberLeft
	ActionTypeMemberKicked      = domain.ActionTypeMemberKicked
	ActionTypeMemberRoleUpdated = domain.ActionTypeMemberRoleUpdated
	ActionTypeGroupCreated      = domain.ActionTypeGroupCreated
	ActionTypeGroupUpdated      = domain.ActionTypeGroupUpdated
	ActionTypeGroupArchived     = domain.ActionTypeGroupArchived
	ActionTypeMemberJoined      = domain.ActionTypeMemberJoined
	ActionTypeMemberRemoved     = domain.ActionTypeMemberRemoved
)

// Type-safe Event factory functions
var (
	NewExpenseCreatedEvent    = domain.NewExpenseCreatedEvent
	NewSettlementCreatedEvent = domain.NewSettlementCreatedEvent
	NewMemberAddedEvent       = domain.NewMemberAddedEvent
	NewMemberRoleUpdatedEvent = domain.NewMemberRoleUpdatedEvent
	NewGroupCreatedEvent      = domain.NewGroupCreatedEvent
	NewGroupUpdatedEvent      = domain.NewGroupUpdatedEvent
	NewGroupArchivedEvent     = domain.NewGroupArchivedEvent
	NewMemberJoinedEvent      = domain.NewMemberJoinedEvent
	NewMemberLeftEvent        = domain.NewMemberLeftEvent
	NewMemberKickedEvent      = domain.NewMemberKickedEvent
)

// Factory functions for repository, usecase, and handler initialization
func NewRepository(database *db.DB, tm *db.TransactionManager) *DBRepository {
	return data.NewRepository(database, tm)
}

func NewUseCase(repo domain.Repository) *domain.UseCase {
	return domain.NewUseCase(repo)
}

func NewHandler(uc *domain.UseCase) *activityhttp.Handler {
	return activityhttp.NewHandler(uc)
}
