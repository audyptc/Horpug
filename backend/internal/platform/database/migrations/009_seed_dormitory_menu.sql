INSERT INTO menus (id, name, path)
SELECT gen_random_uuid(), 'จัดการหอพัก', '/settings/dormitories'
WHERE NOT EXISTS (SELECT 1 FROM menus WHERE path = '/settings/dormitories');
