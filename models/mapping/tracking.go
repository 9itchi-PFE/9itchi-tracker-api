package mapping

import (
	"time"

	"9itchi-tracker-api/models/api"
	"9itchi-tracker-api/models/db"
)

// PixelURLPath is the public path embedded in emails.
func PixelURLPath(trackingID string) string { return "/t/" + trackingID }

func ToRegisterTrackingResponse(r *db.TrackingRecord) *api.RegisterTrackingResponse {
	return &api.RegisterTrackingResponse{TrackingID: r.TrackingID, PixelURLPath: PixelURLPath(r.TrackingID)}
}

func ToEmailStatsResponse(r *db.TrackingRecord, opens []time.Time) *api.EmailStatsResponse {
	items := make([]api.OpenItem, 0, len(opens))
	for _, t := range opens {
		items = append(items, api.OpenItem{OpenedAt: t})
	}
	return &api.EmailStatsResponse{
		SentAt:      r.SentAt,
		FirstOpenAt: r.FirstOpenAt,
		LastOpenAt:  r.LastOpenAt,
		OpenCount:   r.OpenCount,
		Opens:       items,
	}
}
