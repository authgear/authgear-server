ALTER TABLE config ALTER COLUMN app_id TYPE uuid USING app_id::uuid;
