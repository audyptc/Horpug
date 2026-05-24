CREATE TABLE IF NOT EXISTS role_menu_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    menu_id UUID NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, menu_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_menu_permissions_role_id ON role_menu_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_menu_permissions_menu_id ON role_menu_permissions(menu_id);
CREATE INDEX IF NOT EXISTS idx_role_menu_permissions_permission_id ON role_menu_permissions(permission_id);