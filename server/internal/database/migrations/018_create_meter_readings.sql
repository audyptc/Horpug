CREATE TABLE IF NOT EXISTS meter_readings (
    id               UUID PRIMARY KEY,
    room_id          UUID NOT NULL REFERENCES rooms(id),
    meter_type       TEXT NOT NULL CHECK (meter_type IN ('electric', 'water')),
    reading_date     DATE NOT NULL,
    previous_reading NUMERIC(12,2) NOT NULL DEFAULT 0,
    current_reading  NUMERIC(12,2) NOT NULL,
    unit_price       NUMERIC(10,4) NOT NULL,
    note             TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO menus (id, name, path)
VALUES ('44444444-4444-4444-4444-444444444008', 'มิเตอร์', '/meters')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
SELECT
    '11111111-1111-1111-1111-111111111001',
    '44444444-4444-4444-4444-444444444008',
    p.id
FROM permissions p
WHERE p.name IN ('read', 'create', 'update', 'delete')
ON CONFLICT (role_id, menu_id, permission_id) DO NOTHING;
