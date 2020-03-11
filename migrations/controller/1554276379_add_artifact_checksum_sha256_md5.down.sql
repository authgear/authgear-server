-- Put downgrade SQL here
ALTER TABLE artifact ADD COLUMN "checksum" text;

UPDATE artifact
SET "checksum" = "checksum_sha256";

ALTER TABLE artifact ALTER COLUMN "checksum" SET NOT NULL;

ALTER TABLE artifact DROP COLUMN "checksum_sha256";
ALTER TABLE artifact DROP COLUMN "checksum_md5";
