package group

import (
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/group/data"
	"github.com/Saurrabhh/splittr_be/internal/group/domain"
	"github.com/Saurrabhh/splittr_be/internal/group/presentation/http"
)

// Domain Type Aliases
type (
	Group                    = domain.Group
	Member                   = domain.Member
	MemberRole               = domain.MemberRole
	MemberStatus             = domain.MemberStatus
	Preview                  = domain.Preview
	DecideJoinRequestPayload = domain.DecideJoinRequestPayload
	DetailsResponse          = domain.DetailsResponse
	GroupWithMembers         = domain.GroupWithMembers
	JoinResponse             = domain.JoinResponse
	Repository               = domain.Repository
	ActivityLogger           = domain.ActivityLogger
	NotificationSender       = domain.NotificationSender
	UseCase                  = domain.UseCase
)

// DBRepository Data Type Aliases
type DBRepository = data.DBRepository

// Handler Presentation Type Aliases
type Handler = http.Handler

// Constants
const (
	MemberStatusActive   = domain.MemberStatusActive
	MemberStatusPending  = domain.MemberStatusPending
	MemberStatusRejected = domain.MemberStatusRejected
)

const (
	MemberRoleAdmin  = domain.MemberRoleAdmin
	MemberRoleMember = domain.MemberRoleMember
)

// NewRepository Constructors
func NewRepository(database *db.DB, tm *db.TransactionManager) *DBRepository {
	return data.NewRepository(database, tm)
}

func NewUseCase(
	repo Repository,
	tx db.Transactor,
	activitySvc ActivityLogger,
	notificationSvc NotificationSender,
) *UseCase {
	return domain.NewUseCase(repo, tx, activitySvc, notificationSvc)
}

func NewHandler(uc *UseCase) *Handler {
	return http.NewHandler(uc)
}
