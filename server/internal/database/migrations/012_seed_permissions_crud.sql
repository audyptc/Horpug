INSERT INTO permissions (id, name, description)
VALUES
    ('22222222-2222-2222-2222-222222222013', 'permissions.read', 'ดูข้อมูล permission'),
    ('22222222-2222-2222-2222-222222222014', 'permissions.create', 'สร้าง permission ใหม่'),
    ('22222222-2222-2222-2222-222222222015', 'permissions.update', 'แก้ไข permission'),
    ('22222222-2222-2222-2222-222222222016', 'permissions.delete', 'ลบ permission')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT '11111111-1111-1111-1111-111111111001', id
FROM permissions
WHERE name IN (
    'permissions.read',
    'permissions.create',
    'permissions.update',
    'permissions.delete'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;