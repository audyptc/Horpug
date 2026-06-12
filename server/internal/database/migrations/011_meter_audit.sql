-- Add audit columns to electric_meter_readings table
ALTER TABLE electric_meter_readings ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id);
ALTER TABLE electric_meter_readings ADD COLUMN IF NOT EXISTS updated_by UUID REFERENCES users(id);
