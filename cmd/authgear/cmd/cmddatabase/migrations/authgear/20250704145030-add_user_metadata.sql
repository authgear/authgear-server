-- +migrate Up

ALTER TABLE _auth_user ADD COLUMN metadata jsonb NOT NULL DEFAULT '{}'::jsonb;

-- +migrate Down

ALTER TABLE _auth_user DROP COLUMN metadata;
