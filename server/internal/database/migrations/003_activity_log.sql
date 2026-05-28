CREATE TABLE IF NOT EXISTS activity_logs (
    id          UUID PRIMARY KEY,
    actor_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    action      TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id   TEXT NOT NULL,
    new_value   JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_activity_logs_actor_id    ON activity_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_entity_type ON activity_logs(entity_type);
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at  ON activity_logs(created_at DESC);

-- เพิ่ม menu สำหรับ activity logs
INSERT INTO menus (id, name, path) VALUES
    ('44444444-4444-4444-4444-444444444017', 'ประวัติการใช้งาน', '/activity-logs')
ON CONFLICT DO NOTHING;

-- ให้ admin อ่าน activity logs ได้
INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
SELECT
    '11111111-1111-1111-1111-111111111001',
    '44444444-4444-4444-4444-444444444017',
    p.id
FROM permissions p
ON CONFLICT DO NOTHING;
