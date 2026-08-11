ALTER TABLE parking_slots ADD COLUMN dormitory_id UUID;
UPDATE parking_slots SET dormitory_id = (SELECT id FROM dormitories ORDER BY created_at LIMIT 1);
ALTER TABLE parking_slots ALTER COLUMN dormitory_id SET NOT NULL;
ALTER TABLE parking_slots ADD CONSTRAINT fk_parking_slots_dormitory FOREIGN KEY (dormitory_id) REFERENCES dormitories(id);
CREATE INDEX IF NOT EXISTS idx_parking_slots_dormitory_id ON parking_slots(dormitory_id);

ALTER TABLE expenses ADD COLUMN dormitory_id UUID;
UPDATE expenses SET dormitory_id = (SELECT id FROM dormitories ORDER BY created_at LIMIT 1);
ALTER TABLE expenses ALTER COLUMN dormitory_id SET NOT NULL;
ALTER TABLE expenses ADD CONSTRAINT fk_expenses_dormitory FOREIGN KEY (dormitory_id) REFERENCES dormitories(id);
CREATE INDEX IF NOT EXISTS idx_expenses_dormitory_id ON expenses(dormitory_id);

ALTER TABLE announcements ADD COLUMN dormitory_id UUID;
UPDATE announcements SET dormitory_id = (SELECT id FROM dormitories ORDER BY created_at LIMIT 1);
ALTER TABLE announcements ALTER COLUMN dormitory_id SET NOT NULL;
ALTER TABLE announcements ADD CONSTRAINT fk_announcements_dormitory FOREIGN KEY (dormitory_id) REFERENCES dormitories(id);
CREATE INDEX IF NOT EXISTS idx_announcements_dormitory_id ON announcements(dormitory_id);

ALTER TABLE documents ADD COLUMN dormitory_id UUID;
UPDATE documents SET dormitory_id = (SELECT id FROM dormitories ORDER BY created_at LIMIT 1);
ALTER TABLE documents ALTER COLUMN dormitory_id SET NOT NULL;
ALTER TABLE documents ADD CONSTRAINT fk_documents_dormitory FOREIGN KEY (dormitory_id) REFERENCES dormitories(id);
CREATE INDEX IF NOT EXISTS idx_documents_dormitory_id ON documents(dormitory_id);
