INSERT INTO menus (id, name, path)
VALUES ('44444444-4444-4444-4444-444444444005', 'จัดการห้องพัก', '/rooms')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
SELECT
    '11111111-1111-1111-1111-111111111001',
    '44444444-4444-4444-4444-444444444005',
    p.id
FROM permissions p
WHERE p.name IN ('read', 'create', 'update', 'delete')
ON CONFLICT (role_id, menu_id, permission_id) DO NOTHING;
