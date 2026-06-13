-- Ensure reports menu exists, then backfill reports.read for roles that already have analytics.read.
-- This fixes older databases where roles were created before /reports was introduced.

INSERT INTO menus (id, name, path)
SELECT
    '44444444-4444-4444-4444-444444444018',
    'รายงาน',
    '/reports'
WHERE NOT EXISTS (
    SELECT 1
    FROM menus
    WHERE id = '44444444-4444-4444-4444-444444444018'
       OR path = '/reports'
);

-- Keep admin role fully permitted on reports.
INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
SELECT
    '11111111-1111-1111-1111-111111111001',
    m.id,
    p.id
FROM permissions p
JOIN menus m ON m.path = '/reports'
ON CONFLICT DO NOTHING;

-- Backfill reports.read for any role that has analytics.read.
INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
SELECT DISTINCT
    rmp.role_id,
    m_reports.id,
    p_read.id
FROM role_menu_permissions rmp
JOIN menus m_analytics ON m_analytics.id = rmp.menu_id
JOIN permissions p_analytics ON p_analytics.id = rmp.permission_id
JOIN menus m_reports ON m_reports.path = '/reports'
JOIN permissions p_read ON p_read.name = 'read'
WHERE TRIM(BOTH '/' FROM m_analytics.path) = 'analytics'
  AND p_analytics.name = 'read'
ON CONFLICT DO NOTHING;
