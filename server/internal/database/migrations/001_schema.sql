CREATE TABLE IF NOT EXISTS menus (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    path       TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS roles (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permissions (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY,
    full_name  TEXT NOT NULL,
    email      TEXT NOT NULL UNIQUE,
    password   TEXT NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id    UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);

CREATE TABLE IF NOT EXISTS role_menu_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    menu_id       UUID NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, menu_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_menu_permissions_role_id       ON role_menu_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_menu_permissions_menu_id       ON role_menu_permissions(menu_id);
CREATE INDEX IF NOT EXISTS idx_role_menu_permissions_permission_id ON role_menu_permissions(permission_id);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);

CREATE TABLE IF NOT EXISTS rooms (
    id          UUID PRIMARY KEY,
    room_number TEXT NOT NULL UNIQUE,
    floor       INTEGER NOT NULL,
    type        TEXT NOT NULL DEFAULT 'standard',
    status      TEXT NOT NULL DEFAULT 'available',
    rent_price  NUMERIC(10,2) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tenants (
    id                UUID PRIMARY KEY,
    first_name        TEXT NOT NULL,
    last_name         TEXT NOT NULL,
    phone             TEXT NOT NULL,
    id_card           TEXT NOT NULL UNIQUE,
    email             TEXT NOT NULL DEFAULT '',
    emergency_contact TEXT NOT NULL DEFAULT '',
    note              TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS contracts (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL REFERENCES tenants(id),
    room_id    UUID NOT NULL REFERENCES rooms(id),
    start_date DATE NOT NULL,
    end_date   DATE,
    rent_price NUMERIC(10,2) NOT NULL,
    deposit    NUMERIC(10,2) NOT NULL DEFAULT 0,
    status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'expired', 'terminated')),
    note       TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS meter_readings (
    id               UUID PRIMARY KEY,
    room_id          UUID NOT NULL REFERENCES rooms(id),
    meter_type       TEXT NOT NULL CHECK (meter_type IN ('electric', 'water')),
    reading_date     DATE NOT NULL,
    previous_reading NUMERIC(12,2) NOT NULL DEFAULT 0,
    current_reading  NUMERIC(12,2) NOT NULL,
    unit_price       NUMERIC(10,4) NOT NULL,
    note             TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bills (
    id              UUID PRIMARY KEY,
    contract_id     UUID NOT NULL REFERENCES contracts(id),
    billing_month   DATE NOT NULL,
    rent_amount     NUMERIC(10,2) NOT NULL DEFAULT 0,
    electric_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    water_amount    NUMERIC(10,2) NOT NULL DEFAULT 0,
    other_amount    NUMERIC(10,2) NOT NULL DEFAULT 0,
    other_note      TEXT NOT NULL DEFAULT '',
    total_amount    NUMERIC(10,2) NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'unpaid' CHECK (status IN ('unpaid', 'paid', 'overdue')),
    due_date        DATE,
    paid_at         TIMESTAMPTZ,
    note            TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS expenses (
    id           UUID PRIMARY KEY,
    expense_date DATE NOT NULL,
    category     TEXT NOT NULL DEFAULT 'other' CHECK (category IN ('repair', 'utilities', 'supplies', 'salary', 'other')),
    description  TEXT NOT NULL,
    amount       NUMERIC(10,2) NOT NULL DEFAULT 0,
    note         TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS maintenance_requests (
    id            UUID PRIMARY KEY,
    room_id       UUID NOT NULL REFERENCES rooms(id),
    title         TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'done', 'cancelled')),
    priority      TEXT NOT NULL DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    reported_date DATE NOT NULL,
    resolved_date DATE,
    note          TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payments (
    id           UUID PRIMARY KEY,
    bill_id      UUID NOT NULL REFERENCES bills(id) ON DELETE CASCADE,
    amount       NUMERIC(12,2) NOT NULL DEFAULT 0,
    method       TEXT NOT NULL DEFAULT 'cash' CHECK (method IN ('cash', 'transfer', 'qr')),
    payment_date DATE NOT NULL,
    note         TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS announcements (
    id           UUID PRIMARY KEY,
    title        TEXT NOT NULL DEFAULT '',
    content      TEXT NOT NULL DEFAULT '',
    type         TEXT NOT NULL DEFAULT 'general' CHECK (type IN ('general', 'maintenance', 'payment', 'emergency')),
    is_pinned    BOOLEAN NOT NULL DEFAULT FALSE,
    published_at DATE NOT NULL,
    expired_at   DATE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS parking_slots (
    id            UUID PRIMARY KEY,
    slot_number   TEXT NOT NULL DEFAULT '',
    vehicle_type  TEXT NOT NULL DEFAULT 'car' CHECK (vehicle_type IN ('car', 'motorcycle')),
    status        TEXT NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'occupied')),
    tenant_id     UUID REFERENCES tenants(id) ON DELETE SET NULL,
    license_plate TEXT NOT NULL DEFAULT '',
    monthly_fee   NUMERIC(10,2) NOT NULL DEFAULT 0,
    start_date    DATE,
    note          TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS parcels (
    id              UUID PRIMARY KEY,
    tracking_number TEXT NOT NULL DEFAULT '',
    recipient_name  TEXT NOT NULL DEFAULT '',
    room_number     TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'picked_up')),
    received_date   DATE NOT NULL DEFAULT CURRENT_DATE,
    picked_up_date  DATE,
    note            TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS documents (
    id          UUID PRIMARY KEY,
    title       TEXT NOT NULL DEFAULT '',
    category    TEXT NOT NULL DEFAULT 'other' CHECK (category IN ('contract', 'id_card', 'house_registration', 'receipt', 'other')),
    tenant_id   UUID REFERENCES tenants(id) ON DELETE SET NULL,
    file_url    TEXT NOT NULL DEFAULT '',
    issue_date  DATE,
    expiry_date DATE,
    note        TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
