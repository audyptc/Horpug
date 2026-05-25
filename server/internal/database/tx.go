package database

import "context"

// WithTx implements domain.TxManager — runs fn inside a single database transaction.
// Rolls back automatically if fn returns an error or panics.
func (db *DB) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(ctx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
