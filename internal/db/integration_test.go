//go:build integration

package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresConnection(t *testing.T) {
	ctx := context.Background()
	testDB, cleanup, err := SetupTestDB(ctx)
	require.NoError(t, err)
	defer cleanup()

	err = testDB.Ping(ctx)
	assert.NoError(t, err)
}
