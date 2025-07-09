-- +migrate Up

ALTER TABLE _auth_user ADD COLUMN metadata jsonb NULL;

-- +migrate Down

ALTER TABLE _auth_user DROP COLUMN metadata;
