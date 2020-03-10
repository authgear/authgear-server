ALTER TABLE "microservice" ADD COLUMN "raw_config" jsonb;
UPDATE "microservice" SET "raw_config" = '{}';
ALTER TABLE "microservice" ALTER COLUMN "raw_config" SET NOT NULL;
