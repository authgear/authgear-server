ALTER TABLE app ADD is_free boolean NOT NULL DEFAULT FALSE;

UPDATE app SET "is_free" = true;
