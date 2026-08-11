INSERT INTO menus (id, name, path)
SELECT items.id::uuid, items.name, items.path
FROM (VALUES
    ('44444444-4444-4444-4444-444444444023', 'การแจ้งเตือน', '/notifications'),
    ('44444444-4444-4444-4444-444444444024', 'ค้นหา', '/search')
) AS items(id, name, path)
WHERE NOT EXISTS (SELECT 1 FROM menus m WHERE m.path = items.path);

INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
SELECT
    '11111111-1111-1111-1111-111111111001'::uuid,
    m.id,
    p.id
FROM menus m
JOIN permissions p ON TRUE
WHERE m.path IN ('/notifications', '/search')
ON CONFLICT (role_id, menu_id, permission_id) DO NOTHING;