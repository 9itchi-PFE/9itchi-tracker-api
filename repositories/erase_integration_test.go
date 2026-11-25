//go:build integration

package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Everything of the deleted user goes, in one transaction, and nothing of anyone else.
func TestEraseUser(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	repo := NewEraseRepository(store)
	const gone, kept = "00000000-0000-4000-8000-0000000000e1", "00000000-0000-4000-8000-0000000000e2"
	const count = `SELECT (SELECT count(*) FROM tracking_records WHERE user_id = $1) + (SELECT count(*) FROM tracking_opens WHERE tracking_id = 'trk-' || $1::text)`
	left := func(id string) int {
		var n int
		require.NoError(t, store.Conn(ctx).QueryRow(ctx, count, id).Scan(&n))
		return n
	}
	for _, id := range []string{gone, kept} {
		for _, q := range []string{
			`WITH r AS (INSERT INTO tracking_records (tracking_id, email_id, user_id) VALUES ('trk-' || $1::text, gen_random_uuid(), $1) RETURNING tracking_id) INSERT INTO tracking_opens (tracking_id, ip_hash) SELECT tracking_id, repeat('a', 64) FROM r`,
		} {
			_, err := store.Conn(ctx).Exec(ctx, q, id)
			require.NoError(t, err, q)
		}
	}
	require.Equal(t, 2, left(gone))

	rows, err := repo.EraseUser(ctx, gone)
	require.NoError(t, err)
	assert.Positive(t, rows)
	assert.Zero(t, left(gone))
	assert.Equal(t, 2, left(kept), "another student keeps everything")

	rows, err = repo.EraseUser(ctx, gone)
	require.NoError(t, err, "erasing again is not an error")
	assert.Zero(t, rows)
}
