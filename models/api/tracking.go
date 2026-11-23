package api

import "time"

// RegisterTrackingRequest is the validated body of RegisterTracking.
type RegisterTrackingRequest struct {
	EmailID string `json:"email_id" binding:"required,uuid"`
	UserID  string `json:"user_id" binding:"required,uuid"`
}

// RegisterTrackingResponse returns the tracking ID and the pixel path to embed in the email.
type RegisterTrackingResponse struct {
	TrackingID   string `json:"tracking_id"`
	PixelURLPath string `json:"pixel_url_path"`
}

// OpenItem is one logged open (time only).
type OpenItem struct {
	OpenedAt time.Time `json:"opened_at"`
}

// EmailStatsResponse is the open statistics of one email.
type EmailStatsResponse struct {
	SentAt      time.Time  `json:"sent_at"`
	FirstOpenAt *time.Time `json:"first_open_at"`
	LastOpenAt  *time.Time `json:"last_open_at"`
	OpenCount   int        `json:"open_count"`
	Opens       []OpenItem `json:"opens"`
}

// EmailOpenedEvent is the payload of core.EventEmailOpened.
type EmailOpenedEvent struct {
	EmailID    string    `json:"email_id"`
	UserID     string    `json:"user_id"`
	TrackingID string    `json:"tracking_id"`
	OpenedAt   time.Time `json:"opened_at"`
}
