ALTER TABLE secret ADD COLUMN "type" text;

UPDATE secret SET "type" = 'deprecated';

ALTER TABLE secret ALTER COLUMN "type" SET NOT NULL;
