-- Add flat rate billing support for electric meters
ALTER TABLE electric_meter_readings
  ADD COLUMN IF NOT EXISTS billing_type TEXT NOT NULL DEFAULT 'meter' CHECK (billing_type IN ('meter', 'flat')),
  ADD COLUMN IF NOT EXISTS flat_amount   NUMERIC(10,2);

ALTER TABLE electric_meter_readings
  ALTER COLUMN current_reading DROP NOT NULL,
  ALTER COLUMN unit_price      DROP NOT NULL;
