ALTER TABLE app ADD record_version text NOT NULL;
ALTER TABLE plan ADD record_enabled boolean NOT NULL DEFAULT FALSE;