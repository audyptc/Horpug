CREATE TABLE IF NOT EXISTS expenses (
    id           UUID PRIMARY KEY,
    expense_date DATE NOT NULL,
    category     TEXT NOT NULL DEFAULT 'other' CHECK (category IN ('repair', 'utilities', 'supplies', 'salary', 'other')),
    description  TEXT NOT NULL,
    amount       NUMERIC(10,2) NOT NULL DEFAULT 0,
    note         TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
