ALTER TABLE activity_logs
    ADD COLUMN IF NOT EXISTS dormitory_id UUID REFERENCES dormitories(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_activity_logs_dormitory_id
    ON activity_logs(dormitory_id, created_at DESC);
