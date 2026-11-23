//go:build integration

package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationsCreateTables(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	for _, table := range []string{"tracking_records"} {
		var n int
		require.NoError(t, store.Conn(ctx).QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n), table)
		assert.Equal(t, 0, n)
	}
}

func TestUUIDScansIntoString(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	var id string
	require.NoError(t, store.Conn(ctx).QueryRow(ctx, "SELECT gen_random_uuid()").Scan(&id))
	assert.Len(t, id, 36)
}

func TestWithTxRollsBack(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	_ = store.WithTx(ctx, func(ctx context.Context) error {
		_, err := store.Conn(ctx).Exec(ctx, "CREATE TEMP TABLE x(a int)")
		require.NoError(t, err)
		return assert.AnError
	})
	var n int
	err := store.Conn(ctx).QueryRow(ctx, "SELECT count(*) FROM x").Scan(&n)
	assert.Error(t, err)
}
