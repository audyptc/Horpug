CREATE TABLE IF NOT EXISTS tenants (
    id                UUID PRIMARY KEY,
    first_name        TEXT NOT NULL,
    last_name         TEXT NOT NULL,
    phone             TEXT NOT NULL,
    id_card           TEXT NOT NULL UNIQUE,
    email             TEXT NOT NULL DEFAULT '',
    emergency_contact TEXT NOT NULL DEFAULT '',
    note              TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO menus (id, name, path)
VALUES ('44444444-4444-4444-4444-444444444006', 'จัดการผู้เช่า', '/tenants')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
SELECT
    '11111111-1111-1111-1111-111111111001',
    '44444444-4444-4444-4444-444444444006',
    p.id
FROM permissions p
WHERE p.name IN ('read', 'create', 'update', 'delete')
ON CONFLICT (role_id, menu_id, permission_id) DO NOTHING;
