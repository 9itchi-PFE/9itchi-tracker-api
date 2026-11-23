package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"regexp"
	"unicode/utf8"

	"9itchi-tracker-api/core"
	"9itchi-tracker-api/models"
	"9itchi-tracker-api/models/api"
	"9itchi-tracker-api/models/mapping"
	"9itchi-tracker-api/repositories"
	"9itchi-tracker-api/utils"

	"go.uber.org/zap"
)

const (
	maxUserAgentLen = 255
	// maxTrackingIDLen bounds the lookup key of the public pixel route (ids are 43 chars).
	maxTrackingIDLen = 64
	recentOpensLimit = 20
)

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// TrackingService holds all business logic for tracking. Handlers stay thin.
type TrackingService interface {
	// GET /t/:trackingId: records the open. Never fails and never reveals whether the ID exists.
	Pixel(ctx context.Context, trackingID, userAgent, ip string)
	// POST /internal/tracking: create tracking ID for an outgoing email (idempotent per email)
	RegisterTracking(ctx context.Context, req api.RegisterTrackingRequest) (*api.RegisterTrackingResponse, *models.ApiError)
	// GET /emails/:id/stats: sent date, first open, number of opens
	EmailStats(ctx context.Context, userID string, id string) (*api.EmailStatsResponse, *models.ApiError)
}

type trackingService struct {
	repo      repositories.TrackingRepository
	publisher core.Publisher
	secretKey string
}

func NewTrackingService(repo repositories.TrackingRepository, publisher core.Publisher, secretKey string) TrackingService {
	return &trackingService{repo: repo, publisher: publisher, secretKey: secretKey}
}

func (s *trackingService) Pixel(ctx context.Context, trackingID, userAgent, ip string) {
	if trackingID == "" || len(trackingID) > maxTrackingIDLen {
		return
	}
	rec, err := s.repo.RecordOpen(ctx, trackingID, truncateRunes(userAgent, maxUserAgentLen), s.hashIP(ip), core.OpenIgnoreWindow)
	if err != nil {
		// unknown ID or prefetch window: silently ignored
		if !errors.Is(err, repositories.ErrNotFound) {
			utils.Logger.Error("record open failed", zap.Error(err))
		}
		return
	}
	if rec == nil || rec.OpenCount != 1 || rec.LastOpenAt == nil {
		return
	}
	event := api.EmailOpenedEvent{EmailID: rec.EmailID, UserID: rec.UserID, TrackingID: rec.TrackingID, OpenedAt: *rec.LastOpenAt}
	if err := s.publisher.Publish(ctx, core.EventEmailOpened, event); err != nil {
		utils.Logger.Warn("publish failed", zap.String("event", core.EventEmailOpened))
	}
}

func (s *trackingService) RegisterTracking(ctx context.Context, req api.RegisterTrackingRequest) (*api.RegisterTrackingResponse, *models.ApiError) {
	if !uuidRe.MatchString(req.EmailID) || !uuidRe.MatchString(req.UserID) {
		return nil, &models.ApiError{Code: http.StatusBadRequest, Error: errors.New("email_id and user_id must be UUIDs")}
	}
	trackingID, err := newTrackingID()
	if err != nil {
		utils.Logger.Error("generate tracking id failed", zap.Error(err))
		return nil, internalError()
	}
	rec, err := s.repo.Create(ctx, trackingID, req.EmailID, req.UserID)
	if err != nil {
		utils.Logger.Error("create tracking record failed", zap.Error(err))
		return nil, internalError()
	}
	return mapping.ToRegisterTrackingResponse(rec), nil
}

func (s *trackingService) EmailStats(ctx context.Context, userID string, id string) (*api.EmailStatsResponse, *models.ApiError) {
	if !uuidRe.MatchString(id) {
		return nil, &models.ApiError{Code: http.StatusBadRequest, Error: errors.New("invalid email id")}
	}
	rec, err := s.repo.GetByEmailID(ctx, userID, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, &models.ApiError{Code: http.StatusNotFound, Error: errors.New("not found")}
		}
		utils.Logger.Error("get tracking record failed", zap.Error(err))
		return nil, internalError()
	}
	opens, err := s.repo.ListOpens(ctx, rec.TrackingID, recentOpensLimit)
	if err != nil {
		utils.Logger.Error("list opens failed", zap.Error(err))
		return nil, internalError()
	}
	return mapping.ToEmailStatsResponse(rec, opens), nil
}

// hashIP never stores the raw IP: sha256(ip + secret) hex.
func (s *trackingService) hashIP(ip string) string {
	sum := sha256.Sum256([]byte(ip + s.secretKey))
	return hex.EncodeToString(sum[:])
}

// newTrackingID returns 32 random bytes, base64url without padding (unguessable).
func newTrackingID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func truncateRunes(s string, max int) string {
	if len(s) <= max {
		return s
	}
	s = s[:max]
	for !utf8.ValidString(s) {
		s = s[:len(s)-1]
	}
	return s
}

func internalError() *models.ApiError {
	return &models.ApiError{Code: http.StatusInternalServerError, Error: errors.New("internal server error")}
}
