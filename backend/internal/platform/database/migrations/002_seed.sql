-- Room types
INSERT INTO room_types (id, name, sort_order) VALUES
    ('standard', 'ธรรมดา', 1),
    ('deluxe',   'ดีลักซ์', 2),
    ('suite',    'สวีท',   3)
ON CONFLICT (id) DO NOTHING;

-- Roles
INSERT INTO roles (id, name, description, is_active) VALUES
    ('11111111-1111-1111-1111-111111111001', 'admin', 'ระบบผู้ดูแลหลัก', TRUE);

-- Admin user (password: admin1234)
INSERT INTO users (id, full_name, email, password, is_active) VALUES
    ('33333333-3333-3333-3333-333333333001', 'Administrator', 'admin@horpug.local', '$2a$10$h3XVKoNwcgmKbuuVE9eTRejnJFbG3ucYtb8kcYHcW5AaMyufxoOku', TRUE);

INSERT INTO user_roles (user_id, role_id) VALUES
    ('33333333-3333-3333-3333-333333333001', '11111111-1111-1111-1111-111111111001');

-- Menus (path ตรงกับ router จริง)
INSERT INTO menus (id, name, path) VALUES
    -- หน้าหลัก
    ('44444444-4444-4444-4444-444444444001', 'หน้าหลัก',          '/'),
    -- ส่วนตั้งค่า
    ('44444444-4444-4444-4444-444444444002', 'จัดการผู้ใช้',      '/settings/users'),
    ('44444444-4444-4444-4444-444444444003', 'จัดการบทบาท',       '/settings/roles'),
    ('44444444-4444-4444-4444-444444444020', 'จัดการประเภทห้อง',  '/settings/room-types'),
    ('44444444-4444-4444-4444-444444444021', 'ตั้งค่าระบบ',       '/settings/general'),
    -- ส่วนที่ 1: ห้องพัก
    ('44444444-4444-4444-4444-444444444005', 'จัดการห้องพัก',     '/rooms'),
    ('44444444-4444-4444-4444-444444444006', 'จัดการผู้เช่า',     '/tenants'),
    ('44444444-4444-4444-4444-444444444007', 'สัญญาเช่า',         '/contracts'),
    -- ส่วนที่ 2: การเงิน
    ('44444444-4444-4444-4444-444444444008', 'มิเตอร์ไฟฟ้า',      '/electric-meters'),
    ('44444444-4444-4444-4444-444444444022', 'มิเตอร์น้ำ',        '/water-meters'),
    ('44444444-4444-4444-4444-444444444009', 'ใบแจ้งหนี้',        '/bills'),
    ('44444444-4444-4444-4444-444444444012', 'การชำระเงิน',       '/payments'),
    ('44444444-4444-4444-4444-444444444010', 'ค่าใช้จ่าย',        '/expenses'),
    -- ส่วนที่ 3: บริการ
    ('44444444-4444-4444-4444-444444444011', 'แจ้งซ่อม',          '/maintenance'),
    ('44444444-4444-4444-4444-444444444013', 'ประกาศ',            '/announcements'),
    ('44444444-4444-4444-4444-444444444014', 'จอดรถ',             '/parking'),
    ('44444444-4444-4444-4444-444444444015', 'พัสดุ',             '/parcels'),
    ('44444444-4444-4444-4444-444444444016', 'เอกสาร',            '/documents'),
    -- ส่วนที่ 4: รายงาน
    ('44444444-4444-4444-4444-444444444017', 'ประวัติการใช้งาน',  '/activity-logs'),
    ('44444444-4444-4444-4444-444444444018', 'รายงาน',            '/reports'),
    ('44444444-4444-4444-4444-444444444019', 'วิเคราะห์ข้อมูล',  '/analytics');

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
