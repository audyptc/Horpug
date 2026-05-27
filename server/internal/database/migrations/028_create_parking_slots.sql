CREATE TABLE IF NOT EXISTS parking_slots (
    id            UUID PRIMARY KEY,
    slot_number   TEXT NOT NULL DEFAULT '',
    vehicle_type  TEXT NOT NULL DEFAULT 'car' CHECK (vehicle_type IN ('car', 'motorcycle')),
    status        TEXT NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'occupied')),
    tenant_id     UUID REFERENCES tenants(id) ON DELETE SET NULL,
    license_plate TEXT NOT NULL DEFAULT '',
    monthly_fee   NUMERIC(10,2) NOT NULL DEFAULT 0,
    start_date    DATE,
    note          TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
