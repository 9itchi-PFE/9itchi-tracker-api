package repositories

// TrackingRepository is the only layer that touches PostgreSQL for tracking data.
// Use store.Conn(ctx), a context from QueryTimeout, and return MapError(err).
type TrackingRepository interface {
}

type trackingRepository struct {
	store *Store
}

func NewTrackingRepository(store *Store) TrackingRepository {
	return &trackingRepository{store: store}
}
