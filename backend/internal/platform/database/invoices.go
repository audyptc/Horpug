package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// migrateInvoiceNumbers adds running numbers to invoices.
//
// Every invoice is numbered INV<year>-<seq>, counted per dormitory and
// restarting each year (the billing period's year). invoice_counters holds
// the last number handed out per dormitory and year; a number is never
// reused, so deleting an invoice leaves a gap.
//
// Invoices created before numbering existed are numbered once here, in
// period then creation order, and the counters are advanced past them. Safe
// to run repeatedly: only invoices still without a number are touched.
func migrateInvoiceNumbers(ctx context.Context, db *pgxpool.Pool) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS invoice_counters (
			dormitory_id UUID NOT NULL,
			year INTEGER NOT NULL,
			last_seq INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (dormitory_id, year),
			CONSTRAINT invoice_counters_dormitory_fkey FOREIGN KEY (dormitory_id) REFERENCES dormitories(id) ON DELETE CASCADE
		)`,
		`ALTER TABLE invoices ADD COLUMN IF NOT EXISTS dormitory_id UUID`,
		`ALTER TABLE invoices ADD COLUMN IF NOT EXISTS invoice_year INTEGER`,
		`ALTER TABLE invoices ADD COLUMN IF NOT EXISTS invoice_seq INTEGER`,
		`ALTER TABLE invoices ADD COLUMN IF NOT EXISTS invoice_no VARCHAR(20)`,

		`WITH numbered AS (
			SELECT i.id, rm.dormitory_id, i.period_year AS y,
				COALESCE(ic.last_seq, 0) + ROW_NUMBER() OVER (
					PARTITION BY rm.dormitory_id, i.period_year
					ORDER BY i.period_month, i.created_at, i.id
				) AS seq
			FROM invoices i
			JOIN contracts c ON c.id = i.contract_id
			JOIN rooms rm ON rm.id = c.room_id
			LEFT JOIN invoice_counters ic ON ic.dormitory_id = rm.dormitory_id AND ic.year = i.period_year
			WHERE i.invoice_no IS NULL
		)
		UPDATE invoices i
		SET dormitory_id = n.dormitory_id, invoice_year = n.y, invoice_seq = n.seq,
			invoice_no = 'INV' || n.y || '-' || LPAD(n.seq::text, GREATEST(4, LENGTH(n.seq::text)), '0')
		FROM numbered n
		WHERE i.id = n.id`,
		`INSERT INTO invoice_counters (dormitory_id, year, last_seq)
			SELECT dormitory_id, invoice_year, MAX(invoice_seq) FROM invoices
			WHERE invoice_seq IS NOT NULL
			GROUP BY dormitory_id, invoice_year
		ON CONFLICT (dormitory_id, year) DO UPDATE SET last_seq = GREATEST(invoice_counters.last_seq, EXCLUDED.last_seq)`,

		`CREATE UNIQUE INDEX IF NOT EXISTS uq_invoices_number ON invoices(dormitory_id, invoice_year, invoice_seq)`,
		`CREATE INDEX IF NOT EXISTS idx_invoices_invoice_no ON invoices(invoice_no)`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'invoices_dormitory_fkey') THEN
				ALTER TABLE invoices ADD CONSTRAINT invoices_dormitory_fkey FOREIGN KEY (dormitory_id) REFERENCES dormitories(id) ON DELETE RESTRICT;
			END IF;
		END
		$$`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
