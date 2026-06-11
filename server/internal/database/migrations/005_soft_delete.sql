-- Add soft delete to core business entities
ALTER TABLE users          ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE rooms          ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE tenants        ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE contracts      ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE bills          ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE payments       ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE meter_readings ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE expenses       ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;

-- Replace global unique constraints with partial unique indexes so that soft-deleted
-- records do not block re-creation of the same value.
ALTER TABLE users   DROP CONSTRAINT IF EXISTS users_email_key;
ALTER TABLE rooms   DROP CONSTRAINT IF EXISTS rooms_room_number_key;
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_id_card_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_active       ON users   (email)       WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_rooms_room_number_active ON rooms   (room_number) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_id_card_active   ON tenants (id_card)     WHERE deleted_at IS NULL;
