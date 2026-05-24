INSERT INTO users (id, full_name, email, password, is_active)
VALUES
    (
        '33333333-3333-3333-3333-333333333001',
        'Administrator',
        'admin@horpug.local',
        '$2a$10$np/a/xuyJ/ozav.f4TU0xueDuUgCRtejuKTCjvjb/u9C31k.loz8u',
        TRUE
    )
ON CONFLICT (email) DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
VALUES
    (
        '33333333-3333-3333-3333-333333333001',
        '11111111-1111-1111-1111-111111111001'
    )
ON CONFLICT (user_id, role_id) DO NOTHING;