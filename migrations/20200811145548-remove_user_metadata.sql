-- +migrate Up
ALTER TABLE _auth_user DROP COLUMN metadata;

-- +migrate Down
ALTER TABLE _auth_user ADD COLUMN metadata JSONB;
