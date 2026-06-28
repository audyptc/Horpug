ALTER TABLE electric_meter_readings ADD COLUMN IF NOT EXISTS billing_month DATE;
ALTER TABLE water_meter_readings ADD COLUMN IF NOT EXISTS billing_month DATE;
