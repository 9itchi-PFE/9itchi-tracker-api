package mocks

import (
	"context"
	"time"

	"9itchi-tracker-api/models/db"
)

// TrackingRepository is a hand-written mock; set the Fn fields per test.
type TrackingRepository struct {
	CreateFn       func(ctx context.Context, trackingID, emailID, userID string) (*db.TrackingRecord, error)
	GetByEmailIDFn func(ctx context.Context, userID, emailID string) (*db.TrackingRecord, error)
	RecordOpenFn   func(ctx context.Context, trackingID, userAgent, ipHash string, ignoreWithin time.Duration) (*db.TrackingRecord, error)
	ListOpensFn    func(ctx context.Context, trackingID string, limit int) ([]time.Time, error)
}

func (m *TrackingRepository) Create(ctx context.Context, trackingID, emailID, userID string) (*db.TrackingRecord, error) {
	return m.CreateFn(ctx, trackingID, emailID, userID)
}

func (m *TrackingRepository) GetByEmailID(ctx context.Context, userID, emailID string) (*db.TrackingRecord, error) {
	return m.GetByEmailIDFn(ctx, userID, emailID)
}

func (m *TrackingRepository) RecordOpen(ctx context.Context, trackingID, userAgent, ipHash string, ignoreWithin time.Duration) (*db.TrackingRecord, error) {
	return m.RecordOpenFn(ctx, trackingID, userAgent, ipHash, ignoreWithin)
}

func (m *TrackingRepository) ListOpens(ctx context.Context, trackingID string, limit int) ([]time.Time, error) {
	return m.ListOpensFn(ctx, trackingID, limit)
}
