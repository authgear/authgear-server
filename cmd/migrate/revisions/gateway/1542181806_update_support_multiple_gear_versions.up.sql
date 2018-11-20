BEGIN;

ALTER TABLE app ADD auth_version text NOT NULL;
ALTER TABLE app ADD record_version text NOT NULL;

END;
