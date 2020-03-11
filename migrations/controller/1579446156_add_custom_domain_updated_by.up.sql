ALTER TABLE custom_domain ADD COLUMN "updated_by" text REFERENCES _core_user(id);
UPDATE custom_domain SET updated_by = created_by;
ALTER TABLE app ALTER COLUMN "updated_by" SET NOT NULL;

ALTER TABLE custom_domain ADD COLUMN "updated_at" timestamp WITHOUT TIME ZONE;
UPDATE custom_domain SET updated_at = created_at;
ALTER TABLE app ALTER COLUMN "updated_at" SET NOT NULL;
