-- Put upgrade SQL here
ALTER TABLE cloud_code ADD COLUMN "entry_point" text;

UPDATE cloud_code
SET entry_point = '';

ALTER TABLE cloud_code ALTER COLUMN "entry_point" SET NOT NULL;
