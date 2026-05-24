INSERT INTO users (id, full_name, email, password, is_active)
VALUES
    (
        '33333333-3333-3333-3333-333333333001',
        'Administrator',
        'admin@horpug.local',
        'Admin@123456',
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