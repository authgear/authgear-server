ALTER TABLE "microservice" ADD COLUMN "command" jsonb;
UPDATE "microservice" SET "command" = '[]';
ALTER TABLE "microservice" ALTER COLUMN "command" SET NOT NULL;
