//go:build integration

package notification_test

import (
	"context"
	"testing"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db_test"
	"github.com/Saurrabhh/splittr_be/internal/notification"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestNotificationRepo(t *testing.T) (*notification.DBRepository, *user.DBRepository, *activity.DBRepository, func()) {
	ctx := context.Background()
	testDB, cleanup, err := db_test.SetupTestDB(ctx)
	require.NoError(t, err)

	tm := db.NewTransactionManager(testDB)
	notifRepo := notification.NewRepository(testDB, tm)
	userRepo := user.NewRepository(testDB, tm)
	actRepo := activity.NewRepository(testDB, tm)

	return notifRepo, userRepo, actRepo, cleanup
}

func createTestUser(t *testing.T, userRepo *user.DBRepository, name string) *user.User {
	ctx := context.Background()
	email := uuid.New().String() + "@example.com"
	u := &user.User{
		ID:          uuid.New().String(),
		FirebaseUID: "fb-" + uuid.New().String(),
		Email:       &email,
		Name:        name,
	}
	err := userRepo.Create(ctx, u)
	require.NoError(t, err)
	return u
}

func createTestActivity(t *testing.T, actRepo *activity.DBRepository, actorID string, desc string) *activity.Activity {
	ctx := context.Background()
	act := &activity.Activity{
		ID:          uuid.New().String(),
		ActorID:     &actorID,
		ActionType:  activity.ActionTypeExpenseCreated,
		Description: desc,
		EntityType:  activity.EntityTypeExpense,
	}
	err := actRepo.CreateActivity(ctx, act)
	require.NoError(t, err)
	return act
}

func TestRepository_CreateNotification_And_InvalidUUIDs(t *testing.T) {
	notifRepo, userRepo, actRepo, cleanup := setupTestNotificationRepo(t)
	defer cleanup()
	ctx := context.Background()

	u1 := createTestUser(t, userRepo, "User One")
	actor := createTestUser(t, userRepo, "Actor User")
	act := createTestActivity(t, actRepo, actor.ID, "Created expense")

	notifID := uuid.New().String()
	n := &notification.Notification{
		ID:         notifID,
		UserID:     u1.ID,
		ActorID:    &actor.ID,
		ActivityID: &act.ID,
		Title:      "New Expense Added",
		Content:    "Actor added dinner expense",
	}

	err := notifRepo.CreateNotification(ctx, n)
	require.NoError(t, err)
	assert.False(t, n.CreatedAt.IsZero())
	assert.False(t, n.IsRead)

	// Test Invalid Notification ID
	invalidNotif := &notification.Notification{
		ID:      "invalid-uuid",
		UserID:  u1.ID,
		Title:   "Test",
		Content: "Test",
	}
	err = notifRepo.CreateNotification(ctx, invalidNotif)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid notification uuid")

	// Test Invalid Recipient User ID
	invalidUserNotif := &notification.Notification{
		ID:      uuid.New().String(),
		UserID:  "invalid-uuid",
		Title:   "Test",
		Content: "Test",
	}
	err = notifRepo.CreateNotification(ctx, invalidUserNotif)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid recipient uuid")

	// Test Invalid Actor ID
	invalidActor := "invalid-uuid"
	invalidActorNotif := &notification.Notification{
		ID:      uuid.New().String(),
		UserID:  u1.ID,
		ActorID: &invalidActor,
		Title:   "Test",
		Content: "Test",
	}
	err = notifRepo.CreateNotification(ctx, invalidActorNotif)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid actor uuid")

	// Test Invalid Activity ID
	invalidActivity := "invalid-uuid"
	invalidActivityNotif := &notification.Notification{
		ID:         uuid.New().String(),
		UserID:     u1.ID,
		ActivityID: &invalidActivity,
		Title:      "Test",
		Content:    "Test",
	}
	err = notifRepo.CreateNotification(ctx, invalidActivityNotif)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid activity uuid")
}

func TestRepository_ListUserNotifications_And_Pagination(t *testing.T) {
	notifRepo, userRepo, _, cleanup := setupTestNotificationRepo(t)
	defer cleanup()
	ctx := context.Background()

	u1 := createTestUser(t, userRepo, "User One")
	u2 := createTestUser(t, userRepo, "User Two")

	n1 := &notification.Notification{ID: uuid.New().String(), UserID: u1.ID, Title: "N1", Content: "C1"}
	n2 := &notification.Notification{ID: uuid.New().String(), UserID: u1.ID, Title: "N2", Content: "C2"}
	n3 := &notification.Notification{ID: uuid.New().String(), UserID: u1.ID, Title: "N3", Content: "C3"}
	nUser2 := &notification.Notification{ID: uuid.New().String(), UserID: u2.ID, Title: "N_Other", Content: "C_Other"}

	require.NoError(t, notifRepo.CreateNotification(ctx, n1))
	require.NoError(t, notifRepo.CreateNotification(ctx, n2))
	require.NoError(t, notifRepo.CreateNotification(ctx, n3))
	require.NoError(t, notifRepo.CreateNotification(ctx, nUser2))

	// Fetch first 2 notifications for User 1
	list1, err := notifRepo.ListUserNotifications(ctx, u1.ID, 2, nil, nil)
	require.NoError(t, err)
	require.Len(t, list1, 2)

	// Ensure User 2's notification is not returned
	for _, notif := range list1 {
		assert.Equal(t, u1.ID, notif.UserID)
	}

	// Fetch 2nd page using cursor from 2nd item of list1
	lastItem := list1[1]
	list2, err := notifRepo.ListUserNotifications(ctx, u1.ID, 2, &lastItem.CreatedAt, &lastItem.ID)
	require.NoError(t, err)
	require.Len(t, list2, 1)

	// Test Invalid User UUID
	_, err = notifRepo.ListUserNotifications(ctx, "invalid-uuid", 10, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid uuid")
}

func TestRepository_MarkNotificationAsRead(t *testing.T) {
	notifRepo, userRepo, _, cleanup := setupTestNotificationRepo(t)
	defer cleanup()
	ctx := context.Background()

	u1 := createTestUser(t, userRepo, "User One")
	u2 := createTestUser(t, userRepo, "User Two")

	n1 := &notification.Notification{ID: uuid.New().String(), UserID: u1.ID, Title: "Alert", Content: "Details"}
	require.NoError(t, notifRepo.CreateNotification(ctx, n1))

	// User Two attempts to mark User One's notification as read (User ownership check)
	err := notifRepo.MarkNotificationAsRead(ctx, n1.ID, u2.ID)
	require.NoError(t, err) // SQL UPDATE query matches 0 rows without error

	// Verify n1 is still unread for User One
	list, err := notifRepo.ListUserNotifications(ctx, u1.ID, 10, nil, nil)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.False(t, list[0].IsRead)

	// User One marks their own notification as read
	err = notifRepo.MarkNotificationAsRead(ctx, n1.ID, u1.ID)
	require.NoError(t, err)

	// Verify n1 is now read
	list, err = notifRepo.ListUserNotifications(ctx, u1.ID, 10, nil, nil)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.True(t, list[0].IsRead)

	// Test Invalid UUID errors
	err = notifRepo.MarkNotificationAsRead(ctx, "invalid-uuid", u1.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid uuid")

	err = notifRepo.MarkNotificationAsRead(ctx, n1.ID, "invalid-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user uuid")
}

func TestRepository_MarkAllNotificationsAsRead(t *testing.T) {
	notifRepo, userRepo, _, cleanup := setupTestNotificationRepo(t)
	defer cleanup()
	ctx := context.Background()

	u1 := createTestUser(t, userRepo, "User One")
	u2 := createTestUser(t, userRepo, "User Two")

	n1 := &notification.Notification{ID: uuid.New().String(), UserID: u1.ID, Title: "Alert 1", Content: "A1"}
	n2 := &notification.Notification{ID: uuid.New().String(), UserID: u1.ID, Title: "Alert 2", Content: "A2"}
	nOther := &notification.Notification{ID: uuid.New().String(), UserID: u2.ID, Title: "Alert Other", Content: "AO"}

	require.NoError(t, notifRepo.CreateNotification(ctx, n1))
	require.NoError(t, notifRepo.CreateNotification(ctx, n2))
	require.NoError(t, notifRepo.CreateNotification(ctx, nOther))

	// Mark all read for User 1
	err := notifRepo.MarkAllNotificationsAsRead(ctx, u1.ID)
	require.NoError(t, err)

	// Verify User 1's notifications are marked read
	list1, err := notifRepo.ListUserNotifications(ctx, u1.ID, 10, nil, nil)
	require.NoError(t, err)
	require.Len(t, list1, 2)
	assert.True(t, list1[0].IsRead)
	assert.True(t, list1[1].IsRead)

	// Verify User 2's notification is still unread
	list2, err := notifRepo.ListUserNotifications(ctx, u2.ID, 10, nil, nil)
	require.NoError(t, err)
	require.Len(t, list2, 1)
	assert.False(t, list2[0].IsRead)

	// Test Invalid User UUID
	err = notifRepo.MarkAllNotificationsAsRead(ctx, "invalid-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user uuid")
}
