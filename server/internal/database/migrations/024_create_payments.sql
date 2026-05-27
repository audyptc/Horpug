CREATE TABLE IF NOT EXISTS payments (
    id           UUID PRIMARY KEY,
    bill_id      UUID NOT NULL REFERENCES bills(id) ON DELETE CASCADE,
    amount       NUMERIC(12, 2) NOT NULL DEFAULT 0,
    method       TEXT NOT NULL DEFAULT 'cash' CHECK (method IN ('cash', 'transfer', 'qr')),
    payment_date DATE NOT NULL,
    note         TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
