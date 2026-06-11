-- Split meter_readings into electric_meter_readings and water_meter_readings

CREATE TABLE IF NOT EXISTS electric_meter_readings (
    id               UUID PRIMARY KEY,
    room_id          UUID NOT NULL REFERENCES rooms(id),
    reading_date     DATE NOT NULL,
    previous_reading NUMERIC(12,2) NOT NULL DEFAULT 0,
    current_reading  NUMERIC(12,2) NOT NULL,
    unit_price       NUMERIC(10,4) NOT NULL,
    note             TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS water_meter_readings (
    id               UUID PRIMARY KEY,
    room_id          UUID NOT NULL REFERENCES rooms(id),
    billing_type     TEXT NOT NULL DEFAULT 'meter' CHECK (billing_type IN ('meter', 'flat')),
    reading_date     DATE NOT NULL,
    previous_reading NUMERIC(12,2),
    current_reading  NUMERIC(12,2),
    unit_price       NUMERIC(10,4),
    flat_amount      NUMERIC(10,2),
    note             TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ DEFAULT NULL
);

-- Migrate existing data
INSERT INTO electric_meter_readings
    (id, room_id, reading_date, previous_reading, current_reading, unit_price, note, created_at, updated_at, deleted_at)
SELECT id, room_id, reading_date, previous_reading, current_reading, unit_price, note, created_at, updated_at, deleted_at
FROM meter_readings
WHERE meter_type = 'electric';

INSERT INTO water_meter_readings
    (id, room_id, billing_type, reading_date, previous_reading, current_reading, unit_price, note, created_at, updated_at, deleted_at)
SELECT id, room_id, 'meter', reading_date, previous_reading, current_reading, unit_price, note, created_at, updated_at, deleted_at
FROM meter_readings
WHERE meter_type = 'water';

DROP TABLE IF EXISTS meter_readings;
