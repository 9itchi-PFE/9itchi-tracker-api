package db

import "time"

// TrackingRecord maps table tracking_records.
type TrackingRecord struct {
	ID          string     `db:"id" json:"id"`
	TrackingID  string     `db:"tracking_id" json:"tracking_id"`
	EmailID     string     `db:"email_id" json:"email_id"`
	UserID      string     `db:"user_id" json:"user_id"`
	SentAt      time.Time  `db:"sent_at" json:"sent_at"`
	FirstOpenAt *time.Time `db:"first_open_at" json:"first_open_at"`
	OpenCount   int        `db:"open_count" json:"open_count"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}
