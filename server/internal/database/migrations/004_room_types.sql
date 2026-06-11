CREATE TABLE IF NOT EXISTS room_types (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO room_types (id, name, sort_order) VALUES
    ('standard', 'ธรรมดา', 1),
    ('deluxe',   'ดีลักซ์', 2),
    ('suite',    'สวีท',   3)
ON CONFLICT (id) DO NOTHING;
