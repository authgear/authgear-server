BEGIN;

ALTER TABLE _core_user ADD COLUMN "last_login_at" timestamp without time zone;

END;
