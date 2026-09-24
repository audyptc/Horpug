package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// migrateReceiptNumbers adds receipt numbering and voiding to payments.
//
// Every payment is a receipt numbered RC<year>-<seq>, counted per dormitory
// and restarting each year (the year of the payment date). receipt_counters
// holds the last number handed out per dormitory and year; payments are
// voided rather than deleted, so a number is never reused.
//
// Payments recorded before numbering existed are numbered once here, in
// payment-date order, and the counters are advanced past them. Safe to run
// repeatedly: only payments still without a number are touched.
func migrateReceiptNumbers(ctx context.Context, db *pgxpool.Pool) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS receipt_counters (
			dormitory_id UUID NOT NULL,
			year INTEGER NOT NULL,
			last_seq INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (dormitory_id, year),
			CONSTRAINT receipt_counters_dormitory_fkey FOREIGN KEY (dormitory_id) REFERENCES dormitories(id) ON DELETE CASCADE
		)`,
		`ALTER TABLE payments ADD COLUMN IF NOT EXISTS dormitory_id UUID`,
		`ALTER TABLE payments ADD COLUMN IF NOT EXISTS receipt_year INTEGER`,
		`ALTER TABLE payments ADD COLUMN IF NOT EXISTS receipt_seq INTEGER`,
		`ALTER TABLE payments ADD COLUMN IF NOT EXISTS receipt_no VARCHAR(20)`,
		`ALTER TABLE payments ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active'`,
		`ALTER TABLE payments ADD COLUMN IF NOT EXISTS voided_at TIMESTAMPTZ`,
		`ALTER TABLE payments ADD COLUMN IF NOT EXISTS voided_by UUID`,
		`ALTER TABLE payments ADD COLUMN IF NOT EXISTS void_reason VARCHAR(255) NOT NULL DEFAULT ''`,

		// Number the payments that predate numbering, continuing after any
		// number already handed out for that dormitory and year.
		`WITH numbered AS (
			SELECT p.id, rm.dormitory_id, EXTRACT(YEAR FROM p.payment_date)::int AS y,
				COALESCE(rc.last_seq, 0) + ROW_NUMBER() OVER (
					PARTITION BY rm.dormitory_id, EXTRACT(YEAR FROM p.payment_date)
					ORDER BY p.payment_date, p.created_at, p.id
				) AS seq
			FROM payments p
			JOIN invoices i ON i.id = p.invoice_id
			JOIN contracts c ON c.id = i.contract_id
			JOIN rooms rm ON rm.id = c.room_id
			LEFT JOIN receipt_counters rc ON rc.dormitory_id = rm.dormitory_id AND rc.year = EXTRACT(YEAR FROM p.payment_date)::int
			WHERE p.receipt_no IS NULL
		)
		UPDATE payments p
		SET dormitory_id = n.dormitory_id, receipt_year = n.y, receipt_seq = n.seq,
			receipt_no = 'RC' || n.y || '-' || LPAD(n.seq::text, GREATEST(4, LENGTH(n.seq::text)), '0')
		FROM numbered n
		WHERE p.id = n.id`,
		`INSERT INTO receipt_counters (dormitory_id, year, last_seq)
			SELECT dormitory_id, receipt_year, MAX(receipt_seq) FROM payments
			WHERE receipt_seq IS NOT NULL
			GROUP BY dormitory_id, receipt_year
		ON CONFLICT (dormitory_id, year) DO UPDATE SET last_seq = GREATEST(receipt_counters.last_seq, EXCLUDED.last_seq)`,

		`CREATE UNIQUE INDEX IF NOT EXISTS uq_payments_receipt ON payments(dormitory_id, receipt_year, receipt_seq)`,
		`CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status)`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'payments_dormitory_fkey') THEN
				ALTER TABLE payments ADD CONSTRAINT payments_dormitory_fkey FOREIGN KEY (dormitory_id) REFERENCES dormitories(id) ON DELETE RESTRICT;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'payments_voided_by_fkey') THEN
				ALTER TABLE payments ADD CONSTRAINT payments_voided_by_fkey FOREIGN KEY (voided_by) REFERENCES users(id) ON DELETE SET NULL;
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
