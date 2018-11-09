BEGIN;

SET search_path TO app_config;

ALTER TABLE plan DROP record_enabled;

SET search_path TO DEFAULT;

END;
