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
