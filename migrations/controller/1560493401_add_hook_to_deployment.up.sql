-- Put upgrade SQL here
ALTER TABLE "deployment" ADD COLUMN "hook" JSONB NOT NULL DEFAULT '[]'::JSONB;
