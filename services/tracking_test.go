package services

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"9itchi-tracker-api/core"
	"9itchi-tracker-api/models/api"
	"9itchi-tracker-api/models/db"
	"9itchi-tracker-api/repositories"
	"9itchi-tracker-api/services/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testEmailID = "11111111-1111-1111-1111-111111111111"
	testUserID  = "22222222-2222-2222-2222-222222222222"
)

func newSvc(repo *mocks.TrackingRepository, pub *mocks.Publisher) TrackingService {
	return NewTrackingService(repo, pub, "secret")
}

func TestPixel(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		repoRec    *db.TrackingRecord
		repoErr    error
		wantEvents int
	}{
		{"first open publishes once", &db.TrackingRecord{TrackingID: "tid", EmailID: testEmailID, UserID: testUserID, OpenCount: 1, LastOpenAt: &now}, nil, 1},
		{"later open does not publish", &db.TrackingRecord{TrackingID: "tid", OpenCount: 2, LastOpenAt: &now}, nil, 0},
		{"unknown id or prefetch window ignored", nil, repositories.ErrNotFound, 0},
		{"repo error ignored", nil, errors.New("boom"), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotWindow time.Duration
			var gotIPHash, gotUA string
			repo := &mocks.TrackingRepository{RecordOpenFn: func(_ context.Context, _, ua, ipHash string, w time.Duration) (*db.TrackingRecord, error) {
				gotWindow, gotIPHash, gotUA = w, ipHash, ua
				return tt.repoRec, tt.repoErr
			}}
			pub := &mocks.Publisher{}
			newSvc(repo, pub).Pixel(context.Background(), "tid", strings.Repeat("a", 400), "1.2.3.4")

			assert.Equal(t, core.OpenIgnoreWindow, gotWindow)
			assert.Len(t, gotUA, 255)
			assert.Len(t, gotIPHash, 64)
			assert.NotContains(t, gotIPHash, "1.2.3.4")
			assert.Len(t, pub.Events, tt.wantEvents)
			if tt.wantEvents == 1 {
				assert.Equal(t, core.EventEmailOpened, pub.Events[0])
				ev, ok := pub.Last.(api.EmailOpenedEvent)
				require.True(t, ok)
				assert.Equal(t, api.EmailOpenedEvent{EmailID: testEmailID, UserID: testUserID, TrackingID: "tid", OpenedAt: now}, ev)
			}
		})
	}
}

func TestPixelPublishErrorIgnored(t *testing.T) {
	now := time.Now()
	repo := &mocks.TrackingRepository{RecordOpenFn: func(context.Context, string, string, string, time.Duration) (*db.TrackingRecord, error) {
		return &db.TrackingRecord{OpenCount: 1, LastOpenAt: &now}, nil
	}}
	pub := &mocks.Publisher{Err: errors.New("down")}
	assert.NotPanics(t, func() { newSvc(repo, pub).Pixel(context.Background(), "tid", "ua", "ip") })
	assert.Len(t, pub.Events, 1)
}

func TestPixelRejectsOversizedID(t *testing.T) {
	repo := &mocks.TrackingRepository{} // nil fn would panic if called
	newSvc(repo, &mocks.Publisher{}).Pixel(context.Background(), strings.Repeat("x", 65), "ua", "ip")
}

