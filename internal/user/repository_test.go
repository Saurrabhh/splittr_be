//go:build integration

package user_test

import (
	"context"
	"testing"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db_test"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) (*user.DBRepository, func()) {
	ctx := context.Background()
	testDB, cleanup, err := db_test.SetupTestDB(ctx)
	require.NoError(t, err)

	tm := db.NewTransactionManager(testDB)
	repo := user.NewRepository(testDB, tm)
	return repo, cleanup
}

func TestRepository_CreateAndGetByID(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	id := uuid.New().String()
	email := "user1@example.com"
	u := &user.User{
		ID:              id,
		FirebaseUID:     "fb-1",
		Email:           &email,
		Name:            "User One",
		DefaultCurrency: "USD",
	}

	err := repo.Create(ctx, u)
	require.NoError(t, err)
	assert.False(t, u.CreatedAt.IsZero())
	assert.False(t, u.UpdatedAt.IsZero())

	fetched, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, id, fetched.ID)
	assert.Equal(t, "fb-1", fetched.FirebaseUID)
	assert.Equal(t, &email, fetched.Email)
	assert.Nil(t, fetched.Phone)
	assert.Equal(t, "User One", fetched.Name)
	assert.Equal(t, "USD", fetched.DefaultCurrency)
}

func TestRepository_GetByID_NotFoundAndInvalidUUID(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	fetched, err := repo.GetByID(ctx, uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, fetched)

	_, err = repo.GetByID(ctx, "invalid-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid uuid")
}

func TestRepository_GetByFirebaseUID(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	id := uuid.New().String()
	email := "user2@example.com"
	u := &user.User{
		ID:          id,
		FirebaseUID: "fb-2",
		Email:       &email,
		Name:        "User Two",
	}
	require.NoError(t, repo.Create(ctx, u))

	fetched, err := repo.GetByFirebaseUID(ctx, "fb-2")
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, id, fetched.ID)

	notFound, err := repo.GetByFirebaseUID(ctx, "fb-non-existent")
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestRepository_GetByEmailOrPhone(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	email := "user3@example.com"
	phone := "+123456789"
	u := &user.User{
		ID:          uuid.New().String(),
		FirebaseUID: "fb-3",
		Email:       &email,
		Phone:       &phone,
		Name:        "User Three",
	}
	require.NoError(t, repo.Create(ctx, u))

	// By Email
	byEmail, err := repo.GetByEmailOrPhone(ctx, email, "")
	require.NoError(t, err)
	require.NotNil(t, byEmail)
	assert.Equal(t, u.ID, byEmail.ID)

	// By Phone
	byPhone, err := repo.GetByEmailOrPhone(ctx, "", phone)
	require.NoError(t, err)
	require.NotNil(t, byPhone)
	assert.Equal(t, u.ID, byPhone.ID)

	// Not found
	notFound, err := repo.GetByEmailOrPhone(ctx, "other@example.com", "+9999999")
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestRepository_UpdateUser(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	email := "user4@example.com"
	u := &user.User{
		ID:              uuid.New().String(),
		FirebaseUID:     "fb-4",
		Email:           &email,
		Name:            "Before Update",
		DefaultCurrency: "INR",
	}
	require.NoError(t, repo.Create(ctx, u))

	u.Name = "After Update"
	u.DefaultCurrency = "EUR"
	err := repo.UpdateUser(ctx, u)
	require.NoError(t, err)

	fetched, err := repo.GetByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, "After Update", fetched.Name)
	assert.Equal(t, "EUR", fetched.DefaultCurrency)
}

func TestRepository_Friendships(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	email1 := "user5@example.com"
	email2 := "user6@example.com"
	u1 := &user.User{ID: uuid.New().String(), FirebaseUID: "fb-5", Email: &email1, Name: "User Five"}
	u2 := &user.User{ID: uuid.New().String(), FirebaseUID: "fb-6", Email: &email2, Name: "User Six"}
	require.NoError(t, repo.Create(ctx, u1))
	require.NoError(t, repo.Create(ctx, u2))

	// Initial check
	isFriend, err := repo.GetFriendship(ctx, u1.ID, u2.ID)
	require.NoError(t, err)
	assert.False(t, isFriend)

	// Create friendship
	require.NoError(t, repo.CreateFriendship(ctx, u1.ID, u2.ID))

	isFriend, err = repo.GetFriendship(ctx, u1.ID, u2.ID)
	require.NoError(t, err)
	assert.True(t, isFriend)

	// List friends
	friends, err := repo.ListFriends(ctx, u1.ID, 10, nil, nil)
	require.NoError(t, err)
	require.Len(t, friends, 1)
	assert.Equal(t, u2.ID, friends[0].ID)

	// Delete friendship
	require.NoError(t, repo.DeleteFriendship(ctx, u1.ID, u2.ID))

	isFriend, err = repo.GetFriendship(ctx, u1.ID, u2.ID)
	require.NoError(t, err)
	assert.False(t, isFriend)
}

func TestRepository_ConstraintEdgeCases(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	email1 := "unique1@example.com"
	u1 := &user.User{
		ID:          uuid.New().String(),
		FirebaseUID: "fb-dup",
		Email:       &email1,
		Name:        "User Dup 1",
	}
	require.NoError(t, repo.Create(ctx, u1))

	// Test 1: Duplicate firebase_uid constraint
	email2 := "unique2@example.com"
	u2 := &user.User{
		ID:          uuid.New().String(),
		FirebaseUID: "fb-dup", // Duplicate!
		Email:       &email2,
		Name:        "User Dup 2",
	}
	err := repo.Create(ctx, u2)
	require.Error(t, err, "expected error on duplicate firebase_uid")

	// Test 2: Missing both email and phone constraint (email_or_phone_required)
	uNoContact := &user.User{
		ID:          uuid.New().String(),
		FirebaseUID: "fb-no-contact",
		Email:       nil,
		Phone:       nil,
		Name:        "No Contact",
	}
	err = repo.Create(ctx, uNoContact)
	require.Error(t, err, "expected error on missing email and phone check constraint")
}
