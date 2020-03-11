-- Put upgrade SQL here
ALTER TABLE artifact ADD COLUMN "checksum_sha256" text;
ALTER TABLE artifact ADD COLUMN "checksum_md5" text;

UPDATE artifact SET "checksum_sha256" = "checksum";

-- leave it blank
UPDATE artifact SET "checksum_md5" = '';

ALTER TABLE artifact ALTER COLUMN "checksum_sha256" SET NOT NULL;
ALTER TABLE artifact ALTER COLUMN "checksum_md5" SET NOT NULL;

ALTER TABLE artifact DROP COLUMN "checksum";
