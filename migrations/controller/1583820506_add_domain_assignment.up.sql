ALTER TABLE domain ADD COLUMN "assignment" TEXT;
UPDATE domain SET "assignment" = 'microservices';
ALTER TABLE domain ALTER COLUMN "assignment" SET NOT NULL;

ALTER TABLE custom_domain ADD COLUMN "assignment" TEXT;
UPDATE custom_domain SET "assignment" = 'microservices';
ALTER TABLE custom_domain ALTER COLUMN "assignment" SET NOT NULL;
