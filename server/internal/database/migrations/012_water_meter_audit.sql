-- Add audit columns to water_meter_readings table
ALTER TABLE water_meter_readings ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id);
ALTER TABLE water_meter_readings ADD COLUMN IF NOT EXISTS updated_by UUID REFERENCES users(id);
