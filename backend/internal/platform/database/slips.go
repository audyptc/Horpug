package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// migratePaymentSlips adds transfer slips tenants send from the LINE pages.
// A slip waits as "pending" until staff approve it (which records a payment,
// linked in payment_id) or reject it with a reason; the tenant may cancel a
// pending one. The image stays in the file store under file_key.
func migratePaymentSlips(ctx context.Context, db *pgxpool.Pool) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS payment_slips (
			id UUID PRIMARY KEY,
			invoice_id UUID NOT NULL,
			tenant_id UUID NOT NULL,
			amount NUMERIC(10,2) NOT NULL,
			transfer_date DATE NOT NULL,
			note VARCHAR(255) NOT NULL DEFAULT '',
			file_key VARCHAR(255) NOT NULL,
			file_mime VARCHAR(100) NOT NULL,
			file_size BIGINT NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			reject_reason VARCHAR(255) NOT NULL DEFAULT '',
			payment_id UUID,
			reviewed_by UUID,
			reviewed_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT payment_slips_invoice_fkey FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
			CONSTRAINT payment_slips_tenant_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
			CONSTRAINT payment_slips_payment_fkey FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL,
			CONSTRAINT payment_slips_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT chk_payment_slips_amount CHECK (amount > 0),
			CONSTRAINT chk_payment_slips_status CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_payment_slips_invoice_id ON payment_slips(invoice_id)`,
		`CREATE INDEX IF NOT EXISTS idx_payment_slips_status ON payment_slips(status, created_at)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
