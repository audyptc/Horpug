CREATE TABLE IF NOT EXISTS parcels (
    id              UUID PRIMARY KEY,
    tracking_number TEXT NOT NULL DEFAULT '',
    recipient_name  TEXT NOT NULL DEFAULT '',
    room_number     TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'picked_up')),
    received_date   DATE NOT NULL DEFAULT CURRENT_DATE,
    picked_up_date  DATE,
    note            TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
