-- Add audit columns to rooms table
ALTER TABLE rooms ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id);
ALTER TABLE rooms ADD COLUMN IF NOT EXISTS updated_by UUID REFERENCES users(id);
