ALTER TABLE artifact DROP COLUMN "asset_name";

ALTER TABLE artifact ADD COLUMN "storage_type" text;
UPDATE artifact SET "storage_type" = '';
ALTER TABLE artifact ALTER COLUMN "storage_type" SET NOT NULL;

ALTER TABLE artifact ADD COLUMN "storage_data" jsonb;
UPDATE artifact SET "storage_data" = '{}';
ALTER TABLE artifact ALTER COLUMN "storage_data" SET NOT NULL;
