//go:build integration

package repositories

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	itEmailID = "11111111-1111-1111-1111-111111111111"
	itUserID  = "22222222-2222-2222-2222-222222222222"
)

// ageRecord moves sent_at into the past so opens are outside the prefetch window.
func ageRecord(t *testing.T, store *Store, trackingID string) {
	t.Helper()
	ctx := context.Background()
	_, err := store.Conn(ctx).Exec(ctx, `UPDATE tracking_records SET sent_at = now() - interval '1 minute' WHERE tracking_id = $1`, trackingID)
	require.NoError(t, err)
}

func TestTrackingCreateIdempotent(t *testing.T) {
	repo := NewTrackingRepository(newTestStore(t))
	ctx := context.Background()
	a, err := repo.Create(ctx, "tid-a", itEmailID, itUserID)
	require.NoError(t, err)
	b, err := repo.Create(ctx, "tid-b", itEmailID, itUserID)
	require.NoError(t, err)
	assert.Equal(t, "tid-a", b.TrackingID)
	assert.Equal(t, a.ID, b.ID)
}

func TestTrackingTrackingIDUnique(t *testing.T) {
	repo := NewTrackingRepository(newTestStore(t))
	ctx := context.Background()
	_, err := repo.Create(ctx, "same", itEmailID, itUserID)
	require.NoError(t, err)
	_, err = repo.Create(ctx, "same", "33333333-3333-3333-3333-333333333333", itUserID)
	assert.ErrorIs(t, err, ErrConflict)
}

func TestTrackingGetByEmailIDScopedToUser(t *testing.T) {
	repo := NewTrackingRepository(newTestStore(t))
	ctx := context.Background()
	_, err := repo.Create(ctx, "tid", itEmailID, itUserID)
	require.NoError(t, err)
	rec, err := repo.GetByEmailID(ctx, itUserID, itEmailID)
	require.NoError(t, err)
	assert.Equal(t, "tid", rec.TrackingID)
	_, err = repo.GetByEmailID(ctx, "99999999-9999-9999-9999-999999999999", itEmailID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestTrackingRecordOpen(t *testing.T) {
	store := newTestStore(t)
	repo := NewTrackingRepository(store)
	ctx := context.Background()
	_, err := repo.Create(ctx, "tid", itEmailID, itUserID)
	require.NoError(t, err)

	// inside the prefetch window: not counted
	_, err = repo.RecordOpen(ctx, "tid", "ua", "hash", 10*time.Second)
	assert.ErrorIs(t, err, ErrNotFound)
	// unknown id
	_, err = repo.RecordOpen(ctx, "nope", "ua", "hash", 0)
	assert.ErrorIs(t, err, ErrNotFound)

	ageRecord(t, store, "tid")
	first, err := repo.RecordOpen(ctx, "tid", "ua", "hash", 10*time.Second)
	require.NoError(t, err)
	assert.Equal(t, 1, first.OpenCount)
	require.NotNil(t, first.FirstOpenAt)
	second, err := repo.RecordOpen(ctx, "tid", "ua", "hash", 10*time.Second)
	require.NoError(t, err)
	assert.Equal(t, 2, second.OpenCount)
	assert.Equal(t, first.FirstOpenAt.UnixMicro(), second.FirstOpenAt.UnixMicro())

	opens, err := repo.ListOpens(ctx, "tid", 20)
	require.NoError(t, err)
	assert.Len(t, opens, 2)
	assert.False(t, opens[0].Before(opens[1]))
}

func TestTrackingConcurrentOpensCountExactly(t *testing.T) {
	store := newTestStore(t)
	repo := NewTrackingRepository(store)
	ctx := context.Background()
	_, err := repo.Create(ctx, "tid", itEmailID, itUserID)
	require.NoError(t, err)
	ageRecord(t, store, "tid")

	const n = 20
	var firsts atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec, err := repo.RecordOpen(ctx, "tid", "ua", "hash", 10*time.Second)
			if assert.NoError(t, err) && rec.OpenCount == 1 {
				firsts.Add(1)
			}
		}()
	}
	wg.Wait()

	rec, err := repo.GetByEmailID(ctx, itUserID, itEmailID)
	require.NoError(t, err)
	assert.Equal(t, n, rec.OpenCount)
	assert.EqualValues(t, 1, firsts.Load())
	opens, err := repo.ListOpens(ctx, "tid", 100)
	require.NoError(t, err)
	assert.Len(t, opens, n)
}
