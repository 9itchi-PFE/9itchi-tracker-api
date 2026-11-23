package services

import (
	"context"

	"9itchi-tracker-api/models"
	"9itchi-tracker-api/models/api"
	"9itchi-tracker-api/repositories"
)

// TrackingService holds all business logic for tracking. Handlers stay thin.
type TrackingService interface {
	// GET /t/:trackingId: 1x1 transparent GIF, records the open
	Pixel(ctx context.Context, trackingId string) (any, *models.ApiError)
	// POST /internal/tracking: create tracking ID for an outgoing email
	RegisterTracking(ctx context.Context, req api.RegisterTrackingRequest) (any, *models.ApiError)
	// GET /emails/:id/stats: sent date, first open, number of opens
	EmailStats(ctx context.Context, userID string, id string) (any, *models.ApiError)
}

type trackingService struct {
	repo repositories.TrackingRepository
}

func NewTrackingService(repo repositories.TrackingRepository) TrackingService {
	return &trackingService{repo: repo}
}

func (s *trackingService) Pixel(_ context.Context, _ string) (any, *models.ApiError) {
	return nil, models.NotImplemented()
}

func (s *trackingService) RegisterTracking(_ context.Context, _ api.RegisterTrackingRequest) (any, *models.ApiError) {
	return nil, models.NotImplemented()
}

func (s *trackingService) EmailStats(_ context.Context, _ string, _ string) (any, *models.ApiError) {
	return nil, models.NotImplemented()
}
