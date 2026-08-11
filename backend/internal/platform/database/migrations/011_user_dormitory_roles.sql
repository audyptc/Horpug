CREATE TABLE IF NOT EXISTS user_dormitory_roles (
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    dormitory_id UUID NOT NULL REFERENCES dormitories(id) ON DELETE CASCADE,
    role_id      UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, dormitory_id)
);

CREATE INDEX IF NOT EXISTS idx_user_dormitory_roles_dormitory_id
    ON user_dormitory_roles(dormitory_id);

CREATE INDEX IF NOT EXISTS idx_user_dormitory_roles_role_id
    ON user_dormitory_roles(role_id);

INSERT INTO user_dormitory_roles (user_id, dormitory_id, role_id)
SELECT ud.user_id, ud.dormitory_id, ur.role_id
FROM user_dormitories ud
JOIN user_roles ur ON ur.user_id = ud.user_id
ON CONFLICT (user_id, dormitory_id) DO NOTHING;