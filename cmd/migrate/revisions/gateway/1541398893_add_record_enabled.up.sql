BEGIN;

SET search_path TO app_config;

ALTER TABLE plan ADD record_enabled boolean NOT NULL DEFAULT FALSE;

SET search_path TO DEFAULT;

END;
