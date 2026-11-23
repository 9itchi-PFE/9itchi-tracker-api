package repositories

import (
	"context"
	"time"

	"9itchi-tracker-api/models/db"

	"github.com/jackc/pgx/v5"
)

const trackingColumns = `id, tracking_id, email_id, user_id, sent_at, first_open_at, last_open_at, open_count, created_at, updated_at`

// TrackingRepository is the only layer that touches PostgreSQL for tracking data.
type TrackingRepository interface {
	// Create inserts a record, or returns the existing one for the same email_id (idempotent).
	Create(ctx context.Context, trackingID, emailID, userID string) (*db.TrackingRecord, error)
	// GetByEmailID returns the record of an email owned by userID, else ErrNotFound.
	GetByEmailID(ctx context.Context, userID, emailID string) (*db.TrackingRecord, error)
	// RecordOpen counts one open with a single atomic UPDATE and logs it. Returns ErrNotFound when the
	// tracking ID is unknown or the record was sent less than ignoreWithin ago (nothing is counted).
	RecordOpen(ctx context.Context, trackingID, userAgent, ipHash string, ignoreWithin time.Duration) (*db.TrackingRecord, error)
	// ListOpens returns the latest open times of a tracking ID, newest first.
	ListOpens(ctx context.Context, trackingID string, limit int) ([]time.Time, error)
}

type trackingRepository struct {
	store *Store
}

func NewTrackingRepository(store *Store) TrackingRepository {
	return &trackingRepository{store: store}
}

func (r *trackingRepository) Create(ctx context.Context, trackingID, emailID, userID string) (*db.TrackingRecord, error) {
	ctx, cancel := QueryTimeout(ctx)
	defer cancel()
	// The no-op DO UPDATE makes RETURNING yield the existing row on an email_id conflict.
	rows, err := r.store.Conn(ctx).Query(ctx, `
		INSERT INTO tracking_records (tracking_id, email_id, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (email_id) DO UPDATE SET email_id = EXCLUDED.email_id
		RETURNING `+trackingColumns, trackingID, emailID, userID)
	if err != nil {
		return nil, MapError(err)
	}
	rec, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[db.TrackingRecord])
	return rec, MapError(err)
}

func (r *trackingRepository) GetByEmailID(ctx context.Context, userID, emailID string) (*db.TrackingRecord, error) {
	ctx, cancel := QueryTimeout(ctx)
	defer cancel()
	rows, err := r.store.Conn(ctx).Query(ctx,
		`SELECT `+trackingColumns+` FROM tracking_records WHERE email_id = $1 AND user_id = $2`, emailID, userID)
	if err != nil {
		return nil, MapError(err)
	}
	rec, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[db.TrackingRecord])
	return rec, MapError(err)
}

func (r *trackingRepository) RecordOpen(ctx context.Context, trackingID, userAgent, ipHash string, ignoreWithin time.Duration) (*db.TrackingRecord, error) {
	var rec *db.TrackingRecord
	err := r.store.WithTx(ctx, func(ctx context.Context) error {
		ctx, cancel := QueryTimeout(ctx)
		defer cancel()
		rows, err := r.store.Conn(ctx).Query(ctx, `
			UPDATE tracking_records
			SET open_count = open_count + 1,
			    first_open_at = COALESCE(first_open_at, now()),
			    last_open_at = now(),
			    updated_at = now()
			WHERE tracking_id = $1 AND sent_at <= now() - make_interval(secs => $2)
			RETURNING `+trackingColumns, trackingID, ignoreWithin.Seconds())
		if err != nil {
			return err
		}
		rec, err = pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[db.TrackingRecord])
		if err != nil {
			return err
		}
		_, err = r.store.Conn(ctx).Exec(ctx,
			`INSERT INTO tracking_opens (tracking_id, opened_at, user_agent, ip_hash) VALUES ($1, $2, $3, $4)`,
			trackingID, rec.LastOpenAt, userAgent, ipHash)
		return err
	})
	if err != nil {
		return nil, MapError(err)
	}
	return rec, nil
}

func (r *trackingRepository) ListOpens(ctx context.Context, trackingID string, limit int) ([]time.Time, error) {
	ctx, cancel := QueryTimeout(ctx)
	defer cancel()
	rows, err := r.store.Conn(ctx).Query(ctx,
		`SELECT opened_at FROM tracking_opens WHERE tracking_id = $1 ORDER BY opened_at DESC, id DESC LIMIT $2`, trackingID, limit)
	if err != nil {
		return nil, MapError(err)
	}
	out, err := pgx.CollectRows(rows, pgx.RowTo[time.Time])
	return out, MapError(err)
}
