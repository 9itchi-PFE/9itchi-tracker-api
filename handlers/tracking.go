package handlers

import (
	"net/http"

	"9itchi-tracker-api/models"
	"9itchi-tracker-api/models/api"
	"9itchi-tracker-api/services"

	"github.com/gin-gonic/gin"
)

// Pixel: GET /t/:trackingId (public) 1x1 transparent GIF, records the open.
func Pixel(svc services.TrackingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, apiErr := svc.Pixel(c.Request.Context(), c.Param("trackingId"))
		if apiErr != nil {
			models.FormatError(c, apiErr)
			return
		}
		c.JSON(http.StatusOK, result)
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
