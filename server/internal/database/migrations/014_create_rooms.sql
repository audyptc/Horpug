CREATE TABLE IF NOT EXISTS rooms (
    id          UUID PRIMARY KEY,
    room_number TEXT NOT NULL UNIQUE,
    floor       INTEGER NOT NULL,
    type        TEXT NOT NULL DEFAULT 'standard',
    status      TEXT NOT NULL DEFAULT 'available',
    rent_price  NUMERIC(10,2) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
