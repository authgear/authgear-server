-- Put upgrade SQL here
ALTER TABLE cloud_code ADD COLUMN "hook" jsonb;
