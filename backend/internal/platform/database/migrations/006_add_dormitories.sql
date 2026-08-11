CREATE TABLE IF NOT EXISTS dormitories (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    address    TEXT NOT NULL DEFAULT '',
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_dormitories (
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    dormitory_id UUID NOT NULL REFERENCES dormitories(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, dormitory_id)
);

CREATE INDEX IF NOT EXISTS idx_user_dormitories_dormitory_id ON user_dormitories(dormitory_id);

-- seed a default dormitory and backfill existing data so nothing breaks post-migration
INSERT INTO dormitories (id, name) VALUES (gen_random_uuid(), 'สาขาหลัก');

ALTER TABLE rooms ADD COLUMN dormitory_id UUID;
UPDATE rooms SET dormitory_id = (SELECT id FROM dormitories LIMIT 1);
ALTER TABLE rooms ALTER COLUMN dormitory_id SET NOT NULL;
ALTER TABLE rooms ADD CONSTRAINT fk_rooms_dormitory FOREIGN KEY (dormitory_id) REFERENCES dormitories(id);

ALTER TABLE tenants ADD COLUMN dormitory_id UUID;
UPDATE tenants SET dormitory_id = (SELECT id FROM dormitories LIMIT 1);
ALTER TABLE tenants ALTER COLUMN dormitory_id SET NOT NULL;
ALTER TABLE tenants ADD CONSTRAINT fk_tenants_dormitory FOREIGN KEY (dormitory_id) REFERENCES dormitories(id);

DROP INDEX IF EXISTS idx_rooms_room_number_active;
CREATE UNIQUE INDEX idx_rooms_room_number_active ON rooms (dormitory_id, room_number) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_rooms_dormitory_id   ON rooms(dormitory_id);
CREATE INDEX IF NOT EXISTS idx_tenants_dormitory_id ON tenants(dormitory_id);

INSERT INTO user_dormitories (user_id, dormitory_id)
SELECT id, (SELECT id FROM dormitories LIMIT 1) FROM users;