func TestRegisterTracking(t *testing.T) {
	t.Run("success returns pixel path", func(t *testing.T) {
		repo := &mocks.TrackingRepository{CreateFn: func(_ context.Context, tid, _, _ string) (*db.TrackingRecord, error) {
			return &db.TrackingRecord{TrackingID: tid}, nil
		}}
		res, apiErr := newSvc(repo, &mocks.Publisher{}).RegisterTracking(context.Background(), api.RegisterTrackingRequest{EmailID: testEmailID, UserID: testUserID})
		require.Nil(t, apiErr)
		assert.Len(t, res.TrackingID, 43)
		assert.Equal(t, "/t/"+res.TrackingID, res.PixelURLPath)
	})
	t.Run("idempotent returns existing id", func(t *testing.T) {
		repo := &mocks.TrackingRepository{CreateFn: func(context.Context, string, string, string) (*db.TrackingRecord, error) {
			return &db.TrackingRecord{TrackingID: "existing"}, nil
		}}
		res, apiErr := newSvc(repo, &mocks.Publisher{}).RegisterTracking(context.Background(), api.RegisterTrackingRequest{EmailID: testEmailID, UserID: testUserID})
		require.Nil(t, apiErr)
		assert.Equal(t, "/t/existing", res.PixelURLPath)
	})
	t.Run("bad uuid is 400", func(t *testing.T) {
		_, apiErr := newSvc(&mocks.TrackingRepository{}, &mocks.Publisher{}).RegisterTracking(context.Background(), api.RegisterTrackingRequest{EmailID: "x", UserID: testUserID})
		require.NotNil(t, apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.Code)
	})
	t.Run("repo error is generic 500", func(t *testing.T) {
		repo := &mocks.TrackingRepository{CreateFn: func(context.Context, string, string, string) (*db.TrackingRecord, error) {
			return nil, errors.New("pg secret detail")
		}}
		_, apiErr := newSvc(repo, &mocks.Publisher{}).RegisterTracking(context.Background(), api.RegisterTrackingRequest{EmailID: testEmailID, UserID: testUserID})
		require.NotNil(t, apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.Code)
		assert.NotContains(t, apiErr.Error.Error(), "pg secret")
	})
}

func TestNewTrackingIDUnique(t *testing.T) {
	a, err := newTrackingID()
	require.NoError(t, err)
	b, _ := newTrackingID()
	assert.NotEqual(t, a, b)
}

func TestEmailStats(t *testing.T) {
	now := time.Now()
	t.Run("success", func(t *testing.T) {
		repo := &mocks.TrackingRepository{
			GetByEmailIDFn: func(_ context.Context, userID, _ string) (*db.TrackingRecord, error) {
				assert.Equal(t, testUserID, userID)
				return &db.TrackingRecord{TrackingID: "tid", SentAt: now, OpenCount: 2, FirstOpenAt: &now, LastOpenAt: &now}, nil
			},
			ListOpensFn: func(_ context.Context, _ string, limit int) ([]time.Time, error) {
				assert.Equal(t, 20, limit)
				return []time.Time{now, now}, nil
			},
		}
		res, apiErr := newSvc(repo, &mocks.Publisher{}).EmailStats(context.Background(), testUserID, testEmailID)
		require.Nil(t, apiErr)
		assert.Equal(t, 2, res.OpenCount)
		assert.Len(t, res.Opens, 2)
	})
	t.Run("not owner is 404", func(t *testing.T) {
		repo := &mocks.TrackingRepository{GetByEmailIDFn: func(context.Context, string, string) (*db.TrackingRecord, error) {
			return nil, repositories.ErrNotFound
		}}
		_, apiErr := newSvc(repo, &mocks.Publisher{}).EmailStats(context.Background(), testUserID, testEmailID)
		require.NotNil(t, apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.Code)
	})
	t.Run("bad id is 400", func(t *testing.T) {
		_, apiErr := newSvc(&mocks.TrackingRepository{}, &mocks.Publisher{}).EmailStats(context.Background(), testUserID, "nope")
		require.NotNil(t, apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.Code)
	})
	t.Run("list error is 500", func(t *testing.T) {
		repo := &mocks.TrackingRepository{
			GetByEmailIDFn: func(context.Context, string, string) (*db.TrackingRecord, error) { return &db.TrackingRecord{}, nil },
			ListOpensFn:    func(context.Context, string, int) ([]time.Time, error) { return nil, errors.New("x") },
		}
		_, apiErr := newSvc(repo, &mocks.Publisher{}).EmailStats(context.Background(), testUserID, testEmailID)
		require.NotNil(t, apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.Code)
	})
}
