-- Put downgrade SQL here
ALTER TABLE cloud_code ADD COLUMN "config" jsonb;

UPDATE cloud_code SET "config" = '{}'::jsonb;

ALTER TABLE cloud_code ALTER COLUMN "config" SET NOT NULL;
