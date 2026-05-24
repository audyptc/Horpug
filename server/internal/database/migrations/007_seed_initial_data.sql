INSERT INTO roles (id, name, description)
VALUES
    ('11111111-1111-1111-1111-111111111001', 'admin', 'ระบบผู้ดูแลหลัก')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (id, name, description)
VALUES
    ('22222222-2222-2222-2222-222222222001', 'menus.read', 'ดูรายการเมนู'),
    ('22222222-2222-2222-2222-222222222002', 'menus.create', 'สร้างเมนูใหม่'),
    ('22222222-2222-2222-2222-222222222003', 'menus.update', 'แก้ไขเมนู'),
    ('22222222-2222-2222-2222-222222222004', 'menus.delete', 'ลบเมนู'),
    ('22222222-2222-2222-2222-222222222005', 'users.read', 'ดูข้อมูลผู้ใช้'),
    ('22222222-2222-2222-2222-222222222006', 'users.create', 'สร้างผู้ใช้ใหม่'),
    ('22222222-2222-2222-2222-222222222007', 'users.update', 'แก้ไขผู้ใช้'),
    ('22222222-2222-2222-2222-222222222008', 'users.delete', 'ลบผู้ใช้'),
    ('22222222-2222-2222-2222-222222222009', 'roles.read', 'ดูข้อมูล role'),
    ('22222222-2222-2222-2222-222222222010', 'roles.create', 'สร้าง role ใหม่'),
    ('22222222-2222-2222-2222-222222222011', 'roles.update', 'แก้ไข role'),
    ('22222222-2222-2222-2222-222222222012', 'roles.delete', 'ลบ role')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT '11111111-1111-1111-1111-111111111001', id
FROM permissions
WHERE name IN (
    'menus.read',
    'menus.create',
    'menus.update',
    'menus.delete',
    'users.read',
    'users.create',
    'users.update',
    'users.delete',
    'roles.read',
    'roles.create',
    'roles.update',
    'roles.delete'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;