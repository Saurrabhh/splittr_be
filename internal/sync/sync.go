package sync

import (
	"github.com/Saurrabhh/splittr_be/internal/sync/domain"
	"github.com/Saurrabhh/splittr_be/internal/sync/presentation/http"
)

type (
	Handler            = http.Handler
	UseCase            = domain.UseCase
	UserSyncService    = domain.UserSyncService
	GroupSyncService   = domain.GroupSyncService
	ExpenseSyncService = domain.ExpenseSyncService
	SyncResponse       = domain.SyncResponse
	SyncParams         = domain.SyncParams
)

func NewHandler(uc *domain.UseCase) *http.Handler {
	return http.NewHandler(uc)
}

func NewUseCase(userSync domain.UserSyncService, groupSync domain.GroupSyncService, expenseSync domain.ExpenseSyncService) *domain.UseCase {
	return domain.NewUseCase(userSync, groupSync, expenseSync)
}
