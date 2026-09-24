package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// migrateMoveOuts adds move-out settlements and each dormitory's ready-made
// deductions. A contract has at most one move-out.
func migrateMoveOuts(ctx context.Context, db *pgxpool.Pool) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS move_outs (
			id UUID PRIMARY KEY,
			contract_id UUID NOT NULL UNIQUE,
			dormitory_id UUID NOT NULL,
			move_out_date DATE NOT NULL,
			deposit NUMERIC(10,2) NOT NULL DEFAULT 0,
			rent_price NUMERIC(10,2) NOT NULL DEFAULT 0,
			days_stayed INTEGER NOT NULL DEFAULT 0,
			days_in_month INTEGER NOT NULL DEFAULT 0,
			total_deductions NUMERIC(10,2) NOT NULL DEFAULT 0,
			refund_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
			amount_due NUMERIC(10,2) NOT NULL DEFAULT 0,
			note VARCHAR(255) NOT NULL DEFAULT '',
			created_by UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT move_outs_contract_fkey FOREIGN KEY (contract_id) REFERENCES contracts(id) ON DELETE RESTRICT,
			CONSTRAINT move_outs_dormitory_fkey FOREIGN KEY (dormitory_id) REFERENCES dormitories(id) ON DELETE RESTRICT,
			CONSTRAINT move_outs_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS move_out_items (
			id UUID PRIMARY KEY,
			move_out_id UUID NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			item_type VARCHAR(20) NOT NULL,
			description VARCHAR(255) NOT NULL DEFAULT '',
			amount NUMERIC(10,2) NOT NULL,
			reference_id UUID,
			payment_id UUID,
			CONSTRAINT move_out_items_move_out_fkey FOREIGN KEY (move_out_id) REFERENCES move_outs(id) ON DELETE CASCADE,
			CONSTRAINT move_out_items_payment_fkey FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL,
			CONSTRAINT chk_move_out_items_type CHECK (item_type IN ('invoice', 'rent', 'rent_credit', 'electricity', 'water', 'other'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_move_out_items_move_out_id ON move_out_items(move_out_id)`,
		`CREATE INDEX IF NOT EXISTS idx_move_out_items_reference_id ON move_out_items(reference_id)`,
		`CREATE TABLE IF NOT EXISTS deduction_presets (
			id UUID PRIMARY KEY,
			dormitory_id UUID NOT NULL,
			name VARCHAR(150) NOT NULL,
			amount NUMERIC(10,2) NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			CONSTRAINT deduction_presets_dormitory_fkey FOREIGN KEY (dormitory_id) REFERENCES dormitories(id) ON DELETE CASCADE,
			CONSTRAINT chk_deduction_presets_amount CHECK (amount > 0)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_deduction_presets_dormitory_id ON deduction_presets(dormitory_id)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
