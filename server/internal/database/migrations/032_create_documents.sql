CREATE TABLE IF NOT EXISTS documents (
    id          UUID PRIMARY KEY,
    title       TEXT NOT NULL DEFAULT '',
    category    TEXT NOT NULL DEFAULT 'other' CHECK (category IN ('contract', 'id_card', 'house_registration', 'receipt', 'other')),
    tenant_id   UUID REFERENCES tenants(id) ON DELETE SET NULL,
    file_url    TEXT NOT NULL DEFAULT '',
    issue_date  DATE,
    expiry_date DATE,
    note        TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
