INSERT INTO roles (id, name, description)
VALUES
    ('11111111-1111-1111-1111-111111111001', 'admin', 'ระบบผู้ดูแลหลัก')
ON CONFLICT (name) DO NOTHING;