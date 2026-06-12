ALTER TABLE bills DROP COLUMN IF EXISTS other_note;

CREATE TABLE IF NOT EXISTS bill_other_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bill_id    UUID NOT NULL REFERENCES bills(id) ON DELETE CASCADE,
    label      TEXT NOT NULL DEFAULT '',
    amount     NUMERIC(10,2) NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bill_other_items_bill_id ON bill_other_items(bill_id);
