-- Add reports menu and grant all permissions to admin role
INSERT INTO menus (id, name, path) VALUES
    ('44444444-4444-4444-4444-444444444018', 'รายงาน', '/reports')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
SELECT
    '11111111-1111-1111-1111-111111111001',
    '44444444-4444-4444-4444-444444444018',
    p.id
FROM permissions p
ON CONFLICT DO NOTHING;
