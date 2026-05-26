CREATE TABLE IF NOT EXISTS bills (
    id              UUID PRIMARY KEY,
    contract_id     UUID NOT NULL REFERENCES contracts(id),
    billing_month   DATE NOT NULL,
    rent_amount     NUMERIC(10,2) NOT NULL DEFAULT 0,
    electric_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    water_amount    NUMERIC(10,2) NOT NULL DEFAULT 0,
    other_amount    NUMERIC(10,2) NOT NULL DEFAULT 0,
    other_note      TEXT NOT NULL DEFAULT '',
    total_amount    NUMERIC(10,2) NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'unpaid' CHECK (status IN ('unpaid', 'paid', 'overdue')),
    due_date        DATE,
    paid_at         TIMESTAMPTZ,
    note            TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO menus (id, name, path)
VALUES ('44444444-4444-4444-4444-444444444009', 'ใบแจ้งหนี้', '/bills')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
SELECT
    '11111111-1111-1111-1111-111111111001',
    '44444444-4444-4444-4444-444444444009',
    p.id
FROM permissions p
WHERE p.name IN ('read', 'create', 'update', 'delete')
ON CONFLICT (role_id, menu_id, permission_id) DO NOTHING;
