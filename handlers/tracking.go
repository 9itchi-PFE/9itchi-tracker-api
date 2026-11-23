package handlers

import (
	"net/http"

	"9itchi-tracker-api/models"
	"9itchi-tracker-api/models/api"
	"9itchi-tracker-api/services"

	"github.com/gin-gonic/gin"
)

// transparentGIF is a 1x1 transparent GIF89a.
var transparentGIF = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xff, 0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b,
}

// Pixel: GET /t/:trackingId (public) 1x1 transparent GIF, records the open.
// Always 200 with the GIF, even for unknown IDs (never reveal existence).
func Pixel(svc services.TrackingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc.Pixel(c.Request.Context(), c.Param("trackingId"), c.Request.UserAgent(), c.ClientIP())
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Data(http.StatusOK, "image/gif", transparentGIF)
	}
}

// RegisterTracking: POST /internal/tracking (internal) create tracking ID for an outgoing email.
func RegisterTracking(svc services.TrackingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req api.RegisterTrackingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			models.FormatError(c, &models.ApiError{Code: http.StatusBadRequest, Error: err})
			return
		}
		result, apiErr := svc.RegisterTracking(c.Request.Context(), req)
		if apiErr != nil {
			models.FormatError(c, apiErr)
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

// EmailStats: GET /emails/:id/stats (user) sent date, first open, number of opens.
func EmailStats(svc services.TrackingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, apiErr := svc.EmailStats(c.Request.Context(), c.GetString("user_id"), c.Param("id"))
		if apiErr != nil {
			models.FormatError(c, apiErr)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
