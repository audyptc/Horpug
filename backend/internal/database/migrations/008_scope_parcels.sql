ALTER TABLE parcels ADD COLUMN dormitory_id UUID;
UPDATE parcels SET dormitory_id = (SELECT id FROM dormitories ORDER BY created_at LIMIT 1);
ALTER TABLE parcels ALTER COLUMN dormitory_id SET NOT NULL;
ALTER TABLE parcels ADD CONSTRAINT fk_parcels_dormitory FOREIGN KEY (dormitory_id) REFERENCES dormitories(id);
CREATE INDEX IF NOT EXISTS idx_parcels_dormitory_id ON parcels(dormitory_id);
