package domain

import "context"

// TxManager abstracts database transaction management.
// Inject into use cases that need atomic multi-repository operations.
type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
