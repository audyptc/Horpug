-- Roles
INSERT INTO roles (id, name, description, is_active) VALUES
    ('11111111-1111-1111-1111-111111111001', 'admin', 'ระบบผู้ดูแลหลัก', TRUE);

-- Admin user (password: admin1234)
INSERT INTO users (id, full_name, email, password, is_active) VALUES
    ('33333333-3333-3333-3333-333333333001', 'Administrator', 'admin@horpug.local', '$2a$10$np/a/xuyJ/ozav.f4TU0xueDuUgCRtejuKTCjvjb/u9C31k.loz8u', TRUE);

INSERT INTO user_roles (user_id, role_id) VALUES
    ('33333333-3333-3333-3333-333333333001', '11111111-1111-1111-1111-111111111001');

-- Menus
INSERT INTO menus (id, name, path) VALUES
    ('44444444-4444-4444-4444-444444444001', 'หน้าหลัก',       '/dashboard'),
    ('44444444-4444-4444-4444-444444444002', 'จัดการผู้ใช้',   '/users'),
    ('44444444-4444-4444-4444-444444444003', 'จัดการบทบาท',    '/roles'),
    ('44444444-4444-4444-4444-444444444004', 'จัดการสิทธิ์',   '/permissions'),
    ('44444444-4444-4444-4444-444444444005', 'จัดการห้องพัก',  '/rooms'),
    ('44444444-4444-4444-4444-444444444006', 'จัดการผู้เช่า',  '/tenants'),
    ('44444444-4444-4444-4444-444444444007', 'สัญญาเช่า',      '/contracts'),
    ('44444444-4444-4444-4444-444444444008', 'มิเตอร์',        '/meters'),
    ('44444444-4444-4444-4444-444444444009', 'ใบแจ้งหนี้',     '/bills'),
    ('44444444-4444-4444-4444-444444444010', 'ค่าใช้จ่าย',     '/expenses'),
    ('44444444-4444-4444-4444-444444444011', 'แจ้งซ่อม',       '/maintenance'),
    ('44444444-4444-4444-4444-444444444012', 'การชำระเงิน',    '/payments'),
    ('44444444-4444-4444-4444-444444444013', 'ประกาศ',         '/announcements'),
    ('44444444-4444-4444-4444-444444444014', 'จอดรถ',          '/parking'),
    ('44444444-4444-4444-4444-444444444015', 'พัสดุ',          '/parcels'),
    ('44444444-4444-4444-4444-444444444016', 'เอกสาร',         '/documents');

-- Permissions
INSERT INTO permissions (id, name, description) VALUES
    ('22222222-2222-2222-2222-222222222101', 'read',   'Read access'),
    ('22222222-2222-2222-2222-222222222102', 'create', 'Create access'),
    ('22222222-2222-2222-2222-222222222103', 'update', 'Update access'),
    ('22222222-2222-2222-2222-222222222104', 'delete', 'Delete access'),
    ('22222222-2222-2222-2222-222222222105', 'export', 'Export access');

-- Admin gets all permissions on all menus
INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
SELECT
    '11111111-1111-1111-1111-111111111001',
    m.id,
    p.id
FROM menus m
CROSS JOIN permissions p;
