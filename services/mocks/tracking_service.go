package mocks

import (
	"context"

	"9itchi-tracker-api/models"
	"9itchi-tracker-api/models/api"
)

// TrackingService is a hand-written mock for handler tests.
type TrackingService struct {
	PixelFn            func(ctx context.Context, trackingID, userAgent, ip string)
	RegisterTrackingFn func(ctx context.Context, req api.RegisterTrackingRequest) (*api.RegisterTrackingResponse, *models.ApiError)
	EmailStatsFn       func(ctx context.Context, userID, emailID string) (*api.EmailStatsResponse, *models.ApiError)
}

func (m *TrackingService) Pixel(ctx context.Context, trackingID, userAgent, ip string) {
	m.PixelFn(ctx, trackingID, userAgent, ip)
}

func (m *TrackingService) RegisterTracking(ctx context.Context, req api.RegisterTrackingRequest) (*api.RegisterTrackingResponse, *models.ApiError) {
	return m.RegisterTrackingFn(ctx, req)
}

func (m *TrackingService) EmailStats(ctx context.Context, userID, emailID string) (*api.EmailStatsResponse, *models.ApiError) {
	return m.EmailStatsFn(ctx, userID, emailID)
}
