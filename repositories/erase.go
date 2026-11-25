package repositories

import "context"

// EraseRepository removes what this service keeps about one user, when their account is deleted.
type EraseRepository interface {
	// EraseUser deletes the tracking of the student's emails, with their opens.
	// It returns how many rows went. Erasing a user who has nothing here, or twice, is not an error.
	EraseUser(ctx context.Context, userID string) (int64, error)
}

type eraseRepository struct {
	store *Store
}

func NewEraseRepository(store *Store) EraseRepository {
	return &eraseRepository{store: store}
}

// eraseStatements run in this order, in one transaction: all of it goes or none of it.
var eraseStatements = []string{
	`DELETE FROM tracking_records WHERE user_id = $1`,
}

func (r *eraseRepository) EraseUser(ctx context.Context, userID string) (int64, error) {
	ctx, cancel := QueryTimeout(ctx)
	defer cancel()
	var rows int64
	err := r.store.WithTx(ctx, func(ctx context.Context) error {
		rows = 0
		for _, q := range eraseStatements {
			tag, err := r.store.Conn(ctx).Exec(ctx, q, userID)
			if err != nil {
				return MapError(err)
			}
			rows += tag.RowsAffected()
		}
		return nil
	})
	return rows, err
}
