ALTER TABLE artifact DROP COLUMN "storage_type";
ALTER TABLE artifact DROP COLUMN "storage_data";

ALTER TABLE artifact ADD COLUMN "asset_name" text;
UPDATE artifact SET "asset_name" = '';
ALTER TABLE artifact ALTER COLUMN "asset_name" SET NOT NULL;
