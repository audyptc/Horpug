CREATE TABLE IF NOT EXISTS contracts (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL REFERENCES tenants(id),
    room_id    UUID NOT NULL REFERENCES rooms(id),
    start_date DATE NOT NULL,
    end_date   DATE,
    rent_price NUMERIC(10,2) NOT NULL,
    deposit    NUMERIC(10,2) NOT NULL DEFAULT 0,
    status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'expired', 'terminated')),
    note       TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO menus (id, name, path)
VALUES ('44444444-4444-4444-4444-444444444007', 'สัญญาเช่า', '/contracts')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
SELECT
    '11111111-1111-1111-1111-111111111001',
    '44444444-4444-4444-4444-444444444007',
    p.id
FROM permissions p
WHERE p.name IN ('read', 'create', 'update', 'delete')
ON CONFLICT (role_id, menu_id, permission_id) DO NOTHING;
