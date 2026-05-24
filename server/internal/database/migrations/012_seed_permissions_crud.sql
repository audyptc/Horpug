INSERT INTO permissions (id, name, description)
VALUES
    ('22222222-2222-2222-2222-222222222101', 'read', 'Read access'),
    ('22222222-2222-2222-2222-222222222102', 'create', 'Create access'),
    ('22222222-2222-2222-2222-222222222103', 'update', 'Update access'),
    ('22222222-2222-2222-2222-222222222104', 'delete', 'Delete access'),
    ('22222222-2222-2222-2222-222222222105', 'export', 'Export access')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
SELECT
    '11111111-1111-1111-1111-111111111001',
    m.id,
    p.id
FROM menus m
CROSS JOIN permissions p
WHERE p.name IN ('read', 'create', 'update', 'delete', 'export')
ON CONFLICT (role_id, menu_id, permission_id) DO NOTHING;