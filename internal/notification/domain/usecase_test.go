package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/notification/domain"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockNotificationRepository struct {
	mock.Mock
}

func (m *mockNotificationRepository) CreateNotification(ctx context.Context, notif *domain.Notification) error {
	return m.Called(ctx, notif).Error(0)
}

func (m *mockNotificationRepository) ListUserNotifications(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]domain.Notification, error) {
	args := m.Called(ctx, userID, limit, lastTime, lastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Notification), args.Error(1)
}

func (m *mockNotificationRepository) MarkNotificationAsRead(ctx context.Context, id, userID string) (bool, error) {
	args := m.Called(ctx, id, userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockNotificationRepository) MarkAllNotificationsAsRead(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

// --- CreateAlert Tests ---

func TestCreateAlert_Success(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	ctx := context.Background()

	actorID := "actor-1"
	activityID := "act-1"
	userID := "user-1"

	mockRepo.On("CreateNotification", ctx, mock.AnythingOfType("*domain.Notification")).Return(nil)

	uc := domain.NewUseCase(mockRepo)
	notif, err := uc.CreateAlert(ctx, userID, &actorID, &activityID, domain.NewExpenseAddedAlert("Dinner", 100, "INR"))
	require.NoError(t, err)
	require.NotNil(t, notif)

	assert.NotEmpty(t, notif.ID)
	assert.Equal(t, userID, notif.UserID)
	assert.Equal(t, &actorID, notif.ActorID)
	assert.Equal(t, &activityID, notif.ActivityID)
	assert.Equal(t, domain.AlertTypeExpenseAdded, notif.Type)
	assert.Equal(t, "New Expense", notif.Title)
	assert.Equal(t, "New expense 'Dinner' of 100.00 INR added", notif.Content)
	mockRepo.AssertExpectations(t)
}

func TestCreateAlert_NilAlert(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	ctx := context.Background()

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.CreateAlert(ctx, "user-1", nil, nil, nil)
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "alert is required")
}

func TestCreateAlert_RepoError(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	ctx := context.Background()

	mockRepo.On("CreateNotification", ctx, mock.AnythingOfType("*domain.Notification")).Return(errors.New("db insert failure"))

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.CreateAlert(ctx, "user-1", nil, nil, domain.NewPaymentReceivedAlert(50, "INR"))
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeInternal, appErr.Type)
	assert.Contains(t, appErr.Message, "failed to create notification")
	mockRepo.AssertExpectations(t)
}

// --- ListNotifications Tests ---

func TestListNotifications_Success(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	ctx := context.Background()

	now := time.Now()
	expectedNotifs := []domain.Notification{
		{ID: "notif-1", UserID: "user-1", Title: "Alert 1", Content: "Content 1", CreatedAt: now},
		{ID: "notif-2", UserID: "user-1", Title: "Alert 2", Content: "Content 2", CreatedAt: now.Add(-time.Minute)},
	}

	mockRepo.On("ListUserNotifications", ctx, "user-1", int32(11), (*time.Time)(nil), (*string)(nil)).Return(expectedNotifs, nil)

	uc := domain.NewUseCase(mockRepo)
	resp, err := uc.ListNotifications(ctx, "user-1", pagination.Params{Limit: 10})
	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "notif-1", resp.Data[0].ID)
	assert.Equal(t, "notif-2", resp.Data[1].ID)
	mockRepo.AssertExpectations(t)
}

func TestListNotifications_RepoError(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	ctx := context.Background()

	mockRepo.On("ListUserNotifications", ctx, "user-1", int32(11), (*time.Time)(nil), (*string)(nil)).Return(nil, errors.New("db read failure"))

	uc := domain.NewUseCase(mockRepo)
	_, err := uc.ListNotifications(ctx, "user-1", pagination.Params{Limit: 10})
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeInternal, appErr.Type)
	assert.Contains(t, appErr.Message, "failed to retrieve notifications")
	mockRepo.AssertExpectations(t)
}

// --- MarkAsRead Tests ---

func TestMarkAsRead_Success(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	ctx := context.Background()

	userID := "user-1"
	notifID := "notif-1"

	// Verifies user ownership check passed to repository
	mockRepo.On("MarkNotificationAsRead", ctx, notifID, userID).Return(true, nil)

	uc := domain.NewUseCase(mockRepo)
	err := uc.MarkAsRead(ctx, notifID, userID)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMarkAsRead_EmptyID(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	ctx := context.Background()

	uc := domain.NewUseCase(mockRepo)
	err := uc.MarkAsRead(ctx, "", "user-1")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeValidation, appErr.Type)
	assert.Contains(t, appErr.Message, "notification id is required")
}

func TestMarkAsRead_RepoError(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	ctx := context.Background()

	mockRepo.On("MarkNotificationAsRead", ctx, "notif-1", "user-1").Return(false, errors.New("db error"))

	uc := domain.NewUseCase(mockRepo)
	err := uc.MarkAsRead(ctx, "notif-1", "user-1")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeInternal, appErr.Type)
	assert.Contains(t, appErr.Message, "failed to mark notification as read")
	mockRepo.AssertExpectations(t)
}

func TestMarkAsRead_NotFound(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	ctx := context.Background()

	mockRepo.On("MarkNotificationAsRead", ctx, "notif-1", "user-1").Return(false, nil)

	uc := domain.NewUseCase(mockRepo)
	err := uc.MarkAsRead(ctx, "notif-1", "user-1")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeNotFound, appErr.Type)
	assert.Contains(t, appErr.Message, response.MsgNotificationNotFound)
	mockRepo.AssertExpectations(t)
}

// --- MarkAllAsRead Tests ---

func TestMarkAllAsRead_Success(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	ctx := context.Background()

	userID := "user-1"
	mockRepo.On("MarkAllNotificationsAsRead", ctx, userID).Return(nil)

	uc := domain.NewUseCase(mockRepo)
	err := uc.MarkAllAsRead(ctx, userID)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMarkAllAsRead_RepoError(t *testing.T) {
	mockRepo := new(mockNotificationRepository)
	ctx := context.Background()

	mockRepo.On("MarkAllNotificationsAsRead", ctx, "user-1").Return(errors.New("db update failure"))

	uc := domain.NewUseCase(mockRepo)
	err := uc.MarkAllAsRead(ctx, "user-1")
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, response.TypeInternal, appErr.Type)
	assert.Contains(t, appErr.Message, "failed to mark all notifications as read")
	mockRepo.AssertExpectations(t)
}
