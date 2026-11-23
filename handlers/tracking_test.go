package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"9itchi-tracker-api/models"
	"9itchi-tracker-api/models/api"
	"9itchi-tracker-api/services/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPixelServesGIF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var gotID, gotUA string
	svc := &mocks.TrackingService{PixelFn: func(_ context.Context, id, ua, _ string) { gotID, gotUA = id, ua }}
	r := gin.New()
	r.GET("/t/:trackingId", Pixel(svc))

	for _, id := range []string{"known", "unknown-id"} {
		req := httptest.NewRequest(http.MethodGet, "/t/"+id, nil)
		req.Header.Set("User-Agent", "mailclient")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
		assert.Equal(t, "no-store, no-cache, must-revalidate, private", w.Header().Get("Cache-Control"))
		assert.Equal(t, transparentGIF, w.Body.Bytes())
		assert.Equal(t, id, gotID)
		assert.Equal(t, "mailclient", gotUA)
	}
	assert.True(t, strings.HasPrefix(string(transparentGIF), "GIF89a"))
}

func TestRegisterTracking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mocks.TrackingService{RegisterTrackingFn: func(_ context.Context, req api.RegisterTrackingRequest) (*api.RegisterTrackingResponse, *models.ApiError) {
		return &api.RegisterTrackingResponse{TrackingID: "abc", PixelURLPath: "/t/abc"}, nil
	}}
	r := gin.New()
	r.POST("/internal/tracking", RegisterTracking(svc))

	body := `{"email_id":"11111111-1111-1111-1111-111111111111","user_id":"22222222-2222-2222-2222-222222222222"}`
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/internal/tracking", strings.NewReader(body)))
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.JSONEq(t, `{"tracking_id":"abc","pixel_url_path":"/t/abc"}`, w.Body.String())

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/internal/tracking", strings.NewReader(`{"email_id":"x"}`)))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mocks.TrackingService{EmailStatsFn: func(_ context.Context, userID, id string) (*api.EmailStatsResponse, *models.ApiError) {
		if id == "missing" {
			return nil, &models.ApiError{Code: http.StatusNotFound, Error: errors.New("not found")}
		}
		assert.Equal(t, "u1", userID)
		return &api.EmailStatsResponse{OpenCount: 3, Opens: []api.OpenItem{}}, nil
	}}
	r := gin.New()
	r.GET("/emails/:id/stats", func(c *gin.Context) { c.Set("user_id", "u1"); EmailStats(svc)(c) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/emails/e1/stats", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"open_count":3`)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/emails/missing/stats", nil))
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), `"error"`)
}
